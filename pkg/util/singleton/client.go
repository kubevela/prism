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

package singleton

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var kubeConfig *rest.Config
var kubeClient client.Client
var restMapper meta.RESTMapper
var staticClient kubernetes.Interface
var dynamicClient dynamic.Interface

// GetKubeConfig get kubernetes config
func GetKubeConfig() *rest.Config {
	return kubeConfig
}

// SetKubeConfig set kubernetes config
func SetKubeConfig(cfg *rest.Config) {
	kubeConfig = cfg
}

// GetKubeClient get kubernetes client
func GetKubeClient() client.Client {
	return kubeClient
}

// SetKubeClient set kubernetes client
func SetKubeClient(cli client.Client) {
	kubeClient = cli
}

// GetRESTMapper get rest mapper
func GetRESTMapper() meta.RESTMapper {
	return restMapper
}

// GetDynamicClient get dynamic client
func GetDynamicClient() dynamic.Interface {
	return dynamicClient
}

// GetStaticClient get static client
func GetStaticClient() kubernetes.Interface {
	return staticClient
}

var once = sync.Once{}

// InitClient init clients
func InitClient(ctx server.PostStartHookContext) (err error) {
	once.Do(func() {
		err = initClient(ctx)
	})
	return err
}

func initClient(ctx server.PostStartHookContext) (err error) {
	if kubeConfig, err = config.GetConfig(); err != nil {
		return err
	}
	if restMapper, err = apiutil.NewDiscoveryRESTMapper(kubeConfig); err != nil {
		return err
	}
	if kubeClient, err = client.New(kubeConfig, client.Options{
		Scheme: scheme.Scheme,
		Mapper: restMapper,
	}); err != nil {
		return err
	}
	if staticClient, err = kubernetes.NewForConfig(kubeConfig); err != nil {
		return err
	}
	if dynamicClient, err = dynamic.NewForConfig(kubeConfig); err != nil {
		return err
	}
	return nil
}
