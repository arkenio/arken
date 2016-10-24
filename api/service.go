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
package api

import (
	"encoding/json"
	"errors"
	"fmt"
	goarken "github.com/arkenio/arken/goarken/model"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"reflect"
)

func (s *APIServer) ServiceIndex(w http.ResponseWriter, r *http.Request) {

	statusFilter := r.URL.Query().Get("status")

	services := make(map[string]*goarken.Service)

	for _, service := range s.arkenModel.Services {
		if statusFilter == "" || statusFilter == service.Status.Compute() {
			services[service.Name] = service
		}
	}

	if err := json.NewEncoder(w).Encode(services); err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (s *APIServer) ServiceShow(w http.ResponseWriter, r *http.Request) {
	serviceId := mux.Vars(r)["serviceId"]

	if s, ok := s.arkenModel.Services[serviceId]; ok {
		//create new instance to override the actions with a pretty format
		w.Header().Add("Content-Type", "application/json")
		ss := *s
		ss.Actions = goarken.GetPrettyActions(&ss, r.URL)
		if err := json.NewEncoder(w).Encode(ss); err != nil {
			http.Error(w, err.Error(), 500)
		}
	} else {
		http.NotFound(w, r)
	}
}

func (s *APIServer) run(methodName string) func(w http.ResponseWriter, r *http.Request) {
	var value reflect.Value

	return func(w http.ResponseWriter, r *http.Request) {

		serviceId := mux.Vars(r)["serviceId"]

		if service, ok := s.arkenModel.Services[serviceId]; ok {
			value = reflect.ValueOf(s.arkenModel)
			err := value.MethodByName(methodName).Call([]reflect.Value{reflect.ValueOf(service)})
			if err != nil {
				http.NotFound(w, r)
			}
			s.ServiceShow(w, r)
		} else {
			http.NotFound(w, r)
		}

	}
}

func (s *APIServer) ServiceCreate() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		service := &goarken.Service{}
		service.Init()

		err := decoder.Decode(service)
		service.Status = goarken.NewInitialStatus(goarken.STOPPED_STATUS, service)
		service.Actions = make([]string, 0)
		service.Actions = goarken.InitActions(service)

		if service.Config.Passivation == nil {
			service.Config.Passivation = goarken.DefaultPassivation()
		}

		if err != nil {
			log.Errorf("Error when decoding service %s : %s", service.Name, err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		log.Infof("Creating service %s", service.Name)
		_, err = s.arkenModel.CreateService(service, false)
		if err != nil {
			log.Errorf("Error when creating service %s : %s", service.Name, err.Error())
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

	}

}

func (s *APIServer) ServiceDestroy() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceId := mux.Vars(r)["serviceId"]
		service := s.arkenModel.Services[serviceId]
		err := s.arkenModel.DestroyService(service)
		if err == nil {
			io.WriteString(w, "{\"serviceDestroyed\":\"ok\"}")
		} else {
			io.WriteString(w, fmt.Sprintf("{\"serviceDestroyed\":\"ko\", \"error\":\"%s\"}}", err))
		}

	}
}

type NotFoundError struct {
	serviceId string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Service %s not found", e.serviceId)
}

func (s *APIServer) ServiceAction() func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		serviceId := mux.Vars(r)["serviceId"]
		serviceAction := r.URL.Query().Get("action")

		if serviceCluster, ok := s.arkenModel.Services[serviceId]; !ok {
			http.Error(w, "Service not found", http.StatusNotFound)
		} else {
			err := s.runMethodFromAction(r, serviceAction, serviceCluster)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				s.ServiceShow(w, r)
			}
		}
	}

}

func (s *APIServer) ServiceUpdate() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		updatedService := &goarken.Service{}
		err := decoder.Decode(updatedService)

		if err != nil {
			http.Error(w, "Unable to read service : "+err.Error(), http.StatusBadRequest)
		}
		serviceId := mux.Vars(r)["serviceId"]

		if _, ok := s.arkenModel.Services[serviceId]; !ok {
			http.Error(w, "Service not found", http.StatusNotFound)
		} else {
			s.arkenModel.UpdateService(updatedService)
		}
	}
}

func (s *APIServer) runMethodFromAction(r *http.Request, actionName string, service *goarken.Service) error {
	var err error
	switch actionName {
	case "start":
		_, err = s.arkenModel.StartService(service)
	case "stop":
		_, err = s.arkenModel.StopService(service)
	case "passivate":
		s.arkenModel.PassivateService(service)
	case "upgrade":
		s.arkenModel.UpgradeService(service)
	case "finishupgrade":
		s.arkenModel.FinishUpgradeService(service)
	case "rollback":
		s.arkenModel.RollbackService(service)
	default:
		return errors.New("Method not available")
	}
	return err
}

func (s *APIServer) serviceNeedToBeUpgraded() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceId := mux.Vars(r)["serviceId"]
		if service, ok := s.arkenModel.Services[serviceId]; !ok {
			http.Error(w, "Service not found", http.StatusNotFound)
		} else {
			log.Infof("Check if service needs to be upgarded for the service %v", serviceId)
			if serviceId == service.Name {
				s.arkenModel.NeedToBeUpgraded(service)
			}
		}

	}
}
