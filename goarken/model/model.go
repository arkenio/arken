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

import (
	"errors"
	"fmt"
	"github.com/Sirupsen/logrus"
	"time"
)

var log = logrus.New()

/*
The Arken model is a structure that holds a map of Services and
a map of Domain. A Model MUST be backed by a PersistenceDriver and
MAY drivre ServiceDriver.
*/
type Model struct {
	serviceDriver     ServiceDriver
	persistenceDriver PersistenceDriver

	Domains        map[string]*Domain
	Services       map[string]*Service
	eventBroadcast *Broadcaster
	eventBuffer    *eventBuffer
}

// Create an ArkenModel base on a serviceDriver and a PersistenceDriver. The
// ServiceDriver is optional.
func NewArkenModel(sDriver ServiceDriver, pDriver PersistenceDriver) (*Model, error) {
	if pDriver == nil {
		return nil, errors.New("Can't use a nil persistence Driver for Arken model")
	}

	model := &Model{
		Domains:           make(map[string]*Domain),
		Services:          make(map[string]*Service),
		serviceDriver:     sDriver,
		persistenceDriver: pDriver,
		eventBroadcast:    NewBroadcaster(),
	}
	model.eventBuffer = newEventBuffer(model.eventBroadcast)
	err := model.Init()

	if err == nil {
		return model, nil
	} else {
		return nil, err
	}
}

// Returns a channel of ModelEvent where all changes in the model (Service or Domain)
// are published.
func (m *Model) Listen() chan *ModelEvent {
	return FromInterfaceChannel(m.eventBroadcast.Listen())
}

// Inits the model. It loads the model from the persistence driver and then listen
// to changes. It also starts listening on service driver updates.
func (m *Model) Init() error {

	//Load initial data
	domains, err := m.persistenceDriver.LoadAllDomains()
	if err != nil {
		return err
	}
	m.Domains = domains

	services, err := m.persistenceDriver.LoadAllServices()
	if err != nil {
		return err
	}
	m.Services = services

	// Listen to external events
	go m.handlePersistenceModelEventOn(m.persistenceDriver.Listen())

	if m.serviceDriver != nil {
		go m.handlePersistenceModelEventOn(m.serviceDriver.Listen())
	}

	// Launch synchronization
	go m.keepServiceDriverInSync(time.Duration(15) * time.Second)

	// Launch event buffer
	go m.eventBuffer.run(time.Second)

	return nil

}

// Regulary fetch information from service driver, since we may have missed an update
// event (arken stopped or something else).
func (m *Model) keepServiceDriverInSync(duration time.Duration) {

	startedTicker := time.NewTicker(duration)
	allTicker := time.NewTicker(time.Duration(40) * duration)
	defer startedTicker.Stop()
	defer allTicker.Stop()

	//Flag to prevent both ticker to be executed at the same time
	allTickerPassed := false
	for {
		select {
		case <-allTicker.C:
			m.syncServiceByStatus(nil)
			allTickerPassed = true
		case <-startedTicker.C:
			if allTickerPassed == false {
				m.syncServiceByStatus([]string{"started", "error"})
			} else {
				allTickerPassed = false
			}
		}
	}
}

// Launch a synchronize on all Service that are in one of the given statuses
func (m *Model) syncServiceByStatus(status []string) {
	m.onAllService(func(service *Service) {
		if status == nil || inArray(status, service.Status.Compute()) {
			m.SyncService(service)
		}
	})
}

// Synchronize ServiceDriver state with internal Arken state
func (m *Model) SyncService(service *Service) {
	if m.serviceDriver == nil {
		return
	}
	info, err := m.serviceDriver.GetInfo(service)
	if err != nil {
		log.Warningf("Unable to get Status from service driver on %v", service.Name)
	} else {
		if info, ok := info.(*RancherInfoType); ok {
			m.onRancherInfo(info)
		}
	}
}

