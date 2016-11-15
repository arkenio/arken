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
package model

// A driver knows how to create and manage Services.
type ServiceDriver interface {
	// Creates a service and start it if asked.
	Create(s *Service, startOnCreate bool) (interface{}, error)

	// Starts a given service
	Start(s *Service) (interface{}, error)

	// Upgrades a given service
	Upgrade(s *Service) (interface{}, error)

	//Finishes the upgrade 
	FinishUpgrade(s *Service) (interface{}, error)
	
	//Rollbacks after the upgrade
	Rollback(s *Service) (interface{}, error)

	// Stops a given service
	Stop(s *Service) (interface{}, error)

	// Destroys a given service
	Destroy(s *Service) error

	// Returns a channell where ModelEvent are published by the service driver
	Listen() chan *ModelEvent

	// Returns the driver's information for a given service
	GetInfo(s *Service) (interface{}, error)


	// Tells if the service need to be upgrade when compared to its definition
	NeedToBeUpgraded(s *Service) (bool, error)
}

// This drivers allow to persist the model in a backend.
type PersistenceDriver interface {
	LoadAllServices() (map[string]*ServiceCluster, error)
	LoadService(serviceName string) (*ServiceCluster, error)
	PersistService(*Service) (*Service, error)
	DestroyService(*ServiceCluster) error

	LoadAllDomains() (map[string]*Domain, error)
	LoadDomain(serviceName string) (*Domain, error)
	PersistDomain(*Domain) (*Domain, error)
	DestroyDomain(*Domain) error

	Listen() chan *ModelEvent
}
