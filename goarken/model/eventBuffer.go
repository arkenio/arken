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
	"sort"
	"time"
)

// Simple struct that accumulate Model events in a map
// in order to have only one event for a given model name.
type eventBuffer struct {
	eventsMap      map[string]*ModelEvent
	events         chan *ModelEvent
	eventBroadcast *Broadcaster
}

// Creates a new eventBuffer. Buffer has to be started by
// the run() method in order to periodically unbeffer events.
func newEventBuffer(b *Broadcaster) *eventBuffer {
	return &eventBuffer{
		eventsMap:      make(map[string]*ModelEvent),
		events:         make(chan *ModelEvent),
		eventBroadcast: b,
	}
}

// Starts the eventBuffer by accepting new Model event
// and unbuffering at each duration
func (eb *eventBuffer) run(duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			events := make([]*ModelEvent, 0, len(eb.eventsMap))
			for _, v := range eb.eventsMap {
				events = append(events, v)
			}
			sort.Sort(ModelByTime(events))

			for _, event := range events {
				eb.eventBroadcast.Write(event)
				delete(eb.eventsMap, eb.keyFromModelEvent(event))
			}
			break
		case event := <-eb.events:
			eb.eventsMap[eb.keyFromModelEvent(event)] = event
			break
		}
	}
}

// Return a unique key given a ModelEvent base on the type of the model,
// the type of the event and the Model name.
func (eb *eventBuffer) keyFromModelEvent(event *ModelEvent) string {
	if sc, ok := event.Model.(*Service); ok {
		return fmt.Sprintf("SC_%s_%s", event.EventType, sc.Name)
	} else if domain, ok := event.Model.(*Domain); ok {
		return fmt.Sprintf("D_%s_%s", event.EventType, domain.Name)
	} else if service, ok := event.Model.(*Service); ok {
		return fmt.Sprintf("S_%s_%s", event.EventType, service.Name)
	}
	return "unknown"
}
