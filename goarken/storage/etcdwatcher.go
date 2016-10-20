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
package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/arkenio/arken/goarken/model"
	etcd "github.com/coreos/etcd/client"
	"regexp"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
)

var (
	domainRegexp  = regexp.MustCompile("/domain/(.*)(/.*)*")
	serviceRegexp = regexp.MustCompile("/services/(.*)(/.*)*")
	log           = logrus.New()
)

// Watcher implements the PersistenceDriver interface of the Arken
// Model and allows to store the model in etcd.
type Watcher struct {
	kapi          etcd.KeysAPI
	broadcaster   *Broadcaster
	servicePrefix string
	domainPrefix  string
	stop          *Broadcaster
}

func NewWatcher(client etcd.KeysAPI, servicePrefix string, domainPrefix string) *Watcher {

	watcher := &Watcher{
		kapi:          client,
		broadcaster:   NewBroadcaster(),
		servicePrefix: servicePrefix,
		domainPrefix:  domainPrefix,
		stop:          NewBroadcaster(),
	}

	watcher.Init()

	return watcher
}

//Init Domains and Services.
func (w *Watcher) Init() {

	setDomainPrefix(w.domainPrefix)
	SetServicePrefix(w.servicePrefix)

	//Create prefix dir if they do not exist
	createDirIfNotExist(w.servicePrefix, w.kapi)
	createDirIfNotExist(w.domainPrefix, w.kapi)

	if w.domainPrefix != "" {
		go w.doWatch(w.domainPrefix, w.registerDomain)
	}
	if w.servicePrefix != "" {
		go w.doWatch(w.servicePrefix, w.registerService)
	}

}

func createDirIfNotExist(dir string, kapi etcd.KeysAPI) {

	_, err := kapi.Get(context.Background(), dir, nil)
	if err != nil {
		kapi.Set(context.Background(), dir, "", &etcd.SetOptions{PrevExist: etcd.PrevNoExist, Dir: true})
	}
}

func (w *Watcher) Listen() chan *ModelEvent {

	return FromInterfaceChannel(w.broadcaster.Listen())

}

// Loads and watch an etcd directory to register objects like Domains, Services
// etc... The register function is passed the etcd Node that has been loaded.
func (w *Watcher) doWatch(etcdDir string, registerFunc func(*etcd.Node, string)) {

	for {
		watcher := w.kapi.Watcher(etcdDir, &etcd.WatcherOptions{Recursive: true})

		for {
			resp, err := watcher.Next(context.Background())
			if err == nil {
				registerFunc(resp.Node, resp.Action)
			} else {
				break
			}
		}

		log.Warningf("Waiting 1 second and relaunch watch")
		time.Sleep(time.Second)

	}

}

func (w *Watcher) loadPrefix(etcDir string, registerFunc func(*etcd.Node, string)) {
	response, err := w.kapi.Get(context.Background(), etcDir, &etcd.GetOptions{Recursive: true})
	if err == nil {
		for _, serviceNode := range response.Node.Nodes {
			registerFunc(serviceNode, response.Action)

		}
	}
}

//func GetDomainFromPath(domainPath string, client *etcd.Client) (*Domain, error) {
//	// Get service's root node instead of changed node.
//	response, err := client.Get(domainPath, true, true)
//	if err != nil {
//		return nil, errors.New(fmt.Sprintf("Unable to get information for service %s from etcd", domainPath))
//	}
//
//	return GetDomainFromNode(response.Node), nil
//}

func GetDomainFromNode(node *etcd.Node) (*Domain, error) {
	return NewDomain(node)
}

func GetServiceFromPath(servicePath string, kapi etcd.KeysAPI) (*Service, error) {
	// Get service's root node instead of changed node.
	response, err := kapi.Get(context.Background(), servicePath, &etcd.GetOptions{Recursive: true})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to get information for service %s from etcd", servicePath))
	}

	return getServiceFromNode(response.Node), nil
}

func getServiceFromNode(serviceNode *etcd.Node) *Service {

	service, err := newService(serviceNode)
	if err == nil {
		log.Infof("Can not get service from node %v", serviceNode)
	}
	return service
}

