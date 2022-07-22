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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGrafanaDashboardToRequestBody(t *testing.T) {
	in := &GrafanaDashboard{}
	in.SetName("test@local")
	in.Spec = runtime.RawExtension{Raw: []byte(`{"key":"val"}`)}
	in.SetLabels(map[string]string{
		grafanaDashboardFolderIdLabelKey:  "1",
		grafanaDashboardFolderUidLabelKey: "uid",
	})
	bs, err := in.ToRequestBody()
	require.NoError(t, err)
	var m1, m2 map[string]interface{}
	require.NoError(t, json.Unmarshal(bs, &m1))
	require.NoError(t, json.Unmarshal([]byte(`{"dashboard":{"uid":"test","key":"val"},"folderId":1,"folderUid":"uid"}`), &m2))
	require.Equal(t, m2, m1)
	// test bad label
	in.SetLabels(map[string]string{grafanaDashboardFolderIdLabelKey: "bad"})
	_, err = in.ToRequestBody()
	require.NotNil(t, err)
	// test bad spec
	in.Spec = runtime.RawExtension{Raw: []byte(`bad`)}
	in.SetLabels(nil)
	_, err = in.ToRequestBody()
	require.NotNil(t, err)
}

func TestGrafanaDashboardFromResponseBody(t *testing.T) {
	in := &GrafanaDashboard{}
	require.NotNil(t, in.FromResponseBody([]byte(`bad`)))
	require.Errorf(t, in.FromResponseBody([]byte(`{}`)), "no dashboard found in response body")
	// test full load
	require.NoError(t, in.FromResponseBody([]byte(`{"dashboard":{"uid":"test","key":"val"},"meta":{"folderId":1,"folderUid":"a"}}`)))
	require.Equal(t, []byte(`{"key":"val","uid":"test"}`), in.Spec.Raw)
	require.Equal(t, "1", in.GetLabels()[grafanaDashboardFolderIdLabelKey])
	require.Equal(t, "a", in.GetLabels()[grafanaDashboardFolderUidLabelKey])
}

func TestGrafanaDashboardListFromResponseBody(t *testing.T) {
	in := &GrafanaDashboardList{}
	require.NotNil(t, in.FromResponseBody([]byte(`[bad]`), "test"))
	require.Errorf(t, in.FromResponseBody([]byte(`[{}]`), "test"), "invalid dashboard response, no valid uid found")
	require.NoError(t, in.FromResponseBody([]byte(`[{"uid":"a","title":"A"},{"uid":"b","title":"B"}]`), "test"))
	require.Equal(t, len(in.Items), 2)
	require.Equal(t, []byte(`{"title":"A"}`), in.Items[0].Spec.Raw)
	require.Equal(t, "a@test", in.Items[0].Name)
}
