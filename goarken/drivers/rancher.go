// Copyright © 2016 Nuxeo SA (http://nuxeo.com/) and others.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package drivers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	. "github.com/arkenio/arken/goarken/model"
	catalogclient "github.com/dmetzler/go-ranchercatalog/client"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/rancher/go-rancher/client"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

var (
	rancherHostRegexp = regexp.MustCompile("http.*/projects/(.*)")
	log               = logrus.New()
)

type RancherServiceDriver struct {
	rancherClient    *client.RancherClient
	broadcaster      *Broadcaster
	rancherCatClient *catalogclient.RancherCatalogClient
}

func NewRancherServiceDriver(rancherHost string, rancherAccessKey string, rancherSecretKey string) (*RancherServiceDriver, error) {

	clientOpts := &client.ClientOpts{
		Url:       rancherHost,
		AccessKey: rancherAccessKey,
		SecretKey: rancherSecretKey,
	}

	catalaogClientOpts := &catalogclient.ClientOpts{
		Url:       clientOpts.Url,
		AccessKey: clientOpts.AccessKey,
		SecretKey: clientOpts.SecretKey,
	}

	rancherClient, err := client.NewRancherClient(clientOpts)
	rancherCatClient, err := catalogclient.NewRancherCatalogClient(catalaogClientOpts)

	if err != nil {
		return nil, err
	}

	sd := &RancherServiceDriver{
		rancherClient,
		NewBroadcaster(),
		rancherCatClient,
	}

	c, _, err := getRancherSocket(rancherClient)
	if err != nil {
		return nil, err
	}
	go sd.watch(c)

	return sd, nil

}

func getProjectIdFromRancherHost(host string) string {
	matches := rancherHostRegexp.FindStringSubmatch(host)
	if len(matches) > 1 {
		return matches[1]
	} else {
		return ""
	}
}

func getRancherSocket(r *client.RancherClient) (*websocket.Conn, *http.Response, error) {
	rancherUrl, _ := url.Parse(r.Opts.Url)

	projectId := getProjectIdFromRancherHost(r.Opts.Url)
	u := url.URL{
		Scheme:   "ws",
		Host:     rancherUrl.Host,
		Path:     "/v1/subscribe",
		RawQuery: fmt.Sprintf("eventNames=resource.change&include=hosts&include=instances&include=instance&include=instanceLinks&include=ipAddresses&projectId=%s", projectId),
	}

	header := http.Header{
		"Authorization": []string{"Basic " + basicAuth(r.Opts.AccessKey, r.Opts.SecretKey)},
	}

	return r.Websocket(u.String(), header)
}

func (r *RancherServiceDriver) watch(c *websocket.Conn) {

	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			//TODO find a way to recover or to gracefully exit
			return
		}

		publish := &client.Publish{}
		json.Unmarshal([]byte(message), publish)
		if publish.Name == "resource.change" {

			switch publish.ResourceType {
			case "environment":
				var result client.Environment
				err := mapstructure.Decode(publish.Data["resource"], &result)
				if err != nil {
					log.Printf(err.Error())
				} else {
					info := rancherInfoTypeFromEnvironment(&result)
					info.EnvironmentId = publish.ResourceId
					r.broadcaster.Write(NewModelEvent("update", info))
				}
				break
			}
		}
	}

}

func rancherInfoTypeFromEnvironment(e *client.Environment) *RancherInfoType {
	return &RancherInfoType{
		EnvironmentId:   e.Id,
		EnvironmentName: e.Name,
		Location:        &Location{Host: fmt.Sprintf("lb.%s", e.Name), Port: 80},
		HealthState:     e.HealthState,
		CurrentStatus:   convertRancherHealthToStatus(e.HealthState),
		TemplateId:      strings.Replace(e.ExternalId, "catalog://", "", 1),
	}
}

func convertRancherHealthToStatus(health string) string {
	switch health {
	case "healthy":
		return STARTED_STATUS
	case "degraded", "activating", "initializing":
		return STARTING_STATUS
	default:
		return STOPPED_STATUS
	}
	return STOPPED_STATUS
}

func (r *RancherServiceDriver) computeEnvFromService(s *Service) (*client.Environment, error) {
	info := s.Config.RancherInfo

	if info == nil || info.TemplateId == "" {
		return nil, errors.New("Rancher template has to be specified !")
	}
	log.Infof("Looking for template %s", info.TemplateId)
	template, err := r.rancherCatClient.Template.ById(info.TemplateId)
	if err != nil {
		log.Error("Rancher template not found : " + err.Error())
		return nil, errors.New("Rancher template not found : " + err.Error())
	}

	//Start rancher environment
	env := &client.Environment{}
	env.Name = s.Name
	env.Environment = s.Config.Environment
	env.DockerCompose = extractFileContent(template, "docker-compose.yml")
	env.RancherCompose = extractFileContent(template, "rancher-compose.yml")
	env.ExternalId = "catalog://" + info.TemplateId
	return env, nil
}

