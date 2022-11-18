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

package dynamicresource

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type DynamicResourceList struct {
	UnsList *unstructured.UnstructuredList
	codec   Codec
}

var _ runtime.Object = &DynamicResourceList{}

func (in *DynamicResourceList) GetObjectKind() schema.ObjectKind {
	return in
}

func (in *DynamicResourceList) DeepCopyObject() runtime.Object {
	return &DynamicResourceList{
		UnsList: in.UnsList.DeepCopy(),
		codec:   in.codec,
	}
}

func (in *DynamicResourceList) SetGroupVersionKind(kind schema.GroupVersionKind) {}

func (in *DynamicResourceList) GroupVersionKind() schema.GroupVersionKind {
	return in.codec.Source().GroupVersionKindList()
}

func (in *DynamicResourceList) MarshalJSON() ([]byte, error) {
	return json.Marshal(in.UnsList)
}

func (in *DynamicResourceList) UnmarshalJSON(bs []byte) error {
	in.UnsList = &unstructured.UnstructuredList{}
	return in.UnsList.UnmarshalJSON(bs)
}
