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

