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

package singleton_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kubevela/prism/pkg/util/singleton"
)

func TestSingleton(t *testing.T) {
	x := 1
	sgt := singleton.NewSingletonE(func() (int, error) {
		return x + 1, nil
	})
	require.Equal(t, 2, sgt.Get())
	sgt.Set(2)
	require.Equal(t, 2, sgt.Get())
	x = 3
	sgt.Reload()
	require.Equal(t, 4, sgt.Get())
}
