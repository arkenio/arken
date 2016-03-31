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
	"fmt"
	goarken "github.com/arkenio/goarken/model"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
	"io"
)

func (s *APIServer) ServiceIndex(w http.ResponseWriter, r *http.Request) {

	statusFilter := r.URL.Query().Get("status")

	services := make(map[string]*goarken.ServiceCluster)

	for _, cluster := range s.arkenModel.Services {
		for _, service := range cluster.GetInstances() {
			if statusFilter == "" || statusFilter == service.Status.Compute() {
				services[service.Name] = cluster
			}
		}
	}

	if err := json.NewEncoder(w).Encode(services); err != nil {
		panic(err)
	}
}

func (s *APIServer) ServiceShow(w http.ResponseWriter, r *http.Request) {
	serviceId := mux.Vars(r)["serviceId"]
	service := s.arkenModel.Services[serviceId]
	w.Header().Add("Content-Type","application/json")
	if s.arkenModel.Services[serviceId] != nil {
		if err := json.NewEncoder(w).Encode(service); err != nil {
			panic(err)
		}
	}
}

func (s *APIServer) run(methodName string) func(w http.ResponseWriter, r *http.Request) {
	var value reflect.Value

	return func(w http.ResponseWriter, r *http.Request) {
		serviceId := mux.Vars(r)["serviceId"]
		serviceCluster := s.arkenModel.Services[serviceId]

		if s.arkenModel.Services[serviceId] != nil {
			for _, service := range serviceCluster.GetInstances() {
				value = reflect.ValueOf(s.arkenModel)
				err := value.MethodByName(methodName).Call([]reflect.Value{reflect.ValueOf(service)})
				if err != nil {
					break
				}
			}
		}

		s.ServiceShow(w, r)
	}
}

func (s *APIServer) ServiceCreate() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)

		service := &goarken.Service{}
		service.Init()

		err := decoder.Decode(service)
		if err != nil {
			panic(err)
		}
		log.Infof("Creating service %s", service.Name)
		_, err = s.arkenModel.CreateService(service, false)

		if err != nil {
			panic(err)
		}
		w.Header().Add("Content-Type","application/json")
		if err := json.NewEncoder(w).Encode(service); err != nil {
			panic(err)
		}

	}

}

func (s *APIServer) ServiceDestroy() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serviceId := mux.Vars(r)["serviceId"]
		serviceCluster := s.arkenModel.Services[serviceId]

		err := s.arkenModel.DestroyServiceCluster(serviceCluster)
		if err == nil {
			io.WriteString(w,"{\"serviceDestroyed\":\"ok\"}")
		} else {
			io.WriteString(w,fmt.Sprintf("{\"serviceDestroyed\":\"ko\", \"error\":\"%s\"}}", err))
		}

	}
}

func (s *APIServer) ServiceStop() func(w http.ResponseWriter, r *http.Request) {
	return s.run("StopService")
}

func (s *APIServer) ServiceStart() func(w http.ResponseWriter, r *http.Request) {
	return s.run("StartService")
}

func (s *APIServer) ServicePassivate() func(w http.ResponseWriter, r *http.Request) {
	return s.run("PassivateService")
}
