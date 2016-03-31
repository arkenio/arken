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
	"net/http"

	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/arkenio/goarken/model"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

// Create a new instance of the logger. You can have any number of instances.
var log = logrus.New()

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

type APIServer struct {
	arkenModel *model.Model
	port       int
}

func NewAPIServer(model *model.Model) *APIServer {
	return &APIServer{
		arkenModel: model,
		port:       viper.GetInt("port"),
	}
}

func (s *APIServer) Start() {

	var routes = Routes{
		Route{
			"ServiceIndex",
			"GET",
			"/services",
			s.ServiceIndex,
		},
		Route{
			"ServiceShow",
			"GET",
			"/services/{serviceId}",
			s.ServiceShow,
		},
		Route{
			"ServiceCreate",
			"POST",
			"/services",
			s.ServiceCreate(),
		},
		Route{
			"ServiceDelete",
			"DELETE",
			"/services/{serviceId}",
			s.ServiceDestroy(),
		},
		Route{
			"ServiceStop",
			"PUT",
			"/services/{serviceId}/stop",
			s.ServiceStop(),
		},
		Route{
			"ServiceStart",
			"PUT",
			"/services/{serviceId}/start",
			s.ServiceStart(),
		},
		Route{
			"ServicePassivate",
			"PUT",
			"/services/{serviceId}/passivate",
			s.ServicePassivate(),
		},
		Route{
			"DomainShow",
			"GET",
			"/domains/{domain}",
			s.DomainShow,
		},
		Route{
			"DomainIndex",
			"GET",
			"/domains",
			s.DomainIndex,
		},
	}

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(fmt.Sprintf("/api/v1%s", route.Pattern)).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	log.Info(fmt.Sprintf("Starting Arken API server on port : %d", s.port))
	log.Info(fmt.Sprintf("   with driver : %s", viper.GetString("driver")))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", s.port), router))
}
