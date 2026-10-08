/*
Copyright 2026 The KubeVela Authors.

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

package main

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	genericoptions "k8s.io/apiserver/pkg/server/options"
)

func TestDisableMutatingAdmissionPolicyByDefault(t *testing.T) {
	testCases := map[string]struct {
		args            []string
		expectedDisable []string
	}{
		"disabled by default": {
			expectedDisable: []string{mutatingAdmissionPolicyPlugin},
		},
		"explicitly enabled": {
			args: []string{"--enable-admission-plugins=" + mutatingAdmissionPolicyPlugin},
		},
		"explicitly disabled": {
			args:            []string{"--disable-admission-plugins=" + mutatingAdmissionPolicyPlugin},
			expectedDisable: []string{mutatingAdmissionPolicyPlugin},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			o := genericoptions.NewRecommendedOptions("", nil)
			flags := pflag.NewFlagSet(name, pflag.ContinueOnError)
			o.AddFlags(flags)
			require.NoError(t, flags.Parse(tc.args))
			disableMutatingAdmissionPolicyByDefault(o)
			require.Equal(t, tc.expectedDisable, o.Admission.DisablePlugins)
		})
	}

	t.Run("admission turned off", func(t *testing.T) {
		o := &genericoptions.RecommendedOptions{}
		disableMutatingAdmissionPolicyByDefault(o)
		require.Nil(t, o.Admission)
	})
}
