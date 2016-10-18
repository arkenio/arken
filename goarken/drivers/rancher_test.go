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
package drivers

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_RancherDriver(t *testing.T) {

	Convey("Given a rancher host", t, func() {
		rancherHost := "http://192.168.99.106:8080/v1/projects/1a5"

		projectId := getProjectIdFromRancherHost(rancherHost)
		Convey("Then i can extract its project id", func() {
			So(projectId, ShouldEqual, "1a5")
		})
	})
}
