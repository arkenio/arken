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
	"errors"
	"fmt"
	"github.com/arkenio/arken/goarken/drivers"
	"github.com/arkenio/arken/goarken/model"
	"github.com/arkenio/arken/goarken/storage"
	"github.com/coreos/etcd/client"
	"github.com/spf13/viper"
	"time"
)

func CreateEtcdClient() client.KeysAPI {
	etcdAdress := viper.GetString("etcdAddress")

	cfg := client.Config{
		Endpoints:               []string{etcdAdress},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	return client.NewKeysAPI(c)

}

func CreateServiceDriver(etcdClient client.KeysAPI) (model.ServiceDriver, error) {

	rancherHost := viper.GetString("CATTLE_URL")
	if rancherHost == "" {
		rancherHost = viper.GetString("rancher.host") //compat with old catalog def
	}

	accessKey := viper.GetString("CATTLE_ACCESS_KEY")
	if accessKey == "" {
		accessKey = viper.GetString("rancher.accessKey") //compat with old catalog def
	}

	secretKey := viper.GetString("CATTLE_SECRET_KEY")
	if secretKey == "" {
		secretKey = viper.GetString("rancher.secretKey") //compat with old catalog def
	}

	switch viper.GetString("driver") {
	case "rancher":
		log.Infof("Rancher host: %s", rancherHost)
		log.Infof("Rancher Access key: %s", accessKey)
		log.Infof("Rancher Secret key: ************************")
		sd, err := drivers.NewRancherServiceDriver(rancherHost, accessKey, secretKey)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Unable to connect to Rancher : %s", err.Error()))
		}

		return sd, nil

	default:
		return drivers.NewFleetServiceDriver(viper.GetString("etcdAddress")), nil
	}
}

func CreateWatcherFromCli(client client.KeysAPI) *storage.Watcher {
	return storage.NewWatcher(client, viper.GetString("serviceDir"), viper.GetString("domainDir"))

}
