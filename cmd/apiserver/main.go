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

package main

import (
	"k8s.io/klog/v2"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"

	apprtv1alpha1 "github.com/kubevela/prism/pkg/apis/applicationresourcetracker/v1alpha1"
	clusterv1alpha1 "github.com/kubevela/prism/pkg/apis/cluster/v1alpha1"
	o11yconfig "github.com/kubevela/prism/pkg/apis/o11y/config"
	grafanav1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafana/v1alpha1"
	grafanadashboardv1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafanadashboard/v1alpha1"
	grafanadatasourcev1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafanadatasource/v1alpha1"
	apiserveroptions "github.com/kubevela/prism/pkg/util/apiserver/options"
	"github.com/kubevela/prism/pkg/util/log"
	"github.com/kubevela/prism/pkg/util/singleton"
)

func main() {
	cmd, err := builder.APIServer.
		WithLocalDebugExtension().
		ExposeLoopbackMasterClientConfig().
		ExposeLoopbackAuthorizer().
		WithoutEtcd().
		WithResource(&apprtv1alpha1.ApplicationResourceTracker{}).
		WithResource(&clusterv1alpha1.Cluster{}).
		WithResource(&grafanav1alpha1.Grafana{}).
		WithResource(&grafanadatasourcev1alpha1.GrafanaDatasource{}).
		WithResource(&grafanadashboardv1alpha1.GrafanaDashboard{}).
		WithConfigFns(apiserveroptions.WrapConfig).
		WithPostStartHook("init-master-loopback-client", singleton.InitLoopbackClient).
		Build()
	if err != nil {
		klog.Fatal(err)
	}
	log.AddLogFlags(cmd)
	apiserveroptions.AddServerRunFlags(cmd.Flags())
	clusterv1alpha1.AddClusterFlags(cmd.Flags())
	o11yconfig.AddObservabilityFlags(cmd.Flags())
	if err = cmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}
