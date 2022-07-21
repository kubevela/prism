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

// ToBody convert object into body for request
func (in *GrafanaDashboard) ToBody() ([]byte, error) {
	dashboard := map[string]interface{}{}
	if err := json.Unmarshal(in.Spec.Raw, &dashboard); err != nil {
		return nil, err
	}
	dashboard["uid"] = subresource.NewCompoundName(in.GetName()).SubResourceName
	data := map[string]interface{}{"dashboard": dashboard}
	if labels := in.GetLabels(); labels != nil {
		if labels[grafanaDashboardFolderIdLabelKey] != "" {
			id, err := strconv.Atoi(labels[grafanaDashboardFolderIdLabelKey])
			if err != nil {
				return nil, err
			}
			data["folderId"] = id
		}
		if labels[grafanaDashboardFolderUidLabelKey] != "" {
			data["folderUid"] = labels[grafanaDashboardFolderUidLabelKey]
		}
	}
	return json.Marshal(data)
}

// FromBody convert response into object
func (in *GrafanaDashboard) FromBody(body []byte) error {
	data := map[string]interface{}{}
	if err := json.Unmarshal(body, &data); err != nil {
		return err
	}
	dashboard, ok := data["dashboard"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("no dashboard found in response body")
	}
	delete(dashboard, "uid")
	if meta, ok := data["meta"].(map[string]interface{}); ok {
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