// Looks for a string in a string array and return true if the string is found.
// This method is suitable for small array since it does a sequential scan on it.
func inArray(array []string, seek string) bool {
	for _, str := range array {
		if str == seek {
			return true
		}
	}
	return false
}

// Helper to execute a function on each service of the model
func (m *Model) onAllService(serviceHandler func(s *Service)) {
	for _, service := range m.Services {
		serviceHandler(service)
	}
}

// Creates a Service and starts it if asked. If the Domain of the service is provided, then the
// corresponding domain is also created.
func (m *Model) CreateService(service *Service, startOnCreate bool) (*Service, error) {

	s, err := m.persistenceDriver.PersistService(service)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to persist service %s in etcd : %s", service.Name, err.Error()))
	}

	if m.serviceDriver != nil {
		info, err := m.serviceDriver.Create(s, startOnCreate)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Unable to create service in backend : %s", s.Name, err.Error()))
		}

		m.updateInfoFromDriver(s, info)
	}

	s, err = m.saveService(s)
	if err != nil {
		return nil, err
	}

	if s.Domain != "" {

		if domain, ok := m.Domains[s.Domain]; ok {
			domain.Typ = "service"
			domain.Value = s.Name
			_, err = m.UpdateDomain(domain)
			if err != nil {
				log.Errorf("Unable to update domain %s for service %s : %v", s.Domain, s.Name, err)
			}
		} else {
			_, err := m.CreateDomain(&Domain{Name: s.Domain, Typ: "service", Value: s.Name})
			if err != nil {
				log.Errorf("Unable to create domain %s for service %s : %v", s.Domain, s.Name, err)
			}
		}
	}

	m.eventBuffer.events <- NewModelEvent("create", s)

	return s, nil

}

// Creates a Domain
func (m *Model) CreateDomain(domain *Domain) (*Domain, error) {
	domain, err := m.persistenceDriver.PersistDomain(domain)
	if err != nil {
		return nil, err
	} else {
		m.eventBuffer.events <- NewModelEvent("create", domain)
		return domain, nil
	}
}

//Destroys a Domain
func (m *Model) DestroyDomain(domain *Domain) error {

	err := m.persistenceDriver.DestroyDomain(domain)
	if err != nil {
		return err
	} else {
		m.eventBuffer.events <- NewModelEvent("delete", domain)
		return nil
	}

}

// Updates a domain
func (m *Model) UpdateDomain(domain *Domain) (*Domain, error) {
	domain, err := m.persistenceDriver.PersistDomain(domain)
	if err != nil {
		return nil, err
	} else {
		m.eventBuffer.events <- NewModelEvent("update", domain)
		return domain, nil
	}
}

// Starts a service (only works if ServiceDriver is set)
func (m *Model) StartService(service *Service) (*Service, error) {

	if m.serviceDriver != nil {
		info, err := m.serviceDriver.Start(service)
		if err != nil {
			return nil, err
		}
		m.updateInfoFromDriver(service, info)
	}

	service.Status.Expected = STARTED_STATUS
	service.Status.Current = STARTING_STATUS
	AddAction(service, STOP_ACTION, UPDATE_ACTION, DELETE_ACTION)
	service, err := m.saveService(service)

	if err != nil {
		return nil, err
	} else {
		m.eventBuffer.events <- NewModelEvent("update", service)
		return service, nil
	}
}

// Stops a service (only works if ServiceDriver is set)
func (m *Model) StopService(service *Service) (*Service, error) {
	service.Status.Expected = STOPPED_STATUS
	AddAction(service, START_ACTION, DELETE_ACTION)

	if m.serviceDriver != nil {
		info, err := m.serviceDriver.Stop(service)
		if err != nil {
			return nil, err
		}

		m.updateInfoFromDriver(service, info)
	}

	service, err := m.saveService(service)

	if err != nil {
		return nil, err
	} else {
		m.eventBuffer.events <- NewModelEvent("update", service)
		return service, nil
	}
}

