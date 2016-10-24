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
	"time"
)

// That struct hold an event for a Model (Service, ServiceCluster or Domain)
type ModelEvent struct {
	EventType string
	ModelType string
	Model     interface{}
	Time      time.Time
}

// Creates a new ModelEvent
func NewModelEvent(eventType string, model interface{}) *ModelEvent {
	return &ModelEvent{eventType, getModelType(model), model, time.Now()}
}

// Return the event ModelType
func getModelType(model interface{}) string {
	if _, ok := model.(*Domain); ok {
		return "Domain"
	} else if _, ok := model.(*Service); ok {
		return "Service"
	} else {
		return "Unknown"
	}
}

// Transform a generic interface channel to a *ModelEvent channel
func FromInterfaceChannel(fromChannel chan interface{}) chan *ModelEvent {
	result := make(chan *ModelEvent)
	go func() {
		for {
			event := <-fromChannel
			if evt, ok := event.(*ModelEvent); ok {
				result <- evt
			} else {
				panic(event)
			}

		}
	}()
	return result

}

// This type allow to sort ModelEvents by creation time.
type ModelByTime []*ModelEvent

func (m ModelByTime) Len() int {
	return len(m)
}

func (m ModelByTime) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m ModelByTime) Less(i, j int) bool {
	return m[i].Time.Before(m[j].Time)
}

func (me *ModelEvent) String() string {
	return fmt.Sprintf("%s on  [%s] %s", me.EventType, me.ModelType, me.Model)
}
