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

const (
	STARTING_STATUS   = "starting"
	STARTED_STATUS    = "started"
	STOPPING_STATUS   = "stopping"
	STOPPED_STATUS    = "stopped"
	ERROR_STATUS      = "error"
	WARNING_STATUS    = "warning"
	NA_STATUS         = "n/a"
	PASSIVATED_STATUS = "passivated"
)

// Represents the combined status of a service
type Status struct {
	// Is the service alive
	Alive string `json:"alive"`

	// Current status of the service
	Current string `json:"current"`

	// Expected status of the service
	Expected string `json:"expected"`

	Service *Service `json:"-"`
}

// Creates a Status for a service with an initial value
func NewInitialStatus(initialStatus string, service *Service) *Status {
	return &Status{
		Current:  initialStatus,
		Expected: initialStatus,
		Service:  service,
	}
}

// Represents an error for a given status.
type StatusError struct {
	ComputedStatus string
	Status         *Status
}

func (s StatusError) Error() string {
	return s.ComputedStatus
}

func (s *Status) Equals(other *Status) bool {
	if s == nil && other == nil {
		return true
	}
	return s != nil && other != nil && s.Alive == other.Alive &&
		s.Current == other.Current &&
		s.Expected == other.Expected
}

// Computes the real status of the service made by the combination
// of current and expected state.
func (s *Status) Compute() string {

	if s != nil {
		Alive := s.Alive
		Expected := s.Expected
		Current := s.Current
		switch Current {
		case STOPPED_STATUS:
			if Expected == PASSIVATED_STATUS {
				return PASSIVATED_STATUS
			} else if Expected == STOPPED_STATUS {
				return STOPPED_STATUS
			} else {
				return ERROR_STATUS
			}
		case PASSIVATED_STATUS:
			if Expected == PASSIVATED_STATUS {
				return PASSIVATED_STATUS
			} else {
				return WARNING_STATUS
			}
		case STARTING_STATUS:
			if Expected == STARTED_STATUS {
				return STARTING_STATUS
			} else {
				return ERROR_STATUS
			}
		case STARTED_STATUS:
			if Alive != "" {
				if Expected != STARTED_STATUS {
					return WARNING_STATUS
				}
				return STARTED_STATUS
			} else {
				if Expected != STARTED_STATUS {
					return WARNING_STATUS
				}
				return ERROR_STATUS
			}
		case STOPPING_STATUS:
			if Expected == STOPPED_STATUS {
				return STOPPED_STATUS
			} else if Expected == PASSIVATED_STATUS {
				return PASSIVATED_STATUS
			} else {
				return ERROR_STATUS
			}
			// N/A
		default:
			return ERROR_STATUS
		}
	}
	return NA_STATUS
}