// Passivates a service (only works if ServiceDriver is set)
func (m *Model) PassivateService(service *Service) (*Service, error) {
	service.Status.Expected = PASSIVATED_STATUS
	AddAction(service, DELETE_ACTION)
	info, err := m.serviceDriver.Stop(service)
	if err != nil {
		return nil, err
	}

	m.updateInfoFromDriver(service, info)
	service.Status.Current = PASSIVATED_STATUS
	service, err = m.saveService(service)

	if err != nil {
		return nil, err
	} else {
		m.eventBuffer.events <- NewModelEvent("update", service)
		return service, nil
	}
}

func (m *Model) UpdateService(service *Service) (*Service, error) {

	if origService, ok := m.Services[service.Name]; !ok {
		return nil, errors.New("Service not found")
	} else {

		if service.Config != nil {

			if service.Config.Environment != nil {
				origService.Config.Environment = service.Config.Environment
			}

			if service.Config.Passivation != nil {
				origService.Config.Passivation = service.Config.Passivation
			}
		}

		//Updates the domain of the service
		if service.Domain != "" && service.Domain != origService.Domain {
			if oldDomain, ok := m.Domains[origService.Domain]; ok {
				m.DestroyDomain(oldDomain)

				if domain, ok := m.Domains[service.Domain]; ok {
					domain.Typ = "service"
					domain.Value = service.Name
					_, err := m.UpdateDomain(domain)
					if err != nil {
						log.Errorf("Unable to update domain %s for service %s : %v", service.Domain, service.Name, err)
					}
				} else {
					_, err := m.CreateDomain(&Domain{Name: service.Domain, Typ: "service", Value: service.Name})
					if err != nil {
						log.Errorf("Unable to create domain %s for service %s : %v", service.Domain, service.Name, err)
					}
				}
			}

			origService.Domain = service.Domain

		}
		//update the actions available on the service if the service needs to be upgraded
		var updated, err = m.NeedToBeUpgraded(origService)
		if err != nil {
			log.Errorf("Unable to read service %s : %v ", service.Name, err)
		}
		if updated {
			AddAction(origService, UPGRADE_ACTION)
		}

		m.saveService(origService)

		return m.Services[service.Name], nil
	}
}

func (m *Model) NeedToBeUpgraded(service *Service) (bool, error) {
	if _, ok := m.Services[service.Name]; !ok {
		return false, errors.New("Service not found")
	} else {
		return m.serviceDriver.NeedToBeUpgraded(service)

	}
}

func (m *Model) UpgradeService(service *Service) (*Service, error) {

	if s, ok := m.Services[service.Name]; !ok {
		return nil, errors.New("Service not found")
	} else {
		_, err := m.serviceDriver.Upgrade(s)
		if err != nil {
			return nil, err
		} else {
			s.Status.Expected = STARTED_STATUS
			s.Status.Current = STARTING_STATUS
			AddAction(s, FINISHUPGRADE_ACTION, ROLLBACK_ACTION)
			s, err = m.saveService(s)
			if err != nil {
				return nil, err
			}
			return s, nil
		}
	}
	return nil, errors.New("No service in cluster ! Doh !")

}

func (m *Model) FinishUpgradeService(service *Service) (*Service, error) {

	if m.serviceDriver != nil {
		info, err := m.serviceDriver.FinishUpgrade(service)
		if err != nil {
			return nil, err
		}
		m.updateInfoFromDriver(service, info)
	}

	service.Status.Expected = STARTED_STATUS
	service.Status.Current = STARTING_STATUS
	AddAction(service, UPDATE_ACTION)
	service, err := m.saveService(service)

	if err != nil {
		return nil, err
	} else {
		m.eventBuffer.events <- NewModelEvent("update", service)
		return service, nil
	}
}

