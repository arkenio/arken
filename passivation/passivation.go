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

package passivation

import (
	"github.com/Sirupsen/logrus"
	"github.com/arkenio/arken/goarken/model"
	"time"
)

var log = logrus.New()

type PassivationHandler struct {
	arkenModel *model.Model
	Stop       chan interface{}
}

func NewHandler(model *model.Model) *PassivationHandler {
	return &PassivationHandler{
		arkenModel: model,
		Stop:       make(chan interface{}),
	}
}

func (p *PassivationHandler) Start() {
	ticker := time.NewTicker(time.Minute)
	updateChannel := p.arkenModel.Listen()

	for {
		select {
		case <-p.Stop:
			return
		case <-ticker.C:
			// Check every minute which service has to be passivated
			for _, serviceCluster := range p.arkenModel.Services {
				p.passivateServiceIfNeeded(serviceCluster)
			}
		case event := <-updateChannel:
			// When a service changes, check if it has to be started

			if sc, ok := event.Model.(*model.Service); ok {
				service := p.arkenModel.Services[sc.Name]
				//Cluster may be nil if event was a delete
				if service != nil {
					p.restartIfNeeded(service)
				}
			}
		}
	}
}

func (p *PassivationHandler) passivateServiceIfNeeded(service *model.Service) {

	// Checking if the service should be passivated or not
	if p.hasToBePassivated(service) {
		log.Infof("Service %s enters passivation", service.Name)
		if "destroy" == service.Config.Passivation.Action {
			p.arkenModel.DestroyService(service)
		} else if "stop" == service.Config.Passivation.Action {
			p.arkenModel.StopService(service)
		} else {
			// By default passivate
			p.arkenModel.PassivateService(service)
		}

	}

}

func (p *PassivationHandler) hasToBePassivated(service *model.Service) bool {

	config := service.Config.Passivation
	if config.Enabled {
		passiveLimitDuration := time.Duration(config.DelayInSeconds) * time.Second

		return service.StartedSince() != nil &&
			time.Now().After(service.StartedSince().Add(passiveLimitDuration))
	}
	return false
}

func (p *PassivationHandler) restartIfNeeded(service *model.Service) {

	if p.hasToBeRestarted(service) {
		service, err := p.arkenModel.StartService(service)
		if err != nil {
			log.Errorf("Service "+service.Name+" restart has failed: %s", err)
			return
		}
		log.Infof("Service %s restarted", service.Name)
	}
}

func (p *PassivationHandler) hasToBeRestarted(service *model.Service) bool {
	return service.Config.Passivation.Enabled &&
		service.LastAccess != nil &&
		service.Status != nil &&
		service.Status.Expected == model.STARTED_STATUS &&
		service.Status.Current == model.PASSIVATED_STATUS

}
