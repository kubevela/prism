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
	"fmt"

	"cuelang.org/go/cue"

	"github.com/kubevela/prism/pkg/cue/util"
)

type ProviderNotFoundErr string

func (e ProviderNotFoundErr) Error() string {
	return fmt.Sprintf("provider %s not found", string(e))
}

type ProviderFnNotFoundErr struct {
	Provider, Fn string
}

func (e ProviderFnNotFoundErr) Error() string {
	return fmt.Sprintf("function %s not found in provider %s", e.Fn, e.Provider)
}

type ExecuteError struct {
	Path  string
	Value string
	Err   error
}

func (e ExecuteError) Error() string {
	return fmt.Sprintf("execute error for %s: %s (value: %s)", e.Path, e.Err.Error(), e.Value)
}

func NewExecuteError(v cue.Value, err error) ExecuteError {
	path := v.Path().String()
	s, e := util.ToString(v)
	if e != nil {
		s = e.Error()
	}
	return ExecuteError{Path: path, Value: s, Err: err}
}
