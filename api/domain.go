package api

import (
	"encoding/json"
	"github.com/arkenio/goarken"
	"github.com/gorilla/mux"
	"net/http"
)

func (s *APIServer) DomainIndex(w http.ResponseWriter, r *http.Request) {

	statusFilter := r.URL.Query().Get("status")

	domains := make(map[string]*goarken.ServiceCluster)

	for domainName, domain := range s.watcher.Domains {
		if domain.Typ == "service" {
			cluster := s.watcher.Services[domain.Value]
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
	domain := s.watcher.Domains[domainName]

	if domain != nil {

		if domain.Typ == "service" {
			cluster := s.watcher.Services[domain.Value]
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
