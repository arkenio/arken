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
	"github.com/arkenio/arken/goarken/model"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *APIServer) DomainIndex(w http.ResponseWriter, r *http.Request) {

	statusFilter := r.URL.Query().Get("status")

	domains := make(map[string]*model.ServiceCluster)

	for domainName, domain := range s.arkenModel.Domains {
		if domain.Typ == "service" {
			cluster := s.arkenModel.Services[domain.Value]
			if cluster != nil {
				for _, service := range cluster.GetInstances() {
					if statusFilter == "" || statusFilter == service.Status.Compute() {
						domains[domainName] = cluster
					}
				}
			}
		}

	}

	if err := json.NewEncoder(w).Encode(domains); err != nil {
		panic(err)
	}
}

func (s *APIServer) DomainShow(w http.ResponseWriter, r *http.Request) {
	domainName := mux.Vars(r)["domain"]
	domain := s.arkenModel.Domains[domainName]

	if domain != nil {

		if domain.Typ == "service" {
			cluster := s.arkenModel.Services[domain.Value]
			if cluster != nil {

				if err := json.NewEncoder(w).Encode(cluster); err != nil {
					panic(err)
				}

			}
		} else {
			if err := json.NewEncoder(w).Encode(domain); err != nil {
				panic(err)
			}

		}

	}
}
