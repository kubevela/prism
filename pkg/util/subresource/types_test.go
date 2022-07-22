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
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

func TestCompoundName(t *testing.T) {
	require.Equal(t, "test@default", NewCompoundName("test").String())
	require.Equal(t, "test@local", NewCompoundName("test@local").String())
}

func TestGetParentResourceNameFromLabelSelector(t *testing.T) {
	sel := labels.NewSelector()
	require.Equal(t, "default", GetParentResourceNameFromLabelSelector(sel, "key"))
	r, err := labels.NewRequirement("key", selection.Equals, []string{"val"})
	require.NoError(t, err)
	sel = sel.Add(*r)
	require.Equal(t, "val", GetParentResourceNameFromLabelSelector(sel, "key"))
}