func (m *Model) RollbackService(service *Service) (*Service, error) {

	if m.serviceDriver != nil {
		info, err := m.serviceDriver.Rollback(service)
		if err != nil {
			return nil, err
		}
		m.updateInfoFromDriver(service, info)
	}

	service.Status.Expected = STARTED_STATUS
	service.Status.Current = STARTING_STATUS
	AddAction(service, UPDATE_ACTION)
	service, err := m.saveService(service)

	if err != nil {
		return nil, err
	}
	return service, nil

}

// Destroys a service (only works if ServiceDriver is set)
func (m *Model) DestroyService(service *Service) error {
	if m.serviceDriver != nil {
		err := m.serviceDriver.Destroy(service)
		if err != nil {
			return err
		}
	}

	error := m.persistenceDriver.DestroyService(m.Services[service.Name])
	if error != nil {
		return error
	} else {
		m.eventBuffer.events <- NewModelEvent("delete", service)
		return nil
	}
}

func (m *Model) saveService(service *Service) (*Service, error) {
	return m.persistenceDriver.PersistService(service)
}

func (m *Model) updateInfoFromDriver(service *Service, info interface{}) {
	if rancherInfo, ok := info.(*RancherInfoType); ok {
		service.Config.RancherInfo = rancherInfo
	}

	if fleetInfo, ok := info.(*FleetInfoType); ok {
		service.Config.FleetInfo = fleetInfo
	}
	m.eventBuffer.events <- NewModelEvent("update", service)

}

func (m *Model) handlePersistenceModelEventOn(eventStream chan *ModelEvent) {

	for {
		event := <-eventStream

		switch event.EventType {
		case "create":
		case "update":
			if sc, ok := event.Model.(*Service); ok {
				m.Services[sc.Name] = sc
				m.eventBuffer.events <- event
			} else if domain, ok := event.Model.(*Domain); ok {
				m.Domains[domain.Name] = domain
				m.eventBuffer.events <- event
			} else if info, ok := event.Model.(*RancherInfoType); ok {
				m.onRancherInfo(info)
			}

		case "delete":
			if sc, ok := event.Model.(*Service); ok {
				delete(m.Services, sc.Name)
				m.eventBuffer.events <- event
			} else if domain, ok := event.Model.(*Domain); ok {
				delete(m.Domains, domain.Name)
			}
			m.eventBuffer.events <- event
		}

	}

}

func (m *Model) onRancherInfo(info *RancherInfoType) {
	service := m.Services[info.EnvironmentName]
	if service != nil {
		service.Config.RancherInfo = info

		if !service.Location.Equals(info.Location) {
			log.Infof("Service %s changed location from %s to %s", service.Name, service.Location, info.Location)
			service.Location = info.Location

		}

		// Save last status
		computedSatus := service.Status.Compute()

		service.Status.Current = info.CurrentStatus
		//If service is stopped it may be passivated
		if info.CurrentStatus == STOPPED_STATUS && service.Status.Expected == PASSIVATED_STATUS {
			service.Status.Current = PASSIVATED_STATUS
		}

		if service.Status.Current == STARTED_STATUS {
			service.Status.Alive = "1"
		} else {
			service.Status.Alive = ""
		}

		// Compare to initial status and update actions as the service is restarted in case of upgrade and rollback
		newStatus := service.Status.Compute()
		if computedSatus != newStatus {
			log.Infof("Service %s changed its status to : %s", service.Name, newStatus)
			if STOPPED_STATUS == newStatus {
				AddAction(service, START_ACTION)
			}
			if STARTED_STATUS == newStatus {
				AddAction(service, STOP_ACTION)
			}
		}

		s, err := m.persistenceDriver.PersistService(service)

		if err != nil {
			log.Errorf("Error when persisting rancher update : %s", err.Error())
			log.Errorf("Rancher update was : %s", info)
		} else {
			m.eventBuffer.events <- NewModelEvent("update", s)
		}

	}
}
