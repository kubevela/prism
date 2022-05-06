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
	"context"

	metainternalversion "k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/apiserver/pkg/registry/generic"
	apirest "k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	builderrest "sigs.k8s.io/apiserver-runtime/pkg/builder/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewResourceHandlerProvider resource handler provider for handling ApplicationResourceTracker requests
func NewResourceHandlerProvider(cfg *rest.Config) builderrest.ResourceHandlerProvider {
	return func(s *runtime.Scheme, g generic.RESTOptionsGetter) (apirest.Storage, error) {
		cli, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
		if err != nil {
			return nil, err
		}
		return &storage{
			cli:            cli,
			TableConvertor: apirest.NewDefaultTableConvertor(ApplicationResourceTrackerGroupResource),
		}, nil
	}
}

type storage struct {
	cli client.Client
	apirest.TableConvertor
}

// New returns an empty object that can be used with Create and Update after request data has been put into it.
func (s *storage) New() runtime.Object {
	return new(ApplicationResourceTracker)
}

// NamespaceScoped returns true if the storage is namespaced
func (s *storage) NamespaceScoped() bool {
	return true
}

// ShortNames delivers a list of short names for a resource.
func (s *storage) ShortNames() []string {
	return []string{"apprt"}
}

// Get finds a resource in the storage by name and returns it.
func (s *storage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	rt := &unstructured.Unstructured{}
	rt.SetGroupVersionKind(ResourceTrackerGroupVersionKind)
	ns := request.NamespaceValue(ctx)
	if err := s.cli.Get(ctx, types.NamespacedName{Name: name + "-" + ns}, rt); err != nil {
		return nil, err
	}
	return NewApplicationResourceTrackerFromResourceTracker(rt)
}

// NewList returns an empty object that can be used with the List call.
func (s *storage) NewList() runtime.Object {
	return &ApplicationResourceTrackerList{}
}

// List selects resources in the storage which match to the selector. 'options' can be nil.
func (s *storage) List(ctx context.Context, options *metainternalversion.ListOptions) (runtime.Object, error) {
	rts := &unstructured.UnstructuredList{}
	rts.SetGroupVersionKind(ResourceTrackerGroupVersionKind)
	ns := request.NamespaceValue(ctx)
	sel := matchingLabelsSelector{namespace: ns}
	if options != nil {
		sel.selector = options.LabelSelector
	}
	if err := s.cli.List(ctx, rts, sel); err != nil {
		return nil, err
	}
	appRts := &ApplicationResourceTrackerList{}
	for _, rt := range rts.Items {
		appRt, err := NewApplicationResourceTrackerFromResourceTracker(rt.DeepCopy())
		if err != nil {
			return nil, err
		}
		appRts.Items = append(appRts.Items, *appRt)
	}
	return appRts, nil
}

type matchingLabelsSelector struct {
	selector  labels.Selector
	namespace string
}

// ApplyToList applies this configuration to the given list options.
func (m matchingLabelsSelector) ApplyToList(opts *client.ListOptions) {
	opts.LabelSelector = m.selector
	if opts.LabelSelector == nil {
		opts.LabelSelector = labels.NewSelector()
	}
	if m.namespace != "" {
		sel := labels.SelectorFromValidatedSet(map[string]string{labelAppNamespace: m.namespace})
		r, _ := sel.Requirements()
		opts.LabelSelector = opts.LabelSelector.Add(r...)
	}
}
