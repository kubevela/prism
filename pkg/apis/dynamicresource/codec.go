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
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/pkg/strings"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/kubevela/prism/pkg/util/singleton"
)

type Typer interface {
	GroupVersion() schema.GroupVersion
	GroupVersionKind() schema.GroupVersionKind
	GroupVersionKindList() schema.GroupVersionKind
	GroupVersionResource() schema.GroupVersionResource
	Kind() string
	KindList() string
	Resource() string
	Namespaced() bool
}

type Codec interface {
	Source() Typer
	Target() Typer
	Encode(src *unstructured.Unstructured) (*unstructured.Unstructured, error)
	Decode(src *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type typer struct {
	groupVersion schema.GroupVersion
	kind         string
	resource     string
	kindList     string
	namespaced   bool
}

var _ Typer = &typer{}

func (in *typer) GroupVersion() schema.GroupVersion {
	return in.groupVersion
}

func (in *typer) GroupVersionKind() schema.GroupVersionKind {
	return in.groupVersion.WithKind(in.kind)
}

func (in *typer) GroupVersionKindList() schema.GroupVersionKind {
	return in.groupVersion.WithKind(in.kindList)
}

func (in *typer) GroupVersionResource() schema.GroupVersionResource {
	return in.groupVersion.WithResource(in.resource)
}

func (in *typer) Kind() string {
	return in.kind
}

func (in *typer) KindList() string {
	return in.kindList
}

func (in *typer) Resource() string {
	return in.resource
}

func (in *typer) Namespaced() bool {
	return in.namespaced
}

func NewDefaultTyper(apiVersion string, kind string) (Typer, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return nil, err
	}
	resource := strings.ToLower(kind) + "s"
	namespaced := true
	mappings, err := singleton.GetRESTMapper().RESTMappings(gv.WithKind(kind).GroupKind(), gv.Version)
	if err == nil && len(mappings) > 0 {
		resource = mappings[0].Resource.Resource
		namespaced = mappings[0].Scope.Name() == meta.RESTScopeNameNamespace
	}
	return &typer{
		groupVersion: gv,
		kind:         kind,
		resource:     resource,
		kindList:     kind + "List",
		namespaced:   namespaced,
	}, nil
}

const (
	templateParameterKey = "parameter"
	templateOutputKey    = "output"
)

type templateCodec struct {
	source, target Typer
	encodeTemplate string
	decodeTemplate string
}

var _ Codec = &templateCodec{}

func (in *templateCodec) Source() Typer {
	return in.source
}

func (in *templateCodec) Target() Typer {
	return in.target
}

func (in *templateCodec) Encode(source *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return in.convert(source, in.encodeTemplate, in.target)
}

func (in *templateCodec) Decode(target *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return in.convert(target, in.decodeTemplate, in.source)
}

func (in *templateCodec) convert(src *unstructured.Unstructured, template string, dest Typer) (*unstructured.Unstructured, error) {
	bs, err := src.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if template != "" {
		ctx := cuecontext.New()
		param := ctx.CompileBytes(bs)
		val := ctx.CompileString(template).
			FillPath(cue.ParsePath(templateParameterKey), param).
			LookupPath(cue.ParsePath(templateOutputKey))
		if err = val.Err(); err != nil {
			return nil, err
		}
		if bs, err = val.MarshalJSON(); err != nil {
			return nil, err
		}
	}
	out := &unstructured.Unstructured{}
	if err = out.UnmarshalJSON(bs); err != nil {
		return nil, err
	}
	out.SetGroupVersionKind(dest.GroupVersionKind())
	return out, nil
}

func (in *templateCodec) loadTyper(template string, path string) (Typer, error) {
	templateVal := cuecontext.New().CompileString(template).Value()
	if err := templateVal.Err(); err != nil {
		return nil, err
	}
	apiVersion, err := templateVal.LookupPath(cue.ParsePath(path + ".apiVersion")).String()
	if err != nil {
		return nil, err
	}
	kind, err := templateVal.LookupPath(cue.ParsePath(path + ".kind")).String()
	if err != nil {
		return nil, err
	}
	return NewDefaultTyper(apiVersion, kind)
}

func NewTemplateCodec(encodeTemplate, decodeTemplate string) (Codec, error) {
	var err error
	codec := &templateCodec{
		encodeTemplate: encodeTemplate,
		decodeTemplate: decodeTemplate,
	}
	codec.source, err = codec.loadTyper(encodeTemplate, templateParameterKey)
	if err != nil {
		return nil, err
	}
	codec.target, err = codec.loadTyper(decodeTemplate, templateParameterKey)
	if err != nil {
		return nil, err
	}
	return codec, nil
}
