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
	"github.com/arkenio/goarken"
	"github.com/gorilla/mux"
	"net/http"
	"reflect"
)

func (s *APIServer) ServiceIndex(w http.ResponseWriter, r *http.Request) {

	statusFilter := r.URL.Query().Get("status")

	services := make(map[string]*goarken.ServiceCluster)

	for _, cluster := range s.watcher.Services {
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
	service := s.watcher.Services[serviceId]

	if s.watcher.Services[serviceId] != nil {
		if err := json.NewEncoder(w).Encode(service); err != nil {
			panic(err)
		}
	}
}


func (s *APIServer) run(methodName string) func(w http.ResponseWriter, r *http.Request ) {
	var value reflect.Value

	return func(w http.ResponseWriter, r *http.Request) {
		serviceId := mux.Vars(r)["serviceId"]
		serviceCluster := s.watcher.Services[serviceId]

		if s.watcher.Services[serviceId] != nil {
			for _, service := range serviceCluster.GetInstances() {
				value = reflect.ValueOf(service)
				err := value.MethodByName(methodName).Call([]reflect.Value{reflect.ValueOf(s.client)})
				if err != nil {
					break
				}
			}
		}

		s.ServiceShow(w,r)
	}
}


func (s *APIServer) ServiceStop() func(w http.ResponseWriter, r *http.Request) {
	return s.run("Stop")
}

func (s *APIServer) ServiceStart() func(w http.ResponseWriter, r *http.Request) {
	return s.run("Start")
}

func (s *APIServer) ServicePassivate() func(w http.ResponseWriter, r *http.Request) {
	return s.run("Passivate")
}

