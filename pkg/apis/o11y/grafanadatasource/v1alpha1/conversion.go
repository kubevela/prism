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

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubevela/prism/pkg/util/subresource"
)

// GetID get id from GrafanaDatasource
func (in *GrafanaDatasource) GetID() (int, error) {
	obj := struct {
		ID int `json:"id"`
	}{}
	return obj.ID, json.Unmarshal(in.Spec.Raw, &obj)
}

// FromResponseBody load datasource from grafana api create/update response
func (in *GrafanaDatasource) FromResponseBody(respBody []byte) error {
	obj := &struct {
		DataSource map[string]interface{} `json:"datasource"`
	}{}
	if err := json.Unmarshal(respBody, obj); err != nil {
		return err
	}
	bs, err := json.Marshal(obj.DataSource)
	if err != nil {
		return err
	}
	in.Spec = runtime.RawExtension{Raw: bs}
	return err
}

// FromResponseBody load datasources from grafana api
func (in *GrafanaDatasourceList) FromResponseBody(respBody []byte, parentResourceName string) error {
	data := []map[string]interface{}{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return err
	}
	in.Items = []GrafanaDatasource{}
	for _, raw := range data {
		ds := &GrafanaDatasource{}
		uid, ok := raw["uid"].(string)
		if !ok {
			return fmt.Errorf("invalid grafana datasource response, no valid uid found")
		}
		ds.SetName((&subresource.CompoundName{ParentResourceName: parentResourceName, SubResourceName: uid}).String())
		bs, err := json.Marshal(raw)
		if err != nil {
			return err
		}
		ds.Spec = runtime.RawExtension{Raw: bs}
		in.Items = append(in.Items, *ds)
	}
	return nil
}