func (r *RancherServiceDriver) Create(s *Service, startOnCreate bool) (interface{}, error) {

	log.Infof("Creating stack %s on rancher", s.Name)
	env, err := r.computeEnvFromService(s)

	if err != nil {
		return nil, err
	}

	env.StartOnCreate = startOnCreate

	env, err = r.rancherClient.Environment.Create(env)

	if err != nil {
		log.Error("Error when creating service on Rancher side: " + err.Error())
		return nil, errors.New("Error when creating service on Rancher side: " + err.Error())
	}

	return &RancherInfoType{EnvironmentId: env.Id}, nil

}

func (r *RancherServiceDriver) NeedToBeUpgraded(s *Service) (bool, error) {

	newEnv, err := r.computeEnvFromService(s)

	if err != nil {
		log.Error("Error computing the environment from service %v", err)
		return false, err
	}

	rancherId := s.Config.RancherInfo.EnvironmentId
	actualEnv, err := r.rancherClient.Environment.ById(rancherId)
	if err != nil {
		
		log.Error("Error when fetching the environment %v", err)
		return false, err
	}

	return !(reflect.DeepEqual(newEnv.Environment, actualEnv.Environment) &&
		newEnv.DockerCompose == actualEnv.DockerCompose &&
		newEnv.RancherCompose == actualEnv.RancherCompose), nil

}

func (r *RancherServiceDriver) Upgrade(s *Service) (interface{}, error) {

	log.Infof("Upgrading environment %s", s.Name)

	rancherId := s.Config.RancherInfo.EnvironmentId
	env, err := r.rancherClient.Environment.ById(rancherId)

	if err != nil {
		log.Errorf("Error when retrieving environment: %v", err)
	}

	info := s.Config.RancherInfo

	if info == nil || info.TemplateId == "" {
		log.Errorf("Rancher template has to be specified : %v", info)
		return nil, errors.New("Rancher template has to be specified !")
	}

	log.Infof("    Template: %s", info.TemplateId)
	log.Infof("	   Environment: %v", s.Config.Environment)

	template, err := r.rancherCatClient.Template.ById(info.TemplateId)
	if err != nil {
		log.Errorf("Template not found : %v", info.TemplateId)
		return nil, err
	}

	envUpgrade := &client.EnvironmentUpgrade{
		DockerCompose:  extractFileContent(template, "docker-compose.yml"),
		RancherCompose: extractFileContent(template, "rancher-compose.yml"),
		Environment:    s.Config.Environment,
		ExternalId :    "catalog://" + info.TemplateId,
	}

	env, err = r.rancherClient.Environment.ActionUpgrade(env, envUpgrade)

	if err != nil {
		log.Errorf("Environment upgrade failed in rancher : %v", err)
	}

	return s, err
}

func (r *RancherServiceDriver) FinishUpgrade(s *Service) (interface{}, error) {

	log.Infof("Finishing upgrading environment %s", s.Name)

	rancherId := s.Config.RancherInfo.EnvironmentId
	env, err := r.rancherClient.Environment.ById(rancherId)

	if err != nil {
		log.Errorf("Error when retrieving environment: %v", err)
	}

	env, err = r.rancherClient.Environment.ActionFinishupgrade(env)

	if err != nil {
		log.Errorf("Finish upgrade environment failed in rancher : %v", err)
	}

	return s, err
}

func (r *RancherServiceDriver) Rollback(s *Service) (interface{}, error) {

	log.Infof("Rollbacking environment %s", s.Name)

	rancherId := s.Config.RancherInfo.EnvironmentId
	env, err := r.rancherClient.Environment.ById(rancherId)

	if err != nil {
		log.Errorf("Error when retrieving environment: %v", err)
	}

	env, err = r.rancherClient.Environment.ActionRollback(env)

	if err != nil {
		log.Errorf("Rollbacking environment failed in rancher : %v", err)
	}

	return s, err
}

func extractFileContent(template *catalogclient.Template, filename string) string {
	if content, ok := template.Files[filename].(string); ok {
		return content
	}
	return ""
}

func (r *RancherServiceDriver) Start(s *Service) (interface{}, error) {
	rancherId := s.Config.RancherInfo.EnvironmentId
	env, err := r.rancherClient.Environment.ById(rancherId)
	if err != nil {
		return nil, err
	}
	env, err = r.rancherClient.Environment.ActionActivateservices(env)
	return s, err
}

func (r *RancherServiceDriver) Stop(s *Service) (interface{}, error) {
	rancherId := s.Config.RancherInfo.EnvironmentId
	env, err := r.rancherClient.Environment.ById(rancherId)
	if err != nil {
		return nil, err
	}
	env, err = r.rancherClient.Environment.ActionDeactivateservices(env)
	return s, err
}

func (r *RancherServiceDriver) Destroy(s *Service) error {
	rancherId := s.Config.RancherInfo.EnvironmentId
	env, err := r.rancherClient.Environment.ById(rancherId)
	if err != nil {
		return err
	}
	return r.rancherClient.Environment.Delete(env)
}

func (r *RancherServiceDriver) Listen() chan *ModelEvent {
	return FromInterfaceChannel(r.broadcaster.Listen())
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Return the RancheInfotype for the given service
func (r *RancherServiceDriver) GetInfo(s *Service) (interface{}, error) {
	rancherId := s.Config.RancherInfo.EnvironmentId

	env, error := r.rancherClient.Environment.ById(rancherId)
	if error != nil {
		return nil, error
	} else {
		return rancherInfoTypeFromEnvironment(env), nil
	}

}
