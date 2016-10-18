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

// A Domain in the Arken model is of a given and may point to a service (if type is service)
type Domain struct {
	NodeKey string `json:"-"`
	Name    string `json:"name,omitempty"`
	Typ     string `json:"type,omitempty"`
	Value   string `json:"value,omitempty"`
}

func (d *Domain) String() string {
	return d.Value + " at " + d.NodeKey
}

func (domain *Domain) Equals(other *Domain) bool {
	if domain == nil && other == nil {
		return true
	}

	return domain != nil && other != nil &&
		domain.Typ == other.Typ && domain.Value == other.Value
}
