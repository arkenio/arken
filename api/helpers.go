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
package api
import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/arkenio/goarken"
	"github.com/spf13/viper"
)



func CreateEtcdClient() *etcd.Client {
	etcdAdress := viper.GetString("etcdAddress")
	return etcd.NewClient([]string{etcdAdress})
}


func CreateWatcherFromCli(client *etcd.Client) *goarken.Watcher {
	domainDir := viper.GetString("domainDir")
	serviceDir := viper.GetString("serviceDir")
	w := &goarken.Watcher{
		Client:        client,
		DomainPrefix:  domainDir,
		ServicePrefix: serviceDir,
		Domains:       make(map[string]*goarken.Domain),
		Services:      make(map[string]*goarken.ServiceCluster),
	}
	w.Init()
	return w
}

