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

package v1alpha1

import "fmt"

type invalidGrafanaSecretNameError struct{}

func (e invalidGrafanaSecretNameError) Error() string {
	return fmt.Sprintf("secret is not a valid grafana secret, name should has prefix %s", grafanaSecretNamePrefix)
}

// NewInvalidGrafanaSecretNameError create an invalid grafana secret error due to invalid name
func NewInvalidGrafanaSecretNameError() error {
	return invalidGrafanaSecretNameError{}
}

type emptyEndpointGrafanaSecretError struct{}

func (e emptyEndpointGrafanaSecretError) Error() string {
	return fmt.Sprintf("secret is not a valid grafana secret, no endpoint (%s) found in annotation", grafanaSecretEndpointAnnotationKey)
}

// NewEmptyEndpointGrafanaSecretError create an invalid grafana secret error due to no endpoint found
func NewEmptyEndpointGrafanaSecretError() error {
	return emptyEndpointGrafanaSecretError{}
}

type emptyCredentialGrafanaSecretError struct{}

func (e emptyCredentialGrafanaSecretError) Error() string {
	return fmt.Sprintf("secret is not a valid grafana secret, no credential found (token or username/password should be set)")
}

// NewEmptyCredentialGrafanaSecretError create an invalid grafana secret error due to no credential found
func NewEmptyCredentialGrafanaSecretError() error {
	return emptyCredentialGrafanaSecretError{}
}
