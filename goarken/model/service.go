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
	"fmt"
	"github.com/Sirupsen/logrus"
	"time"
)

type Location struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func (s *Location) Equals(other *Location) bool {
	if s == nil && other == nil {
		return true
	}

	return s != nil && other != nil &&
		s.Host == other.Host &&
		s.Port == other.Port
}

func (s *Location) IsFullyDefined() bool {
	if s == nil {
		return false
	}
	return s.Host != "" && s.Port != 0

}

func (s Location) String() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Holds configuration of the service
type ServiceConfig struct {
	Robots      string                 `json:"robots"`
	Environment map[string]interface{} `json:"environment,omitempty"`
	// Rancher backed service information
	RancherInfo *RancherInfoType `json:"rancherInfo,omitempty"`
	// Fleet backed service information
	FleetInfo   *FleetInfoType     `json:"fleetInfo,omitempty"`
	Passivation *PassivationConfig `json:"passivation,omitempty`
}

type RancherInfoType struct {
	EnvironmentId   string    `json:"environmentId,omitempty"`
	EnvironmentName string    `json:"environmentName,omitempty"`
	Location        *Location `json:"location,omitempty"`
	HealthState     string    `json:"healthState,omitempty"`
	CurrentStatus   string    `json:"currentStatus,omitempty"`
	TemplateId      string    `json:"templateId,omitempty"`
}

func (r RancherInfoType) String() string {
	return fmt.Sprintf("RancherInfo for %s : envId: %s, location: %s, currentStatus: %s, rancherHealth: %s", r.EnvironmentName, r.EnvironmentId, r.Location, r.CurrentStatus, r.HealthState)
}

type FleetInfoType struct {
	UnitName string
}

func (config *ServiceConfig) Equals(other *ServiceConfig) bool {
	if config == nil && other == nil {
		return true
	}

	return config != nil && other != nil &&
		config.Robots == other.Robots
}

// Holds information about a given service
type Service struct {
	Index      string         `json:"-"`
	NodeKey    string         `json:"-"`
	Location   *Location      `json:"location"`
	Domain     string         `json:"domain"`
	Name       string         `json:"name"`
	Status     *Status        `json:"status"`
	Actions    interface{}  `json:"actions"` 
	LastAccess *time.Time     `json:"lastAccess"`
	Config     *ServiceConfig `json:"config"`
	log        *logrus.Logger
}


func (s *Service) Init() *Service {

	s.Index = "1"

	status := NewInitialStatus(STOPPED_STATUS, s)
	s.Actions = InitActions(s)

	s.Status = status
	s.Config = &ServiceConfig{
		Passivation: DefaultPassivation(),
	}

	return s

}

func (service *Service) Equals(other *Service) bool {
	if service == nil && other == nil {
		return true
	}

	return service != nil && other != nil &&
		service.Location.Equals(other.Location) &&
		service.Status.Equals(other.Status) &&
		service.Config.Equals(other.Config)
}

func (s *Service) StartedSince() *time.Time {
	if s == nil {
		return nil
	}

	if s.Status != nil &&
		s.Status.Current == STARTED_STATUS {
		return s.LastAccess
	} else {
		return nil
	}
}
