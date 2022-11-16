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

package singleton_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apiserver/pkg/server"
	"sigs.k8s.io/apiserver-runtime/pkg/builder"

	"github.com/kubevela/prism/pkg/util/singleton"
	_ "github.com/kubevela/prism/test/bootstrap"
)

func TestSingletonInit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Test Singleton Init")
}

var _ = Describe("Test init", func() {
	It("Test clients", func() {
		singleton.RESTMapper.Get()
		singleton.KubeClient.Get()
		singleton.StaticClient.Get()
		singleton.DynamicClient.Get()
	})

	It("Test servers", func() {
		singleton.InitGenericAPIServer(&builder.GenericAPIServer{})
		Ω(singleton.GenericAPIServer.Get()).ToNot(BeNil())
		singleton.InitServerConfig(&server.RecommendedConfig{})
		Ω(singleton.APIServerConfig.Get()).ToNot(BeNil())
	})
})
