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
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// ResourceTrackerGroupVersionKind the backend gvk of the target resource
	ResourceTrackerGroupVersionKind = schema.GroupVersionKind{
		Group:   "core.oam.dev",
		Version: "v1beta1",
		Kind:    "ResourceTracker",
	}
)

const (
	labelAppNamespace = "app.oam.dev/namespace"
)

// NewApplicationResourceTrackerFromResourceTracker convert KubeVela ResourceTracker to ApplicationResourceTracker
func NewApplicationResourceTrackerFromResourceTracker(rt *unstructured.Unstructured) (*ApplicationResourceTracker, error) {
	appRt := &ApplicationResourceTracker{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(rt.Object, appRt); err != nil {
		return nil, err
	}
	namespace := metav1.NamespaceDefault
	if labels := rt.GetLabels(); labels != nil && labels[labelAppNamespace] != "" {
		namespace = labels[labelAppNamespace]
	}
	appRt.SetNamespace(namespace)
	appRt.SetName(strings.TrimSuffix(rt.GetName(), "-"+namespace))
	appRt.SetGroupVersionKind(ApplicationResourceTrackerGroupVersionKind)
	return appRt, nil
}
