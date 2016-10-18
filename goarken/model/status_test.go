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

func Test_status(t *testing.T) {
	var status *Status

	Convey("Given a status", t, func() {

		status = &Status{}

		Convey("When started equals expected equals current", func() {
			status.Expected = "started"
			status.Current = "started"

			Convey("The computed status should be started if alive", func() {
				status.Alive = "1"
				So(status.Compute(), ShouldEqual, STARTED_STATUS)
			})

			Convey("The computed status should be error if not alive", func() {
				status.Alive = ""
				So(status.Compute(), ShouldEqual, ERROR_STATUS)
			})

		})

		Convey("When no status is expected and current is started", func() {
			status.Expected = ""
			status.Current = "started"

			Convey("The computed status should be warning if alive", func() {
				status.Alive = "1"
				So(status.Compute(), ShouldEqual, WARNING_STATUS)
			})

			Convey("The computed status should be warning if not alive", func() {
				status.Alive = ""
				So(status.Compute(), ShouldEqual, WARNING_STATUS)
			})

		})

		Convey("When started is expected and current is starting", func() {
			status.Expected = "started"
			status.Current = "starting"

			Convey("Then computed status should be starting", func() {
				So(status.Compute(), ShouldEqual, STARTING_STATUS)

			})

		})

		Convey("When stopped is expected and current is stopped", func() {
			status.Expected = "stopped"
			status.Current = "stopped"
			Convey("Then computed status should be stopped", func() {

				So(status.Compute(), ShouldEqual, STOPPED_STATUS)

			})

		})

		Convey("When stopped is expected and current is stopping", func() {
			status.Expected = "stopped"
			status.Current = "stopping"

			Convey("Then computed status should be starting", func() {

				So(status.Compute(), ShouldEqual, STOPPED_STATUS)

			})

		})

		Convey("When current is stopped", func() {
			status.Expected = "passivated"
			status.Current = "stopped"

			Convey("Then computed status should be passivated", func() {

				So(status.Compute(), ShouldEqual, PASSIVATED_STATUS)

			})
		})

		Convey("When current is stopping", func() {
			status.Expected = "passivated"
			status.Current = "stopping"

			Convey("Then computed status should also be passivated", func() {

				So(status.Compute(), ShouldEqual, PASSIVATED_STATUS)

			})

		})

		Convey("When status is nil", func() {
			var status *Status
			status = nil

			Convey("Then the computed status should be NA", func() {
				So(status.Compute(), ShouldEqual, NA_STATUS)
			})

		})

	})

}
