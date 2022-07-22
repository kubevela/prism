/*
Copyright 2022 The KubeVela Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package subresource

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	// CompoundNameSeparator concatenate child resource name and parent resource name
	CompoundNameSeparator = "@"
	// DefaultParentResourceName indicates the default parent resource name to use when not explicitly specified
	DefaultParentResourceName = "default"
)

// CompoundName the combination for resource name
type CompoundName struct {
	ParentResourceName string
	SubResourceName    string
}

// String .
func (in *CompoundName) String() string {
	return fmt.Sprintf("%s%s%s", in.SubResourceName, CompoundNameSeparator, in.ParentResourceName)
}

// NewCompoundName decode names into parent resource part and subresource part
func NewCompoundName(name string) *CompoundName {
	if !strings.Contains(name, CompoundNameSeparator) {
		return &CompoundName{ParentResourceName: DefaultParentResourceName, SubResourceName: name}
	}
	parts := strings.SplitN(name, CompoundNameSeparator, 2)
	return &CompoundName{ParentResourceName: parts[1], SubResourceName: parts[0]}
}

// GetParentResourceNameFromLabelSelector retrieve parent resource key from label selector
func GetParentResourceNameFromLabelSelector(sel labels.Selector, parentResourceKey string) string {
	requirements, _ := sel.Requirements()
	for _, r := range requirements {
		if r.Key() == parentResourceKey {
			if r.Operator() == selection.Equals && len(r.Values().List()) == 1 {
				return r.Values().List()[0]
			}
		}
	}
	return DefaultParentResourceName
}
