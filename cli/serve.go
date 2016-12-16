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
	"github.com/spf13/cobra"
	"github.com/arkenio/arken/api"
	"github.com/arkenio/arken/passivation"
	"github.com/spf13/viper"
	"os"
	"github.com/arkenio/arken/goarken/model"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts the Arken dameon",
	Run: func(cmd *cobra.Command, args []string) {

		log.Info("Starting Arken daemon...")
		// If a config file is found, read it in.
		if err := viper.ReadInConfig(); err == nil {
			log.Infof("Using config file: %s", viper.ConfigFileUsed())

		} else {
			log.Errorf("Unable to read config file")
		}


		// Initialize GoArken model
		etcdClient := CreateEtcdClient()

		serviceDriver, err := CreateServiceDriver(etcdClient)
		if(err != nil) {
			log.Error("Unable to create Service Driver :")
			log.Error(err.Error())
			os.Exit(-1)
		}

		persistenceDriver := CreateWatcherFromCli(etcdClient)

		arkenModel, err :=  model.NewArkenModel(serviceDriver, persistenceDriver )
		if err != nil {
			log.Error("Unable to initialize Arken model:")
			log.Error(err.Error())
			os.Exit(-1)
		}


		go passivation.NewHandler(arkenModel).Start()
		api.NewAPIServer(arkenModel).Start()


	},
}

func init() {
	RootCmd.AddCommand(serveCmd)

	serveCmd.Flags().Int("port", 8888, "Port to run Application server on")
	viper.BindPFlag("port", serveCmd.Flags().Lookup("port"))

}
