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
package cli
import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/arkenio/goarken/model"
	"github.com/arkenio/goarken/storage"
	"github.com/arkenio/goarken/drivers"
	"github.com/spf13/viper"
)



func CreateEtcdClient() *etcd.Client {
	etcdAdress := viper.GetString("etcdAddress")
	return etcd.NewClient([]string{etcdAdress})
}


func CreateServiceDriver(etcdClient *etcd.Client ) model.ServiceDriver {
	switch viper.GetString("driver") {
	case "rancher":
		sd, err := drivers.NewRancherServiceDriver(viper.GetString("rancher.host"),viper.GetString("rancher.accessKey"),viper.GetString("rancher.secretKey"))
		if err != nil {
			panic(err)
		}
		return sd;

	default:
		return drivers.NewFleetServiceDriver(etcdClient)
	}
}


func CreateWatcherFromCli(client *etcd.Client) *storage.Watcher {
	return storage.NewWatcher(client, viper.GetString("serviceDir"), viper.GetString("domainDir"))

}

