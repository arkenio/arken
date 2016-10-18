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

// Simple structure that implements a publish/subscribe mecanism.
type Broadcaster struct {
	listeners []chan interface{}
}

func NewBroadcaster() *Broadcaster {
	b := &Broadcaster{
		listeners: []chan interface{}{},
	}
	return b
}

func (b *Broadcaster) Write(message interface{}) {
	for _, channel := range b.listeners {
		channel <- message
	}
}

func (b *Broadcaster) Listen() chan interface{} {
	channel := make(chan interface{})
	b.listeners = append(b.listeners, channel)
	return channel

}
