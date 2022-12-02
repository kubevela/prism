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

package engine_test

import (
	"context"
	"fmt"
	"testing"

	"cuelang.org/go/cue"
	"github.com/stretchr/testify/require"

	"github.com/kubevela/prism/pkg/cue/engine"
	"github.com/kubevela/prism/pkg/cue/providers"
)

func TestCompile(t *testing.T) {
	x := `
	import "vela/http"
	x: http.#Do & {
		url: "https://api64.ipify.org/"
		method: "GET"
	}`
	v, err := engine.Compile(context.Background(), x)
	require.NoError(t, err)
	i, err := v.LookupPath(cue.ParsePath("x.response.statusCode")).Int64()
	require.NoError(t, err)
	require.Equal(t, int64(200), i)

	x = `
	x: {
		#do: "unknown"
		#provider: "none"
	}
	`
	v, err = engine.Compile(context.Background(), x)
	require.ErrorContains(t, err, "provider none not found")

	x = `
	import "vela/http"
	x: {
		#do: "unknown"
		#provider: "http"
	}
	`
	v, err = engine.Compile(context.Background(), x)
	require.ErrorContains(t, err, "function unknown not found in provider http")
}

func TestCompileWithoutProviders(t *testing.T) {
	x := `
	x: "src/"
	y: z: x + "test"`
	v, err := engine.Compile(context.Background(), x, engine.WithoutProviders{})
	require.NoError(t, err)
	i, err := v.LookupPath(cue.ParsePath("y.z")).String()
	require.NoError(t, err)
	require.Equal(t, "src/test", i)
}

type TestParams struct {
	Key string `json:"key"`
}

type TestReturns struct {
	Value string `json:"value"`
}

func test(ctx context.Context, params *TestParams) (*TestReturns, error) {
	if params.Key == "" {
		return nil, fmt.Errorf("empty key")
	}
	return &TestReturns{Value: "key: " + params.Key}, nil
}

func TestCompileWithCustomizedProvider(t *testing.T) {
	prds := engine.DefaultProviders.DeepCopy()
	pkg, err := providers.NewPackage(
		"custom",
		`
			package custom
			#Test: {
				#do: "test"
				#provider: "custom"
				key: string
				value?: string
			}
		`,
		map[string]providers.ProviderFn{
			"test": providers.GenericProviderFn[TestParams, TestReturns](test),
		})
	require.NoError(t, err)
	prds.Register(pkg)
	x := `
	import "vela/custom"
	x: y: custom.#Test & {
		key: "k"
	}`
	v, err := engine.Compile(context.Background(), x, engine.WithProviders{Providers: prds})
	require.NoError(t, err)
	i, err := v.LookupPath(cue.ParsePath("x.y.value")).String()
	require.NoError(t, err)
	require.Equal(t, "key: k", i)

	x = `
	import "vela/custom"
	x: y: custom.#Test & {
		key: ""
	}`
	_, err = engine.Compile(context.Background(), x, engine.WithProviders{Providers: prds})
	require.ErrorContains(t, err, "empty key")
}
