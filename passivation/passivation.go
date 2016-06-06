package passivation

import (
	"github.com/Sirupsen/logrus"
	"github.com/arkenio/goarken/model"
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
			if sc, ok := event.Model.(*model.ServiceCluster); ok {
				for _, service := range p.arkenModel.Services[sc.Name].Instances {
					p.restartIfNeeded(service)
				}
			}
		}
	}
}

func (p *PassivationHandler) passivateServiceIfNeeded(serviceCluster *model.ServiceCluster) {

	service, err := serviceCluster.Next()
	if err != nil {
		//No active instance, no need to passivate
		return
	}

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