func (w *Watcher) registerDomain(node *etcd.Node, action string) {

	domainName, err := getDomainForNode(node)
	if err != nil {
		return
	}

	if action == "delete" || action == "expire" {
		w.broadcaster.Write(NewModelEvent("delete", &Domain{Name: domainName}))
		return
	}

	domainKey := w.domainPrefix + "/" + domainName
	response, err := w.kapi.Get(context.Background(), domainKey, &etcd.GetOptions{Recursive: true})

	if err == nil {
		domain, _ := NewDomain(response.Node)
		if domain.Typ != "" && domain.Value != "" { // && !domain.Equals(actualDomain) {
			w.broadcaster.Write(NewModelEvent("update", domain))
		}

	}

}

func NewDomain(domainNode *etcd.Node) (*Domain, error) {
	domain := &Domain{}

	domain.NodeKey = domainNode.Key

	domainName, err := getDomainForNode(domainNode)
	if err != nil {
		return nil, err
	}
	domain.Name = domainName

	for _, node := range domainNode.Nodes {
		switch node.Key {
		case domainNode.Key + "/type":
			domain.Typ = node.Value
		case domainNode.Key + "/value":
			domain.Value = node.Value
		}
	}
	return domain, nil

}

func (w *Watcher) registerService(node *etcd.Node, action string) {

	serviceName, err := getEnvForNode(node)
	if err != nil || serviceName == "" {
		return
	}

	if action == "delete" && node.Key == w.servicePrefix+"/"+serviceName {
		w.broadcaster.Write(NewModelEvent("delete", &Service{Name: serviceName}))
		return
	}

	// Get service's root node instead of changed node.
	response, err := w.kapi.Get(context.Background(), w.servicePrefix+"/"+serviceName, &etcd.GetOptions{Recursive: true, Sort: true})

	if err == nil {
		sc := getServiceFromNode(response.Node)
		w.broadcaster.Write(NewModelEvent("update", sc))
	} else {
		log.Errorf("Unable to get information for service %s from etcd (%v) update on %s", serviceName, err, node.Key)
	}
}

func newService(serviceNode *etcd.Node) (*Service, error) {

	serviceName, err := getEnvForNode(serviceNode)
	if err != nil {
		return nil, err
	}

	service := &Service{}
	service.Location = &Location{}
	service.Config = &ServiceConfig{Robots: ""}
	service.Name = serviceName
	service.NodeKey = serviceNode.Key

	for _, node := range serviceNode.Nodes {
		switch node.Key {
		case service.NodeKey + "/location":
			location := &Location{}
			err := json.Unmarshal([]byte(node.Value), location)
			if err == nil {
				service.Location.Host = location.Host
				service.Location.Port = location.Port
			}

		case service.NodeKey + "/config":
			serviceConfig := &ServiceConfig{}
			err := json.Unmarshal([]byte(node.Value), serviceConfig)
			if err == nil {
				service.Config = serviceConfig
			}
		case service.NodeKey + "/domain":
			service.Domain = node.Value
		case service.NodeKey + "/lastAccess":
			lastAccess := node.Value
			lastAccessTime, err := time.Parse(TIME_FORMAT, lastAccess)
			if err != nil {
				log.Errorf("Error parsing last access date with service %s: %s", service.Name, err)
				break
			}
			service.LastAccess = &lastAccessTime

		case service.NodeKey + "/status":
			service.Status = NewStatus(service, node)
		case service.NodeKey + "/actions":
			var actions []string
			err := json.Unmarshal([]byte(node.Value), &actions)
			if err != nil {
				log.Errorf("Error parsing actions on the service %s: %v", service.Name, err)
				break
			}
			service.Actions = actions
		}
	}
	return service, nil
}

