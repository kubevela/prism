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

package providers

import (
	"context"
	"encoding/json"
	"sync"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/build"

	"github.com/kubevela/prism/pkg/cue/util"
)

const (
	doKey       = "#do"
	providerKey = "#provider"
)

type ProviderFn interface {
	Call(context.Context, cue.Value) (cue.Value, error)
}

type GenericProviderFn[T any, U any] func(context.Context, *T) (*U, error)

func (fn GenericProviderFn[T, U]) Call(ctx context.Context, value cue.Value) (cue.Value, error) {
	params := new(T)
	bs, err := value.MarshalJSON()
	if err != nil {
		return value, err
	}
	if err = json.Unmarshal(bs, params); err != nil {
		return value, err
	}
	ret, err := fn(ctx, params)
	if err != nil {
		return value, err
	}
	return value.FillPath(cue.ParsePath(""), ret), nil
}

type Providers interface {
	GetProviderFn(provider string, fn string) (ProviderFn, error)
	Register(pkg *Package)
	Complete(ctx context.Context, value cue.Value) (cue.Value, error)
	Imports() []*build.Instance
	DeepCopy() Providers
}

func NewProviders() Providers {
	return &providers{
		m: map[string]*Package{},
	}
}

type providers struct {
	mu sync.RWMutex
	m  map[string]*Package
}

func (p *providers) GetProviderFn(provider string, fn string) (ProviderFn, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	prd, ok := p.m[provider]
	if !ok {
		return nil, ProviderNotFoundErr(provider)
	}
	f, ok := prd.Fns[fn]
	if !ok {
		return nil, ProviderFnNotFoundErr{Provider: provider, Fn: fn}
	}
	return f, nil
}

func (p *providers) Register(pkg *Package) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.m[pkg.Name] = pkg
}

func (p *providers) Complete(ctx context.Context, value cue.Value) (cue.Value, error) {
	newValue := value
	executed := map[string]bool{}
	for {
		var next *cue.Value
		// 1. find the next to execute
		util.Iterate(newValue, func(v cue.Value) (stop bool) {
			_, done := executed[v.Path().String()]
			fn, _ := v.LookupPath(cue.ParsePath(doKey)).String()
			if !done && fn != "" {
				next = &v
				return true
			}
			return false
		})
		if next == nil {
			break
		}
		// 2. execute
		fn, _ := next.LookupPath(cue.ParsePath(doKey)).String()
		prd, _ := next.LookupPath(cue.ParsePath(providerKey)).String()
		f, err := p.GetProviderFn(prd, fn)
		if err != nil {
			return newValue, err
		}
		val, err := f.Call(ctx, *next)
		if err != nil {
			return newValue, NewExecuteError(val, err)
		}
		newValue = newValue.FillPath(next.Path(), val)
		executed[next.Path().String()] = true
	}
	return newValue, nil
}

func (p *providers) Imports() []*build.Instance {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var imports []*build.Instance
	for _, v := range p.m {
		if v.Import != nil {
			imports = append(imports, v.Import)
		}
	}
	return imports
}

func (p *providers) DeepCopy() Providers {
	newp := &providers{m: map[string]*Package{}}
	p.mu.Lock()
	defer p.mu.Unlock()
	for k, v := range p.m {
		newp.m[k] = v
	}
	return newp
}
