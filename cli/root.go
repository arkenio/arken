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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)


type Config struct {
	Port          int
	DomainPrefix  string
	ServicePrefix string
	EtcdAddress   string
}


var cfgFile string
var config Config

// This represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "arken",
	Short: "The Arken cluster management daemon",
	Long: `This is the Arken cluster management daemon, that knows
how to create/start/stop your application environments.`,
// Uncomment the following line if your bare application
// has an action associated with it:
//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.arken.yaml)")
	RootCmd.PersistentFlags().StringVar(&config.DomainPrefix, "domainDir", "/domains", "etcd prefix to get domains")
	RootCmd.PersistentFlags().StringVar(&config.ServicePrefix, "serviceDir", "/services", "etcd prefix to get services")
	RootCmd.PersistentFlags().StringVar(&config.EtcdAddress, "etcdAddress", "http://127.0.0.1:4001/", "Address of an etcd endpoint")

	viper.BindPFlag("domainDir", RootCmd.PersistentFlags().Lookup("domainDir"))
	viper.BindPFlag("serviceDir", RootCmd.PersistentFlags().Lookup("serviceDir"))
	viper.BindPFlag("etcdAddress", RootCmd.PersistentFlags().Lookup("etcdAddress"))


}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".arken") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AutomaticEnv()          // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
