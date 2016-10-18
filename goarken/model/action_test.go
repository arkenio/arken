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
	"reflect"
	"testing"
)

func Test_actions(t *testing.T) {

	Convey("Given a service", t, func() {

		service := &Service{}
		service.Init()
		actions := make([]string, 0)

		Convey("When the service is starting ", func() {
			service.Status.Expected = "started"
			service.Status.Current = "strating"

			Convey("The returned list of actions should be empty", func() {
				service.Actions = GetActions(service)
				So(len(service.Actions.([]string)), ShouldEqual, 0)
			})

		})

		Convey("When the service is stopped ", func() {
			service.Status.Expected = "stopped"
			service.Status.Current = "stopped"

			Convey("The returned list of actions should be: start, delete, update", func() {
				service.Actions = GetActions(service)
				actions = append(actions, START_ACTION, DELETE_ACTION, UPDATE_ACTION)
				So(reflect.DeepEqual(service.Actions, actions), ShouldEqual, true)
			})

		})

		Convey("When the service is strated", func() {
			service.Status.Expected = "started"
			service.Status.Current = "started"
			actions = make([]string, 0)
			//simultate starting by adding the same actions as on start
			AddAction(service, STOP_ACTION, UPDATE_ACTION, DELETE_ACTION)

			Convey("Then the returned list of actions should be stop, update, delete", func() {

				service.Actions = GetActions(service)
				actions = append(actions, DELETE_ACTION, UPDATE_ACTION, STOP_ACTION)
				So(reflect.DeepEqual(service.Actions, actions), ShouldEqual, true)

			})

		})

		Convey("When the service is stopping", func() {
			service.Status.Expected = "stopped"
			service.Status.Current = "stopping"

			Convey("Then the returned list of actions should be empty", func() {

				service.Actions = GetActions(service)
				So(len(service.Actions.([]string)), ShouldEqual, 0)

			})

		})

		Convey("After the service has been updated and started", func() {
			service.Status.Expected = "started"
			service.Status.Current = "started"
			AddAction(service, STOP_ACTION) //as my service is started

			actions = make([]string, 0)
			//simultate the update by adding the upgrade action
			AddAction(service, UPGRADE_ACTION)

			Convey("Then the returned actions should be stop, upgrade, delete", func() {

				service.Actions = GetActions(service)
				actions = append(actions, DELETE_ACTION, STOP_ACTION, UPGRADE_ACTION)
				So(reflect.DeepEqual(service.Actions, actions), ShouldEqual, true)

			})
		})

		Convey("After the service has been upgraded", func() {
			service.Status.Expected = "started"
			service.Status.Current = "started"
			AddAction(service, DELETE_ACTION, STOP_ACTION, UPGRADE_ACTION)
			actions = make([]string, 0)

			//simultate the upgarde by adding the finish upgrade and rollbackactions
			AddAction(service, FINISHUPGRADE_ACTION, ROLLBACK_ACTION)
			Convey("Then the returned actions should be stop, finishupgrade and rollback", func() {
				service.Actions = GetActions(service)
				actions = append(actions, DELETE_ACTION, STOP_ACTION, FINISHUPGRADE_ACTION, ROLLBACK_ACTION)
				So(reflect.DeepEqual(service.Actions, actions), ShouldEqual, true)

			})

		})

		Convey("After the service has been upgraded and finish upgrade or rollback invoked", func() {
			service.Status.Expected = "started"
			service.Status.Current = "started"
			AddAction(service, DELETE_ACTION, STOP_ACTION, UPGRADE_ACTION)
			actions = make([]string, 0)

			//simultate the finish upgarde by adding back the Update action
			AddAction(service, UPDATE_ACTION)
			Convey("Then the returned actions should be delete, stop and update", func() {

				service.Actions = GetActions(service)
				actions = append(actions, DELETE_ACTION, STOP_ACTION, UPDATE_ACTION)
				So(reflect.DeepEqual(service.Actions, actions), ShouldEqual, true)

			})

		})

	})

}
