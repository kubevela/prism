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

package dynamicresource_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubevela/prism/pkg/apis/dynamicresource"
)

var _ = Describe("Test codec", func() {
	It("Test typer", func() {
		_, err := dynamicresource.NewDefaultTyper("a/b/c", "")
		Ω(err).NotTo(Succeed())
		typer, err := dynamicresource.NewDefaultTyper("test.oam.dev/v1beta1", "Tester")
		Ω(err).To(Succeed())
		gv := schema.GroupVersion{Group: "test.oam.dev", Version: "v1beta1"}
		Ω(typer.GroupVersion()).To(Equal(gv))
		Ω(typer.GroupVersionKind()).To(Equal(gv.WithKind("Tester")))
		Ω(typer.GroupVersionKindList()).To(Equal(gv.WithKind("TesterList")))
		Ω(typer.GroupVersionResource()).To(Equal(gv.WithResource("testers")))
		Ω(typer.Kind()).To(Equal("Tester"))
		Ω(typer.KindList()).To(Equal("TesterList"))
		Ω(typer.Resource()).To(Equal("testers"))
	})

	It("Test template codec", func() {
		By("Normal codec")
		codec, err := dynamicresource.NewTemplateCodec(encoderTemplate, decoderTemplate)
		Ω(err).To(Succeed())
		sourceGVK := schema.GroupVersionKind{Group: "test.oam.dev", Version: "v1alpha2", Kind: "Tester"}
		targetGVK := schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}
		Ω(codec.Source().GroupVersionKind()).To(Equal(sourceGVK))
		Ω(codec.Target().GroupVersionKind()).To(Equal(targetGVK))

		By("Normal codec good encode and decode")
		src := &unstructured.Unstructured{}
		src.SetGroupVersionKind(sourceGVK)
		tgt, err := codec.Encode(src)
		Ω(err).To(Succeed())
		Ω(tgt.GroupVersionKind()).To(Equal(targetGVK))
		tgt = &unstructured.Unstructured{}
		tgt.SetGroupVersionKind(targetGVK)
		src, err = codec.Decode(tgt)
		Ω(err).To(Succeed())
		Ω(src.GroupVersionKind()).To(Equal(sourceGVK))

		By("Normal codec bad encode")
		src = &unstructured.Unstructured{}
		src.SetGroupVersionKind(schema.GroupVersionKind{Group: "bad", Version: "unknown", Kind: "v0"})
		_, err = codec.Encode(src)
		Ω(err).NotTo(Succeed())

		By("Codec with bad encoder")
		_, err = dynamicresource.NewTemplateCodec(`bad-key: bad-val`, "")
		Ω(err).NotTo(Succeed())
		_, err = dynamicresource.NewTemplateCodec(`parameter: apiVersion: 1`, "")
		Ω(err).NotTo(Succeed())
		_, err = dynamicresource.NewTemplateCodec(`parameter: apiVersion: "v1"`, "")
		Ω(err).NotTo(Succeed())
		_, err = dynamicresource.NewTemplateCodec(encoderTemplate, "bad-good")
		Ω(err).NotTo(Succeed())
	})
})