func (w *Watcher) PersistService(s *Service) (*Service, error) {
	if s.NodeKey != "" {
		log.Debugf("Persisting key %s ", s.NodeKey)
		resp, err := w.kapi.Get(context.Background(), s.NodeKey, &etcd.GetOptions{Recursive: true, Sort: false})
		if err != nil {
			return nil, errors.New("No service with key " + s.NodeKey + " found in etcd")
		}

		oldService, err := newService(resp.Node)
		if err != nil {
			return nil, err
		} else {
			if oldService.Status.Expected != s.Status.Expected {
				_, err = w.kapi.Set(context.Background(), fmt.Sprintf("%s/status/expected", s.NodeKey), s.Status.Expected, nil)
			}

			if err == nil && oldService.Status.Current != s.Status.Current {
				_, err = w.kapi.Set(context.Background(), fmt.Sprintf("%s/status/current", s.NodeKey), s.Status.Current, nil)
			}

			if err == nil && oldService.Status.Alive != s.Status.Alive {
				_, err = w.kapi.Set(context.Background(), fmt.Sprintf("%s/status/alive", s.NodeKey), s.Status.Alive, nil)
			}

			bytes, err2 := json.Marshal(s.Location)
			if err2 == nil && oldService.Location != s.Location {
				_, err = w.kapi.Set(context.Background(), fmt.Sprintf("%s/location", s.NodeKey), string(bytes), nil)
			} else {
				err = err2
			}
			log.Debugf("Persisting actions %v on service %s", s.Actions, s.Name)
			if len(s.Actions.([]string)) > 0 { //don't perists actions on intermediate states
				bytes, err2 = json.Marshal(s.Actions.([]string))
				if err2 == nil {
					_, err = w.kapi.Set(context.Background(), fmt.Sprintf("%s/actions", s.NodeKey), string(bytes), nil)
				} else {
					err = err2
				}
			}

			bytes, err2 = json.Marshal(s.Config)
			if err2 == nil {
				_, err = w.kapi.Set(context.Background(), fmt.Sprintf("%s/config", s.NodeKey), string(bytes), nil)
			} else {
				err = err2
			}

			if err == nil && oldService.Domain != s.Domain {
				_, err = w.kapi.Set(context.Background(), fmt.Sprintf("%s/domain", s.NodeKey), s.Domain, nil)
			}

			if err != nil {
				return nil, err
			}

		}

	} else {
		s.NodeKey = computeNodeKey(s, w.servicePrefix)

		_, err := w.kapi.Create(context.Background(), fmt.Sprintf("%s/status/expected", s.NodeKey), s.Status.Expected)
		if err == nil {
			_, err = w.kapi.Create(context.Background(), fmt.Sprintf("%s/status/current", s.NodeKey), s.Status.Current)
		}
		if err == nil {
			bytes, err := json.Marshal(s.Config)
			if err == nil {
				_, err = w.kapi.Create(context.Background(), fmt.Sprintf("%s/config", s.NodeKey), string(bytes))
			}
		}
		if err == nil {
			bytes, err := json.Marshal(s.Actions.([]string))
			if err == nil {
				_, err = w.kapi.Create(context.Background(), fmt.Sprintf("%s/actions", s.NodeKey), string(bytes))
			}
		}
		if err == nil {
			_, err = w.kapi.Create(context.Background(), fmt.Sprintf("%s/domain", s.NodeKey), s.Domain)
		}

		if err != nil {
			//Rollback creation
			log.Warnf("Rollback creation of service %s", s.Name)
			w.kapi.Delete(context.Background(), s.NodeKey, &etcd.DeleteOptions{Recursive: true})
			return nil, err
		}

	}
	return s, nil

}

func computeNodeKey(s *Service, servicePrefix string) string {
	return fmt.Sprintf("/%s/%s", servicePrefix, s.Name)
}

func computeDomainNodeKey(domainName string, domainPrefix string) string {
	return fmt.Sprintf("/%s/%s", domainPrefix, domainName)
}

func computeServiceKey(serviceName string, servicePrefix string) string {
	return fmt.Sprintf("/%s/%s/", servicePrefix, serviceName)
}

func getEnvForNode(node *etcd.Node) (string, error) {
	matches := serviceRegexp.FindStringSubmatch(node.Key)
	if len(matches) > 1 {
		parts := strings.Split(matches[1], "/")
		return parts[0], nil
	} else {
		return "", errors.New("Unable to extract env for node " + node.Key)
	}
}

func setDomainPrefix(domainPrefix string) {
	domainRegexp = regexp.MustCompile(domainPrefix + "/(.*)(/.*)*")
}

func getDomainForNode(node *etcd.Node) (string, error) {
	matches := domainRegexp.FindStringSubmatch(node.Key)
	if len(matches) > 1 {
		parts := strings.Split(matches[1], "/")
		return parts[0], nil
	} else {
		return "", errors.New("Unable to extract domain for node " + node.Key)
	}
}

