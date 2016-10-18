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
	"errors"
	"fmt"
	. "github.com/arkenio/arken/goarken/model"
	"os"
	"os/exec"
	"strings"
)

type FleetServiceDriver struct {
	etcdAddress string
}

func NewFleetServiceDriver(etcdAddress string) *FleetServiceDriver {
	return &FleetServiceDriver{etcdAddress}
}

func (f *FleetServiceDriver) Create(s *Service, startOnCreate bool) (interface{}, error) {
	return nil, errors.New("Not implemented")
}

func (f *FleetServiceDriver) Start(s *Service) (interface{}, error) {
	err := f.fleetcmd(s, "start")
	return s, err
}

func (f *FleetServiceDriver) Stop(s *Service) (interface{}, error) {
	err := f.fleetcmd(s, "stop")
	return s, err
}


func (f *FleetServiceDriver) Upgrade(s *Service) (interface{}, error) {
	err := f.fleetcmd(s, "stop")
	if(err != nil) {
		return nil, err;
	}
	err = f.fleetcmd(s, "start")
	return s, err
}

func (f *FleetServiceDriver) FinishUpgrade(s *Service) (interface{}, error) {
	return nil, errors.New("Not implemented")
}

func (f *FleetServiceDriver) Rollback(s *Service) (interface{}, error) {
	return nil, errors.New("Not implemented")
}

func (f *FleetServiceDriver) Passivate(s *Service) (interface{}, error) {
	log.Info(fmt.Sprintf("Passivating service %s", s.Name))
	err := f.fleetcmd(s, "destroy")
	if err != nil {
		return s, err
	}

	//TODO make it at the model level
	//statusKey := s.NodeKey + "/status"
	//responseCurrent, error := f.client.Set(statusKey+"/current", PASSIVATED_STATUS, 0)
	//if error != nil && responseCurrent == nil {
	//	log.Errorf("Setting status current to 'passivated' has failed for Service "+s.Name+": %s", error)
	//}

	//response, error := f.client.Set(statusKey+"/expected", PASSIVATED_STATUS, 0)
	//if error != nil && response == nil {
	//	log.Errorf("Setting status expected to 'passivated' has failed for Service "+s.Name+": %s", error)
	//}
	return s, nil
}

func (f *FleetServiceDriver) Destroy(s *Service) error {
	err := f.fleetcmd(s, "destroy")
	return err
}

func unitNameFromService(s *Service) string {
	return "nxio@" + strings.Split(s.Name, "_")[1] + ".service"
}

func (f *FleetServiceDriver) fleetcmd(s *Service, command string) error {
	//TODO Use fleet's REST API

	cmd := exec.Command("/usr/bin/fleetctl", "--endpoint="+f.etcdAddress, command, unitNameFromService(s))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func UnitName(s *Service) string {
	return "nxio@" + strings.Split(s.Name, "_")[1] + ".service"
}

func (f *FleetServiceDriver) Listen() chan *ModelEvent {
	return nil
}

func (f *FleetServiceDriver) GetInfo(s *Service) (interface{}, error) {
	return &FleetInfoType{UnitName: unitNameFromService(s)}, nil
}

func (r *FleetServiceDriver) NeedToBeUpgraded(s *Service) (bool, error) {
	return false, nil
}
