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
	"path"
	"sync"
	"time"

	"cuelang.org/go/pkg/strings"
	"github.com/emicklei/go-restful/v3"
	"golang.org/x/exp/slices"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	genericapi "k8s.io/apiserver/pkg/endpoints"
	"k8s.io/apiserver/pkg/registry/rest"
	"k8s.io/apiserver/pkg/server"
)

var DefaultDynamicAPIServer *DynamicAPIServer

type ResourceProvider interface {
	rest.Storage
	GetGroupVersion() schema.GroupVersion
	GetGroupVersionKind() schema.GroupVersionKind
	GetGroupVersionResource() schema.GroupVersionResource
}

type DynamicAPIServer struct {
	server *server.GenericAPIServer
	config *server.Config

	scheme               *runtime.Scheme
	parameterCodec       runtime.ParameterCodec
	negotiatedSerializer runtime.NegotiatedSerializer

	mu                      sync.Mutex
	apiGroups               map[string]*metav1.APIGroup
	apiGroupVersions        map[schema.GroupVersion]*genericapi.APIGroupVersion
	apiGroupVersionHandlers map[schema.GroupVersion]*restful.WebService
}

func NewDynamicAPIServer(svr *server.GenericAPIServer, config *server.Config) *DynamicAPIServer {
	s := &DynamicAPIServer{server: svr, config: config}
	s.scheme = runtime.NewScheme()
	s.parameterCodec = runtime.NewParameterCodec(s.scheme)
	s.negotiatedSerializer = serializer.NewCodecFactory(s.scheme)
	s.apiGroups = map[string]*metav1.APIGroup{}
	s.apiGroupVersions = map[schema.GroupVersion]*genericapi.APIGroupVersion{}
	metav1.AddToGroupVersion(s.scheme, metav1.Unversioned)
	return s
}

func (in *DynamicAPIServer) removeWebService(prefix string) {
	var toRemove *restful.WebService
	for _, svc := range in.server.Handler.GoRestfulContainer.RegisteredWebServices() {
		if svc.RootPath() == prefix {
			toRemove = svc
			break
		}
	}
	if toRemove != nil {
		_ = in.server.Handler.GoRestfulContainer.Remove(toRemove)
	}
}

func NewGroupVersionForDiscovery(gv schema.GroupVersion) metav1.GroupVersionForDiscovery {
	return metav1.GroupVersionForDiscovery{
		GroupVersion: gv.String(),
		Version:      gv.Version,
	}
}

func (in *DynamicAPIServer) AddGroupDiscovery(gv schema.GroupVersion) {
	in.mu.Lock()
	defer in.mu.Unlock()
	gv4discovery := NewGroupVersionForDiscovery(gv)
	apiGroup, exists := in.apiGroups[gv.Group]
	if !exists {
		apiGroup = &metav1.APIGroup{
			Name:             gv.Group,
			Versions:         []metav1.GroupVersionForDiscovery{},
			PreferredVersion: gv4discovery,
		}
	}
	if slices.Contains(apiGroup.Versions, gv4discovery) {
		return
	}
	apiGroup.Versions = append(apiGroup.Versions, gv4discovery)
	in.apiGroups[gv.Group] = apiGroup
	in.server.DiscoveryGroupManager.RemoveGroup(gv.Group)
	in.server.DiscoveryGroupManager.AddGroup(*apiGroup)
}

func (in *DynamicAPIServer) RemoveGroupDiscovery(gv schema.GroupVersion) {
	in.mu.Lock()
	defer in.mu.Unlock()
	apiGroup, exists := in.apiGroups[gv.Group]
	if !exists {
		return
	}
	gv4discovery := NewGroupVersionForDiscovery(gv)
	if idx := slices.Index(apiGroup.Versions, gv4discovery); idx >= 0 {
		apiGroup.Versions = slices.Delete(apiGroup.Versions, idx, idx)
	}
	if len(apiGroup.Versions) > 0 && apiGroup.PreferredVersion == gv4discovery {
		apiGroup.PreferredVersion = apiGroup.Versions[0]
	}
	if len(apiGroup.Versions) == 0 {
		in.server.DiscoveryGroupManager.RemoveGroup(gv.Group)
		delete(in.apiGroups, gv.Group)
		return
	}
	in.apiGroups[gv.Group] = apiGroup
	in.server.DiscoveryGroupManager.RemoveGroup(gv.Group)
	in.server.DiscoveryGroupManager.AddGroup(*apiGroup)
}

