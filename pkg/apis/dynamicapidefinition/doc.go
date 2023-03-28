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

// Api versions allow the api contract for a resource to be changed while keeping
// backward compatibility by support multiple concurrent versions
// of the same resource

// Package v1alpha1 contains types required for v1alpha1
// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:defaulter-gen=TypeMeta
// +groupName=prism.oam.dev
package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// DynamicAPIDefinition .
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DynamicAPIDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec DynamicAPIDefinitionSpec `json:"spec,omitempty"`
}

// DynamicAPIDefinitionList list for DynamicAPIDefinition
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type DynamicAPIDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []DynamicAPIDefinition `json:"items"`
}

type DynamicAPIEngineType string

const (
	EngineTypeObjectCodec DynamicAPIEngineType = "object-codec"
	EngineTypeCueX        DynamicAPIEngineType = "cuex"
)

type DynamicAPIDefinitionSpec struct {
	Engine DynamicAPIEngine `json:"engine"`
}

type DynamicAPIEngine struct {
	Type      DynamicAPIEngineType `json:"type"`
	Templates map[string]string    `json:"templates,omitempty"`
}
