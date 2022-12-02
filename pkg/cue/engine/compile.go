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

package engine

import (
	"context"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/parser"

	"github.com/kubevela/prism/pkg/cue/providers"
	"github.com/kubevela/prism/pkg/cue/providers/http"
)

type CompileOption interface {
	ApplyTo(*CompileConfig)
}

type CompileConfig struct {
	Imports []*build.Instance
	providers.Providers
}

var DefaultProviders = providers.NewProviders()

func init() {
	DefaultProviders.Register(http.Package)
}

func NewCompileConfig() *CompileConfig {
	return &CompileConfig{
		Imports:   []*build.Instance{},
		Providers: DefaultProviders,
	}
}

type WithImports []*build.Instance

func (op WithImports) ApplyTo(cfg *CompileConfig) {
	cfg.Imports = append(cfg.Imports, op...)
}

type WithoutProviders struct{}

func (op WithoutProviders) ApplyTo(cfg *CompileConfig) { cfg.Providers = nil }

type WithProviders struct {
	providers.Providers
}

func (op WithProviders) ApplyTo(cfg *CompileConfig) { cfg.Providers = op }

func Compile(ctx context.Context, src string, opts ...CompileOption) (cue.Value, error) {
	cfg := NewCompileConfig()
	for _, op := range opts {
		op.ApplyTo(cfg)
	}
	bi := build.NewContext().NewInstance("", nil)
	bi.Imports = cfg.Imports
	if cfg.Providers != nil {
		bi.Imports = append(bi.Imports, cfg.Providers.Imports()...)
	}
	f, err := parser.ParseFile("-", src)
	if err != nil {
		return cue.Value{}, err
	}
	if err = bi.AddSyntax(f); err != nil {
		return cue.Value{}, err
	}
	val := cuecontext.New().BuildInstance(bi)
	if cfg.Providers != nil {
		val, err = cfg.Providers.Complete(ctx, val)
	}
	return val, err
}
