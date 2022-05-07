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
	"k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var kubeConfig *rest.Config
var kubeClient client.Client

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

// InitLoopbackClient init clients
func InitLoopbackClient(ctx server.PostStartHookContext) (err error) {
	if kubeConfig, err = config.GetConfig(); err != nil {
		return err
	}
	if kubeClient, err = client.New(kubeConfig, client.Options{Scheme: scheme.Scheme}); err != nil {
		return err
	}
	return nil
}
