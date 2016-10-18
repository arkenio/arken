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
	"net/url"
	"strings"
)

const (
	START_ACTION         = "start"
	STOP_ACTION          = "stop"
	DELETE_ACTION        = "delete"
	UPDATE_ACTION        = "update"
	UPGRADE_ACTION       = "upgrade"
	FINISHUPGRADE_ACTION = "finishupgrade"
	ROLLBACK_ACTION      = "rollback"
)

//represents the action as returned in the service
type PrettyAction struct {
	//the name of the action
	Name string `json:"name"`
	//is a POST/PUT /DELETE
	Method string `json:"method"`
	//the url to invoke to execute this action on the service
	Url string `json:"url"`
}

//compute default actions based on the status of the service
//avoids to return a list of actions when the service is in starting or stopping status
//doesn't modify the list of actions on the service, and returns a simple string array as the actions are persisted
func GetActions(s *Service) []string {
	if s.Status != nil {
		actions := make([]string, 0)
		Current := s.Status.Current
		Expected := s.Status.Expected
		switch Current {
		case STOPPED_STATUS:
			if s.Actions != nil && len(s.Actions.([]string)) > 0 {
				return s.Actions.([]string)
			}
			if Expected == PASSIVATED_STATUS {
				actions = append(actions, START_ACTION, DELETE_ACTION, UPDATE_ACTION)
				return actions
			} else if Expected == STOPPED_STATUS {
				actions = append(actions, START_ACTION, UPDATE_ACTION)
			} else {
				return actions
			}
		case PASSIVATED_STATUS:
			if s.Actions != nil && len(s.Actions.([]string)) > 0 {
				return s.Actions.([]string)
			}
			if Expected == PASSIVATED_STATUS {
				actions = append(actions, START_ACTION, DELETE_ACTION)
			} else {
				return actions
			}
		case STARTING_STATUS:
			return actions
		case STARTED_STATUS:
			if s.Actions != nil && len(s.Actions.([]string)) > 0 {
				return s.Actions.([]string)
			}
			actions = append(actions, DELETE_ACTION, UPDATE_ACTION, STOP_ACTION)
		case STOPPING_STATUS:
			return actions
		default:
			return actions
		}
		return actions
	}
	return s.Actions.([]string)
}

//returns an array of actions with more details ( method, url)
func GetPrettyActions(s *Service, url *url.URL) []PrettyAction {
	actions := GetActions(s)
	prettyActions := make([]PrettyAction, 0)
	for _, a := range actions {
		prettyAction := PrettyAction{}
		prettyAction.Name = a
		prettyAction.Url = url.RequestURI()
		switch a {
		case START_ACTION:
			prettyAction.Method = "POST"
		case STOP_ACTION:
			prettyAction.Method = "POST"
		case DELETE_ACTION:
			prettyAction.Method = "DELETE"
		case UPDATE_ACTION:
			prettyAction.Method = "PUT"
		case UPGRADE_ACTION:
			prettyAction.Method = "POST"
		case FINISHUPGRADE_ACTION:
			prettyAction.Method = "POST"
		case ROLLBACK_ACTION:
			prettyAction.Method = "POST"
		}

		i := strings.Index(prettyAction.Url, "?action=")
		if i > 0 {
			prettyAction.Url = prettyAction.Url[:i]
		}

		if prettyAction.Method == "POST" {
			prettyAction.Url = prettyAction.Url + "?action=" + a
		}
		prettyActions = append(prettyActions, prettyAction)
	}
	return prettyActions
}

// called on create service, the service is stopped
func InitActions(s *Service) []string {

	if s.Actions == nil {
		s.Actions = make([]string, 0)
	}
	s.Actions = append(s.Actions.([]string), START_ACTION, DELETE_ACTION, UPDATE_ACTION)
	return s.Actions.([]string)
}

func AddAction(s *Service, actions ...string) {
	if s.Actions == nil {
		s.Actions = make([]string, 0)
	} else {
		s.Actions = s.Actions.([]string)
	}
	for _, action := range actions {
		canAdd := true
		switch action {
		case START_ACTION:
			for i, a := range s.Actions.([]string) {
				if a == action {
					canAdd = false
				}
				if a == STOP_ACTION { //remove stop
					s.Actions = append(s.Actions.([]string)[:i], s.Actions.([]string)[i+1:]...)
				}
			}
		case STOP_ACTION:
			for i, a := range s.Actions.([]string) {
				if a == action {
					canAdd = false
				}
				if a == START_ACTION {
					s.Actions = append(s.Actions.([]string)[:i], s.Actions.([]string)[i+1:]...)
				}
			}
		case DELETE_ACTION:
			for _, a := range s.Actions.([]string) {
				if a == action {
					canAdd = false
				}
			}
		case UPDATE_ACTION:
			var actions = make([]string, 0)
			for _, a := range s.Actions.([]string) {
				if a != UPGRADE_ACTION && a != FINISHUPGRADE_ACTION && a != ROLLBACK_ACTION {
					actions = append(actions, a)
				}
				if a == action {
					canAdd = false
				}
			}
			s.Actions = actions
		case UPGRADE_ACTION:
			for i, a := range s.Actions.([]string) {
				if a == UPDATE_ACTION {
					s.Actions = append(s.Actions.([]string)[:i], s.Actions.([]string)[i+1:]...)
				}
				if a == action || a == FINISHUPGRADE_ACTION || a == ROLLBACK_ACTION {
					canAdd = false
				}
			}
		case FINISHUPGRADE_ACTION:
			for i, a := range s.Actions.([]string) {
				if a == action {
					canAdd = false
				}
				if a == UPGRADE_ACTION {
					s.Actions = append(s.Actions.([]string)[:i], s.Actions.([]string)[i+1:]...)
				}
			}
		case ROLLBACK_ACTION:
			for i, a := range s.Actions.([]string) {
				if a == action {
					canAdd = false
				}
				if a == UPGRADE_ACTION {
					s.Actions = append(s.Actions.([]string)[:i], s.Actions.([]string)[i+1:]...)
				}
			}
		}
		if canAdd {
			s.Actions = append(s.Actions.([]string), action)
		}
	}
}
