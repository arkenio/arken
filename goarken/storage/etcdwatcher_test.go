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
package storage

import (
	"fmt"
	"github.com/arkenio/arken/goarken/model"
	. "github.com/arkenio/arken/goarken/model"
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/integration"
	"github.com/coreos/etcd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"sync"
	"testing"
	"time"
)

func wait(listening chan *ModelEvent, wg *sync.WaitGroup) {
	defer wg.Done()
	<-listening

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

	testServiceName := "testService"

	var w *Watcher
	var updateChan chan *model.ModelEvent

	notifCount := 0

	Convey("Given a Model with one service", t, func() {
		kapi.Delete(context.Background(), "/domains", &client.DeleteOptions{Recursive: true})
		kapi.Delete(context.Background(), "/services", &client.DeleteOptions{Recursive: true})

		w = NewWatcher(kapi, "/services", "/domains")
		updateChan = w.Listen()

		go func() {
			for {
				select {
				case <-updateChan:
					notifCount = notifCount + 1
				}
			}
		}()

		service := &Service{Name: testServiceName}
		service.Init()

		w.PersistService(service)

		Convey("When i get an etcd node from that service", func() {
			resp, _ := kapi.Get(context.Background(), fmt.Sprintf("/services/%s/1/status/expected", testServiceName), &client.GetOptions{Recursive: true})
			Convey("Then it can get the env key from it", func() {
				So(resp, ShouldNotBeNil)
				name, error := getEnvForNode(resp.Node)
				So(error, ShouldBeNil)
				So(name, ShouldEqual, testServiceName)
			})

		})

		Convey("When I create a Service", func() {

			Convey("Then the list of service contains 1 service", func() {
				services, _ := w.LoadAllServices()
				So(len(services), ShouldEqual, 1)
			})

			Convey("Then it is able to load that service", func() {
				sc, _ := w.LoadService(testServiceName)
				So(sc, ShouldNotBeNil)
			})

		})

		Convey("When i destroy a service", func() {
			services, _ := w.LoadAllServices()
			nbServices := len(services)

			sc, _ := w.LoadService(testServiceName)

			w.DestroyService(sc)

			Convey("Then the service should be destroyed", func() {
				sc, _ := w.LoadService("serviceToBeDestroyed")
				So(sc, ShouldBeNil)
			})

			Convey("Then the number of services should have decreased", func() {
				services, _ := w.LoadAllServices()
				So(len(services), ShouldEqual, nbServices-1)
			})
		})

		Convey("When i modify a service", func() {

			initialNotifCount := notifCount
			sc, _ := w.LoadService(testServiceName)

			service = sc.Instances[0]

			service.Status.Expected = STARTED_STATUS
			service.Config.RancherInfo = &RancherInfoType{EnvironmentId: "bla"}

			w.PersistService(service)

			Convey("Then the service should be modified", func() {
				sc, _ := w.LoadService(testServiceName)
				service = sc.Instances[0]
				So(service.Status.Expected, ShouldEqual, STARTED_STATUS)
				So(service.Config.RancherInfo.EnvironmentId, ShouldEqual, "bla")
			})

			Convey("Then notification should have been sent", func() {
				So(notifCount, ShouldBeGreaterThan, initialNotifCount)
			})

		})
	})

	Convey("Given a Model with one domain", t, func() {
		kapi.Delete(context.Background(), "/domains", &client.DeleteOptions{Recursive: true})
		kapi.Delete(context.Background(), "/services", &client.DeleteOptions{Recursive: true})

		w = NewWatcher(kapi, "/services", "/domains")
		updateChan = w.Listen()

		go func() {
			for {
				select {
				case <-updateChan:
					notifCount = notifCount + 1
				}
			}
		}()

		domain := &Domain{}
		domain.Name = "test.domain.com"
		domain.Typ = "service"
		domain.Value = "testService"

		w.PersistDomain(domain)

		Convey("They key of that domain shoulde be well computed", func() {
			So(computeDomainNodeKey(domain.Name, "domains"), ShouldEqual, "/domains/test.domain.com")
		})

		Convey("When i get an etcd node from that domain", func() {
			resp, _ := kapi.Get(context.Background(), fmt.Sprintf("/domains/%s/type", domain.Name), &client.GetOptions{Recursive: true})
			So(resp, ShouldNotBeNil)
			Convey("Then it can get the env key from it", func() {
				name, _ := getDomainForNode(resp.Node)
				So(name, ShouldEqual, domain.Name)
			})

		})

		Convey("When it load all domains", func() {
			domains, _ := w.LoadAllDomains()
			Convey("Then there should be one domain", func() {
				So(len(domains), ShouldEqual, 1)
			})

		})

		Convey("When it load one domain", func() {
			domain, _ = w.LoadDomain(domain.Name)
			Convey("Then it should be equals to created domain", func() {
				So(domain, ShouldNotBeNil)
				So(domain.Name, ShouldEqual, "test.domain.com")
				So(domain.Value, ShouldEqual, "testService")
				So(domain.Typ, ShouldEqual, "service")
			})
		})

		Convey("When it destroy a domain", func() {
			w.DestroyDomain(domain)
			domains, _ := w.LoadAllDomains()
			So(len(domains), ShouldEqual, 0)
		})

		Convey("When i modify a domain", func() {

			initialNotifCount := notifCount
			domain, _ := w.LoadDomain("test.domain.com")

			domain.Value = "testService2"

			w.PersistDomain(domain)

			Convey("Then the domain should be modified", func() {
				domain, _ := w.LoadDomain("test.domain.com")
				So(domain, ShouldNotBeNil)
				So(domain.Value, ShouldEqual, "testService2")
			})

			Convey("Then notifications	 should have been sent", func() {
				So(notifCount, ShouldBeGreaterThan, initialNotifCount)
			})

		})

		Convey("When two objet listens for event", func() {
			domain, _ := w.LoadDomain("test.domain.com")
			domain.Value = "testService2"

			var wg sync.WaitGroup
			wg.Add(2)
			go wait(w.Listen(), &wg)
			go wait(w.Listen(), &wg)

			w.PersistDomain(domain)

			Convey("Then both object should have been notified", func() {
				wg.Wait()
				So(true, ShouldBeTrue)
			})

		})
	})

}
