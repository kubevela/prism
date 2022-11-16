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

	"k8s.io/apimachinery/pkg/util/runtime"
)

type Singleton[T any] struct {
	mu     sync.Mutex
	loaded bool

	loader func() T
	data   T
}

func (in *Singleton[T]) Get() T {
	in.mu.Lock()
	if !in.loaded && in.loader != nil {
		in.mu.Unlock()
		in.Set(in.loader())
		in.mu.Lock()
	}
	defer in.mu.Unlock()
	return in.data
}

func (in *Singleton[T]) Set(data T) {
	in.mu.Lock()
	defer in.mu.Unlock()
	in.loaded = true
	in.data = data
}

func (in *Singleton[T]) Reload() {
	if in.loader != nil {
		in.Set(in.loader())
	}
}

func NewSingleton[T any](loader func() T) *Singleton[T] {
	return &Singleton[T]{
		loader: loader,
	}
}

func NewSingletonE[T any](loaderE func() (T, error)) *Singleton[T] {
	loader := func() T {
		t, err := loaderE()
		runtime.Must(err)
		return t
	}
	return NewSingleton(loader)
}
