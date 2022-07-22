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

package apiserver

import (
	"encoding/json"
	"fmt"
	"strings"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/utils/pointer"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetMetadataNameInFieldSelectorFromInternalVersionListOptions if fieldSelector is set in list options and metadata.name is specified
// return the name, otherwise nil
func GetMetadataNameInFieldSelectorFromInternalVersionListOptions(options *metainternalversion.ListOptions) *string {
	if options != nil && options.FieldSelector != nil && !options.FieldSelector.Empty() {
		if name, found := options.FieldSelector.RequiresExactMatch("metadata.name"); found {
			return pointer.String(name)
		}
	}
	return nil
}

// NewMatchingLabelSelectorFromInternalVersionListOptions create MatchingLabelsSelector from InternalVersion ListOptions
func NewMatchingLabelSelectorFromInternalVersionListOptions(options *metainternalversion.ListOptions) client.MatchingLabelsSelector {
	sel := labels.NewSelector()
	if options != nil && options.LabelSelector != nil && !options.LabelSelector.Empty() {
		sel = options.LabelSelector
	}
	return client.MatchingLabelsSelector{Selector: sel}
}

// NewListOptions create ListOptions from ListOption
func NewListOptions(options ...client.ListOption) *client.ListOptions {
	opts := &client.ListOptions{}
	for _, opt := range options {
		opt.ApplyToList(opts)
	}
	return opts
}

// BuildQueryParamsFromLabelSelector build query params from label selector with provided keys
func BuildQueryParamsFromLabelSelector(sel labels.Selector, keys ...string) (params string) {
	requirements, _ := sel.Requirements()
	for _, r := range requirements {
		if slices.Contains(keys, r.Key()) {
			if r.Operator() == selection.Equals || r.Operator() == selection.In {
				params += fmt.Sprintf("&%s=%s", r.Key(), strings.Join(r.Values().List(), ","))
			}
		}
	}
	return params
}

// GetStringFromRawExtension load string from raw extension
func GetStringFromRawExtension(data *runtime.RawExtension, path ...string) (val string) {
	if data != nil && data.Raw != nil {
		m := map[string]interface{}{}
		_ = json.Unmarshal(data.Raw, &m)
		val, _, _ = unstructured.NestedString(m, path...)
	}
	return val
}
