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
	"context"
	"net"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/endpoints/openapi"
	"k8s.io/apiserver/pkg/registry/rest"
	genericapiserver "k8s.io/apiserver/pkg/server"
	genericoptions "k8s.io/apiserver/pkg/server/options"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	basecompatibility "k8s.io/component-base/compatibility"
	baseversion "k8s.io/component-base/version"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	netutils "k8s.io/utils/net"

	cueserver "github.com/kubevela/pkg/cue/server"
	apiserveroptions "github.com/kubevela/pkg/util/apiserver/options"
	"github.com/kubevela/pkg/util/log"
	"github.com/kubevela/pkg/util/singleton"

	apprtv1alpha1 "github.com/kubevela/prism/pkg/apis/applicationresourcetracker/v1alpha1"
	clusterv1alpha1 "github.com/kubevela/prism/pkg/apis/cluster/v1alpha1"
	"github.com/kubevela/prism/pkg/apis/generated"
	o11yconfig "github.com/kubevela/prism/pkg/apis/o11y/config"
	grafanav1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafana/v1alpha1"
	grafanadashboardv1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafanadashboard/v1alpha1"
	grafanadatasourcev1alpha1 "github.com/kubevela/prism/pkg/apis/o11y/grafanadatasource/v1alpha1"
	apiserver "github.com/kubevela/prism/pkg/dynamicapiserver"
)

// standaloneDebugMode allows the apiserver to be run locally without
// authorization/admission for testing, mirroring what
// sigs.k8s.io/apiserver-runtime's WithLocalDebugExtension used to provide.
var standaloneDebugMode bool

func main() {
	cmd := newCommand()
	utilruntime.Must(cmd.Execute())
}

func newCommand() *cobra.Command {
	codecs := serializer.NewCodecFactory(clientgoscheme.Scheme)
	o := genericoptions.NewRecommendedOptions("", codecs.LegacyCodec(
		apprtv1alpha1.GroupVersion, clusterv1alpha1.GroupVersion, grafanav1alpha1.GroupVersion))

	cmd := &cobra.Command{
		Use:   "vela-prism",
		Short: "Launch the vela-prism aggregated apiserver",
		RunE: func(c *cobra.Command, args []string) error {
			return runServer(c.Context(), o)
		},
	}
	cmd.SetContext(genericapiserver.SetupSignalContext())

	flags := cmd.Flags()
	o.AddFlags(flags)
	utilfeature.DefaultMutableFeatureGate.AddFlag(flags)
	flags.BoolVar(&standaloneDebugMode, "standalone-debug-mode", false,
		"Under the local-debug mode the apiserver will allow all access to its resources without "+
			"authorizing the requests, this flag is only intended for debugging in your workstation "+
			"and the apiserver will be crashing if its binding address is not 127.0.0.1.")

	log.AddLogFlags(cmd)
	apiserveroptions.AddServerRunFlags(flags)
	clusterv1alpha1.AddClusterFlags(flags)
	o11yconfig.AddObservabilityFlags(flags)

	return cmd
}

func runServer(ctx context.Context, o *genericoptions.RecommendedOptions) error {
	// prism's resources are synthesized live from Secrets/ManagedClusters/the
	// Grafana HTTP API rather than persisted, so no etcd storage is needed.
	o.Etcd = nil
	o.Authentication.RemoteKubeConfigFileOptional = true
	if standaloneDebugMode {
		if o.SecureServing.BindAddress.String() != "127.0.0.1" {
			klog.Fatal(`--bind-address must be "127.0.0.1" if --standalone-debug-mode is set`)
		}
		o.Authorization = nil
		o.Admission = nil
	}

	if err := o.SecureServing.MaybeDefaultWithSelfSignedCerts(
		"localhost", nil, []net.IP{netutils.ParseIPSloppy("127.0.0.1")}); err != nil {
		return err
	}

	codecs := serializer.NewCodecFactory(clientgoscheme.Scheme)
	serverConfig := genericapiserver.NewRecommendedConfig(codecs)
	serverConfig.EffectiveVersion = basecompatibility.NewEffectiveVersionFromString(baseversion.DefaultKubeBinaryVersion, "", "")
	serverConfig = apiserveroptions.WrapConfig(serverConfig)

	if err := o.ApplyTo(serverConfig); err != nil {
		return err
	}

	namer := openapi.NewDefinitionNamer(clientgoscheme.Scheme)
	serverConfig.OpenAPIConfig = genericapiserver.DefaultOpenAPIConfig(generated.GetOpenAPIDefinitions, namer)
	serverConfig.OpenAPIConfig.Info.Title = "Vela Prism"
	serverConfig.OpenAPIConfig.Info.Version = "1.0.0"
	// cueserver.RegisterGenericAPIServer's /cue and /cuex webservices don't
	// follow REST verb-naming conventions, so exclude them from OpenAPI spec
	// generation rather than have the builder fail on their operation names.
	serverConfig.OpenAPIConfig.IgnorePrefixes = []string{"/cue", "/cuex"}
	serverConfig.OpenAPIV3Config = genericapiserver.DefaultOpenAPIV3Config(generated.GetOpenAPIDefinitions, namer)
	serverConfig.OpenAPIV3Config.Info.Title = "Vela Prism"
	serverConfig.OpenAPIV3Config.Info.Version = "1.0.0"
	serverConfig.OpenAPIV3Config.IgnorePrefixes = []string{"/cue", "/cuex"}

	genericServer, err := serverConfig.Complete().New("vela-prism", genericapiserver.NewEmptyDelegate())
	if err != nil {
		return err
	}
	genericServer = cueserver.RegisterGenericAPIServer(genericServer)
	singleton.InitGenericAPIServer(genericServer)
	singleton.InitServerConfig(serverConfig)

	parameterCodec := runtime.NewParameterCodec(clientgoscheme.Scheme)

	prismGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(apprtv1alpha1.Group, clientgoscheme.Scheme, parameterCodec, codecs)
	prismGroupInfo.VersionedResourcesStorageMap[apprtv1alpha1.Version] = map[string]rest.Storage{
		apprtv1alpha1.ApplicationResourceTrackerResource: &apprtv1alpha1.ApplicationResourceTracker{},
		clusterv1alpha1.ClusterResource:                  &clusterv1alpha1.Cluster{},
	}

	o11yGroupInfo := genericapiserver.NewDefaultAPIGroupInfo(grafanav1alpha1.Group, clientgoscheme.Scheme, parameterCodec, codecs)
	o11yGroupInfo.VersionedResourcesStorageMap[grafanav1alpha1.Version] = map[string]rest.Storage{
		grafanav1alpha1.GrafanaResource:                     &grafanav1alpha1.Grafana{},
		grafanadatasourcev1alpha1.GrafanaDatasourceResource: &grafanadatasourcev1alpha1.GrafanaDatasource{},
		grafanadashboardv1alpha1.GrafanaDashboardResource:   &grafanadashboardv1alpha1.GrafanaDashboard{},
	}

	if err := genericServer.InstallAPIGroups(&prismGroupInfo, &o11yGroupInfo); err != nil {
		return err
	}

	if err := genericServer.AddPostStartHook("start-dynamic-server", apiserver.StartDefaultDynamicAPIServer); err != nil {
		return err
	}

	return genericServer.PrepareRun().RunWithContext(ctx)
}
