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
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_cluster(t *testing.T) {
	var cluster *ServiceCluster

	Convey("Given an service cluster", t, func() {
		cluster = &ServiceCluster{}

		Convey("When the cluster is initialized", func() {
			Convey("Then it should be empty", func() {
				So(len(cluster.Instances), ShouldEqual, 0)

			})
		})

		Convey("When the cluster contains an inactive service", func() {
			cluster.Add(getService("1", "nxio-0001", false))
			Convey("Then it can't get a next service", func() {
				service, err := cluster.Next()

				So(len(cluster.Instances), ShouldEqual, 1)
				So(service, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When the cluster contains active service", func() {
			cluster.Add(getService("2", "nxio-0001", true))
			Convey("Then it can get a next service", func() {
				service, err := cluster.Next()

				So(len(cluster.Instances), ShouldEqual, 1)
				So(service, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})

			Convey("Then returned service should always be the same", func() {
				service, _ := cluster.Next()
				firstKey := service.Index
				service, _ = cluster.Next()
				So(service.Index, ShouldEqual, firstKey)

			})

		})

		Convey("When the cluster contains several services", func() {
			cluster.Add(getService("1", "nxio-0001", true))
			cluster.Add(getService("2", "nxio-0001", false))
			cluster.Add(getService("3", "nxio-0001", true))

			Convey("Then it should loadbalance between services", func() {
				service, err := cluster.Next()
				So(service, ShouldNotBeNil)
				So(err, ShouldBeNil)

				firstKey := service.Index

				service, err = cluster.Next()
				So(service, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(service.Index, ShouldNotEqual, firstKey)
			})

			Convey("Then it should never loadbalance on an inactive service", func() {
				for i := 0; i < len(cluster.Instances); i++ {
					service, err := cluster.Next()
					So(service, ShouldNotBeNil)
					So(err, ShouldBeNil)
					So(service.Index, ShouldNotEqual, "2")
				}
			})

			Convey("Then it can get each service by its key", func() {

				service := cluster.Get("1")
				So(service.Index, ShouldEqual, "1")
				So(service.Status.Current, ShouldEqual, "started")

				service = cluster.Get("2")
				So(service.Index, ShouldEqual, "2")
				So(service.Status.Current, ShouldEqual, "stopped")
			})

		})

		Convey("When removing a key to a cluster", func() {
			cluster.Add(getService("1", "nxio-0001", true))
			cluster.Add(getService("2", "nxio-0001", false))
			cluster.Add(getService("3", "nxio-0001", true))

			initSize := len(cluster.Instances)

			cluster.Remove("2")

			Convey("Then it should containe one less instance", func() {
				So(len(cluster.Instances), ShouldEqual, initSize-1)

			})
		})

	})

}

func Test_Service(t *testing.T) {
	var service1, service2 *Service

	Convey("Given two service with same values", t, func() {
		service1 = getService("1", "nxio-0001", true)
		service2 = getService("1", "nxio-0001", true)
		Convey("When i dont change anything", func() {
			Convey("Then they are equal", func() {

				So(service1.Equals(service2), ShouldEqual, true)

			})

		})

		Convey("When host is not the same", func() {
			service2.Location.Host = "otherhost"
			Convey("Then they are not equal", func() {

				So(service1.Equals(service2), ShouldEqual, false)

			})

		})

		Convey("When port is not the same", func() {
			service2.Location.Port = 9090
			Convey("Then they are not equal", func() {

				So(service1.Equals(service2), ShouldEqual, false)

			})

		})

		Convey("When current status is not the same", func() {
			service2.Status.Current = "other"
			Convey("Then they are not equal", func() {

				So(service1.Equals(service2), ShouldEqual, false)

			})

		})

		Convey("When expected status is not the same", func() {
			service2.Status.Expected = "other"
			Convey("Then they are not equal", func() {

				So(service1.Equals(service2), ShouldEqual, false)

			})

		})
		Convey("When alive status is not the same", func() {
			service2.Status.Alive = "other"
			Convey("Then they are not equal", func() {

				So(service1.Equals(service2), ShouldEqual, false)

			})

		})
	})
}

func getService(index string, name string, active bool) *Service {
	var s *Status

	if active {
		s = &Status{"1", "started", "started", &Service{}}
	} else {
		s = &Status{"", "stopped", "started", &Service{}}
	}

	return &Service{
		Index:    index,
		Location: &Location{"127.0.0.1", 8080},
		Domain:   "dummydomain.com",
		Name:     name,
		Status:   s,
	}

}
