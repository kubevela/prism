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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/server"
)

func (in *DynamicResource) Install(s *server.GenericAPIServer) error {
	scheme := runtime.NewScheme()
	metav1.AddToGroupVersion(scheme, metav1.Unversioned)
	gv := in.GetGroupVersion()
	metav1.AddToGroupVersion(scheme, gv)
	scheme.AddKnownTypeWithName(in.codec.Source().GroupVersionKind(), in.New())
	scheme.AddKnownTypeWithName(in.codec.Source().GroupVersionKindList(), in.NewList())
	return s.InstallAPIGroups(&server.APIGroupInfo{
		Scheme:              scheme,
		PrioritizedVersions: []schema.GroupVersion{gv},
		VersionedResourcesStorageMap: map[string]map[string]rest.Storage{
			gv.Version: {
				in.codec.Source().Resource(): in,
			},
		},
		OptionsExternalVersion: &metav1.Unversioned,
		ParameterCodec:         runtime.NewParameterCodec(scheme),
		NegotiatedSerializer:   serializer.NewCodecFactory(scheme),
	})
}

func (in *DynamicResource) ResourceName() string {
	return in.codec.Source().Resource()
}
