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
package main

import (
	. "github.com/arkenio/arken/goarken/model"
	"github.com/arkenio/arken/goarken/storage"
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/integration"
	"github.com/coreos/etcd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
	"time"
)

type MockServiceDriver struct {
	calls  map[string]int
	events *Broadcaster
}

func NewMockServiceDriver() *MockServiceDriver {
	return &MockServiceDriver{
		calls:  make(map[string]int),
		events: NewBroadcaster(),
	}
}

func (sd *MockServiceDriver) Create(s *Service, startOnCreate bool) (interface{}, error) {
	sd.calls["create"] = sd.calls["create"] + 1
	sd.events.Write(NewModelEvent("update", s))
	return &RancherInfoType{EnvironmentId: "rancherId"}, nil
}

func (sd *MockServiceDriver) Start(s *Service) (interface{}, error) {
	sd.calls["start"] = sd.calls["start"] + 1
	sd.events.Write(NewModelEvent("update", s))
	return &RancherInfoType{EnvironmentId: "rancherId"}, nil
}

func (sd *MockServiceDriver) Upgrade(s *Service) (interface{}, error) {
	sd.calls["upgrade"] = sd.calls["upgrade"] + 1
	sd.events.Write(NewModelEvent("update", s))
	return &RancherInfoType{EnvironmentId: "rancherId"}, nil
}

func (sd *MockServiceDriver) FinishUpgrade(s *Service) (interface{}, error) {
	sd.calls["finishupgrade"] = sd.calls["finishupgrade"] + 1
	sd.events.Write(NewModelEvent("update", s))
	return &RancherInfoType{EnvironmentId: "rancherId"}, nil
}

func (sd *MockServiceDriver) Rollback(s *Service) (interface{}, error) {
	sd.calls["rollback"] = sd.calls["rollback"] + 1
	return &RancherInfoType{EnvironmentId: "rancherId"}, nil
}

func (sd *MockServiceDriver) Stop(s *Service) (interface{}, error) {
	sd.calls["stop"] = sd.calls["stop"] + 1
	sd.events.Write(NewModelEvent("update", s))
	return &RancherInfoType{EnvironmentId: "rancherId"}, nil
}

func (sd *MockServiceDriver) Destroy(s *Service) error {
	sd.calls["destroy"] = sd.calls["destroy"] + 1
	sd.events.Write(NewModelEvent("update", s))
	return nil
}

func (sd *MockServiceDriver) Listen() chan *ModelEvent {
	return FromInterfaceChannel(sd.events.Listen())
}

func (sd *MockServiceDriver) GetInfo(s *Service) (interface{}, error) {
	return &RancherInfoType{EnvironmentId: "rancherId"}, nil
}

func (w *MockServiceDriver) StopDriver() {

}

func (w *MockServiceDriver) NeedToBeUpgraded(s *Service) (bool, error) {
	return false, nil
}

func Test_EtcdWatcher(t *testing.T) {
	//Wait for potential other etcd cluster to stop
	time.Sleep(3 * time.Second)

	defer testutil.AfterTest(t)
	cl := integration.NewCluster(t, 1)
	cl.Launch(t)
	defer cl.Terminate(t)

	u := cl.URL(0)

	cfg := client.Config{
		Endpoints:               []string{u},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	kapi := client.NewKeysAPI(c)
	if err != nil {
		panic(err)
	}

	var model *Model

	sd := NewMockServiceDriver()
	Convey("Given a model", t, func() {

		pd := storage.NewWatcher(kapi, "/services", "/domains")
		model, _ = NewArkenModel(sd, pd)

		for _, s := range model.Services {
			model.DestroyService(s)
		}

		for _, d := range model.Domains {
			model.DestroyDomain(d)
		}

		Convey("When i create a service", func() {
			initialCreateCount := sd.calls["create"]
			service := &Service{}
			service.Init()

			service.Name = "testService"

			service, err := model.CreateService(service, false)

			Convey("Then the service should be available in all services", func() {
				So(err, ShouldBeNil)
				So(len(model.Services), ShouldEqual, 1)
				sc := model.Services["testService"]
				So(sc, ShouldNotBeNil)
			})

			Convey("Then its status should be stopped", func() {

				service := model.Services["testService"]
				st := StatusError{service.Status.Compute(), service.Status}
				So(st.ComputedStatus, ShouldEqual, STOPPED_STATUS)
			})

			Convey("Then the service should be created in the backend", func() {
				time.Sleep(time.Second)
				So(sd.calls["create"], ShouldEqual, initialCreateCount+1)
				instance := model.Services["testService"]

				So(instance.Config, ShouldNotBeNil)
				So(instance.Config.RancherInfo, ShouldNotBeNil)
				So(instance.Config.RancherInfo.EnvironmentId, ShouldEqual, "rancherId")
			})

			Convey("When I start the service", func() {
				initialStartCount := sd.calls["start"]
				model.StartService(service)

				Convey("Then the service should be started in the backend", func() {
					So(sd.calls["start"], ShouldEqual, initialStartCount+1)
				})

				Convey("Then its status should be starting", func() {
					So(getServiceStatus(model, "testService"), ShouldEqual, STARTING_STATUS)
				})

			})

			Convey("When I start the service and the service is started", func() {
				model.StartService(service)

				service := model.Services[service.Name]
				service.Status.Current = STARTED_STATUS
				service.Status.Alive = "1"

				Convey("Then its status should be started", func() {
					So(getServiceStatus(model, "testService"), ShouldEqual, STARTED_STATUS)

				})

			})

		})

		Convey("When I create a service with a domain", func() {
			service := &Service{}
			service.Init()

			service.Name = "testService"
			service.Domain = "test.domain.com"

			model.CreateService(service, false)
			Convey("Then a domain should be created", func() {
				time.Sleep(2 * time.Second)
				So(len(model.Domains), ShouldEqual, 1)
				So(model.Domains["test.domain.com"], ShouldNotBeNil)
			})
		})

		Convey("Given a started a service ", func() {
			service := &Service{}
			service.Init()
			service.Name = "testService"
			model.CreateService(service, true)

			Convey("The actions on the service should be stop, delete, update", func() {
				actions := make([]string, 0)
				actions = append(actions, START_ACTION, DELETE_ACTION, UPDATE_ACTION)
				So(reflect.DeepEqual(service.Actions, actions), ShouldEqual, true)
			})

			Convey("When i passivate the service", func() {
				initial := sd.calls["stop"]
				service, err = model.PassivateService(service)
				Convey("Then the service is stopped and in passivated status", func() {
					So(err, ShouldBeNil)
					So(service.Status.Compute(), ShouldEqual, PASSIVATED_STATUS)
					So(sd.calls["stop"], ShouldEqual, initial+1)
				})
			})

		})

	})
}

func getServiceStatus(model *Model, serviceName string) string {
	service := model.Services[serviceName]
	st := StatusError{service.Status.Compute(), service.Status}
	return st.ComputedStatus
}
