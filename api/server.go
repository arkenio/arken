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
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/pjebs/restgate"
	"github.com/spf13/viper"
	"gopkg.in/tylerb/graceful.v1"
	"html/template"
	"time"
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

	log.Info(fmt.Sprintf("Starting Arken API server on port : %d", s.port))
	log.Info(fmt.Sprintf("   with driver : %s", viper.GetString("driver")))

	app := negroni.New()
	app.Use(negroni.NewRecovery())
	app.UseHandler(s.getRoutes())


	graceful.Run(fmt.Sprintf(":%d", s.port), 5*time.Second, app)

}

func (s *APIServer) getRoutes() *mux.Router {

	negAPI := negroni.New()
	if gate := s.getRestGate(); gate != nil {
		negAPI.Use(gate)
	}
	negAPI.Use(negronilogrus.NewMiddleware())
	negAPI.UseHandler(s.getAPIRouter())

	mainRouter := mux.NewRouter().StrictSlash(true)

	mainRouter.PathPrefix("/doc").Handler(http.FileServer(FS(false)))
	mainRouter.PathPrefix("/swagger.yaml").HandlerFunc(serveSwaggerYaml)
	mainRouter.PathPrefix("/api").Handler(negAPI)

	return mainRouter
}


func (s *APIServer) getRestGate() *restgate.RESTGate {
	if apiKeys := viper.GetStringMap("apiKeys"); len(apiKeys) > 0 {
		Key := make([]string, len(apiKeys))
		Secret := make([]string, len(apiKeys))
		i := 0
		for k, value := range apiKeys {
			key, value := s.extractKeyValueFromConf(value)
			if key != "" && value != "" {
				Key[i] = key
				log.Infof("Adding key for %s : %s", k, Key[i])
				Secret[i] = value

				i++
			} else {
				log.Warnf("Unable to parse accessKey %s", k)
			}
		}

		return restgate.New("AuthKey", "AuthSecret", restgate.Static, restgate.Config{Context: nil, Key: Key, Secret: Secret, HTTPSProtectionOff: true})
	} else {
		return nil
	}
}

func (s *APIServer) extractKeyValueFromConf(value interface{}) (string, string) {
	var k,v string
	defer func() (string,string) {
		if r := recover(); r != nil {
			return "",""
		}
		return k,v
	}()

	mapstring := value.(map[interface{}]interface{})
	k = mapstring[string("accessKey")].(string)
	v = mapstring[string("secretKey")].(string)
	return k,v

}


func (s *APIServer) getAPIRouter() *mux.Router {


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
			"ServiceAction",
			"PUT",
			"/services/{serviceId}",
			s.ServiceAction(),
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

	apiRouter := mux.NewRouter()
	for _, route := range routes {
		apiRouter.
			Methods(route.Method).
			Path(fmt.Sprintf("/api/v1%s", route.Pattern)).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}

	return apiRouter
}

func serveSwaggerYaml(w http.ResponseWriter, r *http.Request) {

	type TemplateVars struct {
		Host string
	}

	swaggerTpl := FSMustString(false, "/swagger.tpl")
	t := template.Must(template.New("swagger").Parse(swaggerTpl))

	t.Execute(w, &TemplateVars{r.RemoteAddr})
}
