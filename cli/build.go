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
	"context"
	"github.com/arkenio/arken/build"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

// serveCmd represents the serve command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build an image for a stack",
	Run: func(cmd *cobra.Command, args []string) {

		if err := viper.ReadInConfig(); err == nil {
			log.Infof("Using config file: %s", viper.ConfigFileUsed())
		} else {
			log.Errorf("Unable to read config file")
		}

		params := make(map[string]string)

		params["nuxeoversion"] = viper.GetString("nuxeoversion")
		params["nuxeopackages"] = strings.Join(strings.Split(strings.TrimSuffix(strings.TrimPrefix(viper.GetStringSlice("nuxeopackages")[0], "["), "]"), ","), " ")

		log.Infof("param : %v", viper.GetStringSlice("nuxeopackages"))
		log.Infof("Buildin packages : %v", params["nuxeopackages"])
		desc := &build.BuildDescriptor{
			LocalTag: "myimage",
			Params:   params,
		}

		service := build.NewBuildService()
		service.Build(context.Background(), desc)

	},
}

func init() {
	RootCmd.AddCommand(buildCmd)

	buildCmd.Flags().String("nuxeoversion", "8.10", "Nuxeo version to use to build image")
	viper.BindPFlag("nuxeoversion", buildCmd.Flags().Lookup("nuxeoversion"))

	buildCmd.Flags().StringSlice("nuxeopackages", []string{"nuxeo-web-ui"}, "List of packages to install")
	viper.BindPFlag("nuxeopackages", buildCmd.Flags().Lookup("nuxeopackages"))
}