func (in *DynamicAPIServer) AddGroupVersionResourceHandler(gvr schema.GroupVersionResource, storage rest.Storage) error {
	in.mu.Lock()
	defer in.mu.Unlock()
	gv := gvr.GroupVersion()
	apiGroupVersion, exists := in.apiGroupVersions[gv]
	if !exists {
		apiGroupVersion = &genericapi.APIGroupVersion{
			Root:    server.APIGroupPrefix,
			Storage: map[string]rest.Storage{},

			GroupVersion:     gv,
			MetaGroupVersion: nil,

			ParameterCodec:        in.parameterCodec,
			Serializer:            in.negotiatedSerializer,
			Creater:               in.scheme,
			Convertor:             in.scheme,
			ConvertabilityChecker: in.scheme,
			UnsafeConvertor:       runtime.UnsafeObjectConvertor(in.scheme),
			Defaulter:             in.scheme,
			Typer:                 in.scheme,
			Namer:                 runtime.Namer(meta.NewAccessor()),

			EquivalentResourceRegistry: in.server.EquivalentResourceRegistry,

			Admit:               in.config.AdmissionControl,
			MinRequestTimeout:   time.Duration(in.config.MinRequestTimeout) * time.Second,
			MaxRequestBodyBytes: in.config.MaxRequestBodyBytes,
			Authorizer:          in.server.Authorizer,
		}
	}
	apiGroupVersion.Storage[gvr.Resource] = storage
	in.apiGroupVersions[gv] = apiGroupVersion
	in.removeGroupVersionHandler(gv)
	_, err := apiGroupVersion.InstallREST(in.server.Handler.GoRestfulContainer)
	return err
}

func (in *DynamicAPIServer) RemoveGroupVersionResourceHandler(gvr schema.GroupVersionResource, storage rest.Storage) error {
	in.mu.Lock()
	defer in.mu.Unlock()
	gv := gvr.GroupVersion()
	apiGroupVersion, exists := in.apiGroupVersions[gv]
	if !exists {
		return nil
	}
	delete(apiGroupVersion.Storage, gvr.Resource)
	in.removeGroupVersionHandler(gv)
	if len(apiGroupVersion.Storage) == 0 {
		delete(in.apiGroupVersions, gv)
		return nil
	}
	_, err := apiGroupVersion.InstallREST(in.server.Handler.GoRestfulContainer)
	return err
}

func (in *DynamicAPIServer) removeGroupVersionHandler(gv schema.GroupVersion) {
	prefix := path.Join(server.APIGroupPrefix, gv.Group, gv.Version)
	webservices := in.server.Handler.GoRestfulContainer.RegisteredWebServices()
	if idx := slices.IndexFunc(webservices, func(ws *restful.WebService) bool {
		return ws.RootPath() == prefix
	}); idx >= 0 {
		_ = in.server.Handler.GoRestfulContainer.Remove(webservices[idx])
	}
}

func (in *DynamicAPIServer) AddResource(r ResourceProvider) error {
	in.AddScheme(r.GetGroupVersionKind(), r)
	in.AddGroupDiscovery(r.GetGroupVersion())
	return in.AddGroupVersionResourceHandler(r.GetGroupVersionResource(), r)
}

func (in *DynamicAPIServer) RemoveResource(r ResourceProvider) error {
	in.RemoveScheme(r.GetGroupVersionKind())
	in.RemoveGroupDiscovery(r.GetGroupVersion())
	return in.RemoveGroupVersionResourceHandler(r.GetGroupVersionResource(), r)
}

func (in *DynamicAPIServer) AddScheme(gvk schema.GroupVersionKind, storage rest.Storage) {
	in.mu.Lock()
	defer in.mu.Unlock()
	in.scheme.AddKnownTypeWithName(gvk, storage.New())
	metav1.AddToGroupVersion(in.scheme, gvk.GroupVersion())
	if listStorage, ok := storage.(rest.Lister); ok {
		in.scheme.AddKnownTypeWithName(gvk.GroupVersion().WithKind(gvk.Kind+"List"), listStorage.NewList())
	}
}

func (in *DynamicAPIServer) RemoveScheme(gvk schema.GroupVersionKind) {
	in.mu.Lock()
	defer in.mu.Unlock()
	newScheme := runtime.NewScheme()
	for _gvk := range in.scheme.AllKnownTypes() {
		_gvk.Kind = strings.TrimSuffix(_gvk.Kind, "List")
		if _gvk == gvk {
			continue
		}
		obj, err := in.scheme.New(_gvk)
		if err != nil {
			continue
		}
		newScheme.AddKnownTypeWithName(_gvk, obj)
		if _gvk.Version != runtime.APIVersionInternal {
			metav1.AddToGroupVersion(newScheme, _gvk.GroupVersion())
		}
	}
	in.scheme = newScheme
	in.parameterCodec = runtime.NewParameterCodec(in.scheme)
	in.negotiatedSerializer = serializer.NewCodecFactory(in.scheme)
}
