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

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGrafanaDatasourceFromResponseBody(t *testing.T) {
	in := &GrafanaDatasource{}
	require.NotNil(t, in.FromResponseBody([]byte(`bad`)))
	require.NoError(t, in.FromResponseBody([]byte(`{"datasource":{"key":"val"}}`)))
	require.Equal(t, []byte(`{"key":"val"}`), in.Spec.Raw)
}

func TestGrafanaDatasourceListFromResponseBody(t *testing.T) {
	in := &GrafanaDatasourceList{}
	require.NotNil(t, in.FromResponseBody([]byte(`bad`), "test"))
	require.Errorf(t, in.FromResponseBody([]byte(`[{}]`), "test"), "invalid grafana datasource response, no valid uid found")
	require.NoError(t, in.FromResponseBody([]byte(`[{"uid":"a","key":"A"},{"uid":"b","key":"B"}]`), "test"))
	require.Equal(t, 2, len(in.Items))
	require.Equal(t, "a@test", in.Items[0].GetName())
	require.Equal(t, []byte(`{"key":"A","uid":"a"}`), in.Items[0].Spec.Raw)
}
