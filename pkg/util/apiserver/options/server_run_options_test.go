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

package options

import (
	"testing"
	"time"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"k8s.io/apiserver/pkg/server"
)

func TestServerRunOptions(t *testing.T) {
	cfg := &server.RecommendedConfig{}
	AddServerRunFlags(pflag.CommandLine)
	defaultServerRunOptions.RequestTimeout = time.Second * 10
	WrapConfig(cfg)
	require.Equal(t, cfg.RequestTimeout, time.Second*10)
}
