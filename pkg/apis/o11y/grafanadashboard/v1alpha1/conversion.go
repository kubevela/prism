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
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubevela/prism/pkg/util/subresource"
)

const (
	grafanaDashboardFolderIdLabelKey  = "o11y.oam.dev/grafana-dashboard-folder-id"
	grafanaDashboardFolderUidLabelKey = "o11y.oam.dev/grafana-dashboard-folder-uid"
)

// ToRequestBody convert object into body for request
func (in *GrafanaDashboard) ToRequestBody() ([]byte, error) {
	dashboard := map[string]interface{}{}
	if err := json.Unmarshal(in.Spec.Raw, &dashboard); err != nil {
		return nil, err
	}
	dashboard["uid"] = subresource.NewCompoundName(in.GetName()).SubResourceName
	data := map[string]interface{}{"dashboard": dashboard}
	if labels := in.GetLabels(); labels != nil {
		if raw := labels[grafanaDashboardFolderIdLabelKey]; raw != "" {
			id, err := strconv.Atoi(raw)
			if err != nil {
				return nil, err
			}
			data["folderId"] = id
		}
		if raw := labels[grafanaDashboardFolderUidLabelKey]; raw != "" {
			data["folderUid"] = raw
		}
	}
	return json.Marshal(data)
}

// FromResponseBody convert response into object
func (in *GrafanaDashboard) FromResponseBody(body []byte) error {
	data := map[string]interface{}{}
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	dashboard, ok := data["dashboard"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no dashboard found in response body")
	}
	delete(dashboard, "uid")
	meta, _ := data["meta"].(map[string]interface{})
	return in.load(dashboard, meta)
}

// FromResponseBody convert response into objects
func (in *GrafanaDashboardList) FromResponseBody(body []byte, parentResourceName string) error {
	var data []map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	in.Items = make([]GrafanaDashboard, len(data))
	for idx, raw := range data {
		gdb := &GrafanaDashboard{}
		uid, ok := raw["uid"].(string)
		if !ok {
			return fmt.Errorf("invalid dashboard response, no valid uid found")
		}
		gdb.SetName((&subresource.CompoundName{ParentResourceName: parentResourceName, SubResourceName: uid}).String())
		dashboard := map[string]interface{}{}
		for _, key := range []string{"title", "id", "tags"} {
			if raw[key] != nil {
				dashboard[key] = raw[key]
			}
		}
		if err := gdb.load(dashboard, raw); err != nil {
			return err
		}
		in.Items[idx] = *gdb
	}
	return nil
}

func (in *GrafanaDashboard) load(dashboard map[string]interface{}, meta map[string]interface{}) error {
	if meta != nil {
		labels := in.GetLabels()
		if labels == nil {
			labels = map[string]string{}
		}
		if id, validId := meta["folderId"].(int); validId {
			labels[grafanaDashboardFolderIdLabelKey] = strconv.Itoa(id)
		}
		if uid, validUid := meta["folderUid"].(string); validUid {
			labels[grafanaDashboardFolderUidLabelKey] = uid
		}
		in.SetLabels(labels)
	}
	bs, err := json.Marshal(dashboard)
	if err != nil {
		return err
	}
	in.Spec = runtime.RawExtension{Raw: bs}
	return nil
}
