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

package main

import (
	"flag"
	"github.com/arkenio/arken/cli"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"github.com/Sirupsen/logrus"
)

var log = logrus.New()

func main() {
	flag.Parse()
	handleSignals()
	cli.Execute()
}

func handleSignals() {
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	signal.Notify(signals, os.Interrupt, syscall.SIGUSR1)
	signal.Notify(signals, os.Interrupt, syscall.SIGUSR2)

	go func() {
		isProfiling := false

		defer func() {
			if isProfiling {
				pprof.StopCPUProfile()
			}
		}()

		for {
			sig := <-signals
			switch sig {
			case syscall.SIGTERM, syscall.SIGINT:
				//Exit gracefully
				log.Infof("Shutting down...")
				os.Exit(0)
			case syscall.SIGUSR1:
				pprof.Lookup("goroutine").WriteTo(os.Stdout, 2)
			case syscall.SIGUSR2:
				if !isProfiling {
					f, err := os.Create("/tmp/arken.profile")
					if err != nil {
						log.Fatal(err)
					} else {
						pprof.StartCPUProfile(f)
						isProfiling = true
					}
				} else {
					pprof.StopCPUProfile()
					isProfiling = false
				}

			}
		}

	}()
}
