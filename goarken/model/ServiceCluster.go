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
	"errors"
	"sync"
)

// A Service cluster holds a list of Service instances.
// It provides a Next() method to get the next available
// service with a round robin policy.
type ServiceCluster struct {
	Name      string     `json:"name"`
	Instances []*Service `json:"instances"`
	lastIndex int
	lock      sync.RWMutex
}

// Create a new Service cluster
func NewServiceCluster(name string) *ServiceCluster {
	sc := &ServiceCluster{
		Name: name,
	}
	return sc
}

// Returns the next available service instance for that cluster.
// This may return a StatusError if no service is available.
func (cl *ServiceCluster) Next() (*Service, error) {
	if cl == nil {
		return nil, StatusError{}
	}
	cl.lock.RLock()
	defer cl.lock.RUnlock()
	if len(cl.Instances) == 0 {
		return nil, errors.New("no alive instance found")
	}
	var instance *Service
	for tries := 0; tries < len(cl.Instances); tries++ {
		index := (cl.lastIndex + 1) % len(cl.Instances)
		cl.lastIndex = index

		instance = cl.Instances[index]
		log.Debugf("Checking instance %d Status : %s", index, instance.Status.Compute())
		if instance.Status.Compute() == STARTED_STATUS && instance.Location.IsFullyDefined() {
			return instance, nil
		}

	}

	lastStatus := instance.Status

	if lastStatus == nil && !instance.Location.IsFullyDefined() {
		// Generates too much garbage
		log.Debugf("No Status and no location for %s", instance.Name)
		return nil, StatusError{ERROR_STATUS, lastStatus}
	}

	log.Debugf("No instance started for %s", instance.Name)
	if lastStatus != nil {
		log.Debugf("Last Status :")
		log.Debugf("   current  : %s", lastStatus.Current)
		log.Debugf("   expected : %s", lastStatus.Expected)
		log.Debugf("   alive : %s", lastStatus.Alive)
	} else {
		log.Debugf("No status available")
	}

	return nil, StatusError{instance.Status.Compute(), lastStatus}
}

// Removes an instance in the cluster.
func (cl *ServiceCluster) Remove(instanceIndex string) {

	match := -1
	for k, v := range cl.Instances {
		if v.Index == instanceIndex {
			match = k
		}
	}

	cl.Instances = append(cl.Instances[:match], cl.Instances[match+1:]...)
	cl.Dump("remove")
}

// Get an service by its key (index). Returns nil if not found.
func (cl *ServiceCluster) Get(instanceIndex string) *Service {
	for i, v := range cl.Instances {
		if v.Index == instanceIndex {
			return cl.Instances[i]
		}
	}
	return nil
}

// Adds a service to the cluster
func (cl *ServiceCluster) Add(service *Service) {

	for index, v := range cl.Instances {
		if v.Index == service.Index {
			cl.Instances[index] = service
			return
		}
	}

	cl.Instances = append(cl.Instances, service)
}

// Dump all service description to logs.
func (cl *ServiceCluster) Dump(action string) {
	for _, v := range cl.Instances {
		log.Debugf("Dump after %s %s -> %s:%d", action, v.Index, v.Location.Host, v.Location.Port)
	}
}

// Return the list of the cluster's instances.
func (cl *ServiceCluster) GetInstances() []*Service {
	return cl.Instances
}
