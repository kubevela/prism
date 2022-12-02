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

package server

import (
	"context"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"

	"github.com/kubevela/prism/pkg/cue/engine"
)

func Render(bs []byte, path ...string) ([]byte, error) {
	ctx := cuecontext.New()
	val := ctx.CompileBytes(bs)
	if err := val.Err(); err != nil {
		return nil, err
	}
	for _, p := range path {
		val = val.LookupPath(cue.ParsePath(p))
	}
	return val.MarshalJSON()
}

// Compile cue bytes and get the json output for the given path
func Compile(bs []byte, path ...string) ([]byte, error) {
	val, err := engine.Compile(context.Background(), string(bs))
	if err != nil {
		return nil, err
	}
	if err = val.Err(); err != nil {
		return nil, err
	}
	for _, p := range path {
		val = val.LookupPath(cue.ParsePath(p))
	}
	return val.MarshalJSON()
}