func SetServicePrefix(servicePrefix string) {
	serviceRegexp = regexp.MustCompile(servicePrefix + "/(.*)(/.*)*")
}

func NewStatus(service *Service, node *etcd.Node) *Status {
	status := &Status{}
	statusKey := service.NodeKey + "/status"
	status.Service = service
	for _, subNode := range node.Nodes {
		switch subNode.Key {
		case statusKey + "/alive":
			status.Alive = subNode.Value
		case statusKey + "/current":
			status.Current = subNode.Value
		case statusKey + "/expected":
			status.Expected = subNode.Value
		}
	}
	return status
}

func (w *Watcher) LoadAllServices() (map[string]*Service, error) {
	result := make(map[string]*Service)

	response, err := w.kapi.Get(context.Background(), w.servicePrefix, &etcd.GetOptions{Recursive: true, Sort: true})
	if err == nil {
		for _, serviceNode := range response.Node.Nodes {
			sc := getServiceFromNode(serviceNode)
			result[sc.Name] = sc
		}
	} else {
		return nil, err
	}
	return result, nil
}

func (w *Watcher) LoadService(serviceName string) (*Service, error) {

	response, err := w.kapi.Get(context.Background(), computeServiceKey(serviceName, w.servicePrefix), &etcd.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return nil, err
	} else {
		return getServiceFromNode(response.Node), nil
	}

}

func (w *Watcher) DestroyService(sc *Service) error {
	_, err := w.kapi.Delete(context.Background(), computeServiceKey(sc.Name, w.servicePrefix), &etcd.DeleteOptions{Recursive: true})

	return err
}

func (w *Watcher) LoadAllDomains() (map[string]*Domain, error) {

	result := make(map[string]*Domain)

	response, err := w.kapi.Get(context.Background(), w.domainPrefix, &etcd.GetOptions{Recursive: true, Sort: true})
	if err == nil {
		for _, domainNode := range response.Node.Nodes {
			domain, err := NewDomain(domainNode)
			if err == nil {
				result[domain.Name] = domain
			}
		}
	} else {
		return nil, err
	}
	return result, nil
}
func (w *Watcher) LoadDomain(domainName string) (*Domain, error) {
	response, err := w.kapi.Get(context.Background(), computeDomainNodeKey(domainName, w.domainPrefix), &etcd.GetOptions{Recursive: true, Sort: true})
	if err != nil {
		return nil, err
	} else {
		domain, _ := NewDomain(response.Node)
		return domain, nil
	}

}

func (w *Watcher) PersistDomain(d *Domain) (*Domain, error) {

	if d.NodeKey != "" {
		resp, err := w.kapi.Get(context.Background(), d.NodeKey, &etcd.GetOptions{Recursive: true, Sort: false})

		if err != nil {
			return nil, err
		} else {
			oldDomain, _ := NewDomain(resp.Node)
			if oldDomain.Typ != d.Typ {
				w.kapi.Set(context.Background(), fmt.Sprintf("%s/type", d.NodeKey), d.Typ, &etcd.SetOptions{PrevExist: etcd.PrevExist})
			}

			if oldDomain.Value != d.Value {
				w.kapi.Set(context.Background(), fmt.Sprintf("%s/value", d.NodeKey), d.Value, &etcd.SetOptions{PrevExist: etcd.PrevExist})
			}
		}

	} else {
		d.NodeKey = computeDomainNodeKey(d.Name, w.domainPrefix)
		_, err := w.kapi.Create(context.Background(), fmt.Sprintf("%s/type", d.NodeKey), d.Typ)
		if err == nil {
			_, err = w.kapi.Create(context.Background(), fmt.Sprintf("%s/value", d.NodeKey), d.Value)
		}

		if err != nil {
			//Rollback creation
			log.Warnf("Rollback creation of domain %s", d.Name)
			w.kapi.Delete(context.Background(), d.NodeKey, &etcd.DeleteOptions{Recursive: true})
			return nil, err
		}

	}
	return d, nil

}
func (w *Watcher) DestroyDomain(d *Domain) error {
	_, err := w.kapi.Delete(context.Background(), d.NodeKey, &etcd.DeleteOptions{Recursive: true})
	return err

}
