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

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	"github.com/kubevela/prism/pkg/apis/o11y/config"
)

// ToSecret convert grafana instance to underlying secret
func (in *Grafana) ToSecret() *corev1.Secret {
	secret := &corev1.Secret{Data: map[string][]byte{}}
	secret.ObjectMeta = in.ObjectMeta
	secret.SetName(grafanaSecretNamePrefix + in.GetName())
	secret.SetNamespace(config.ObservabilityNamespace)
	secret.SetOwnerReferences(nil)
	annotations := in.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[grafanaSecretEndpointAnnotationKey] = in.Spec.Endpoint
	secret.SetAnnotations(annotations)
	if in.Spec.Access.Token != nil {
		secret.Data[grafanaSecretTokenKey] = []byte(*in.Spec.Access.Token)
	}
	if in.Spec.Access.BasicAuth != nil {
		secret.Data[grafanaSecretUsernameKey] = []byte(in.Spec.Access.Username)
		secret.Data[grafanaSecretPasswordKey] = []byte(in.Spec.Access.Password)
	}
	return secret
}

// NewGrafanaFromSecret create grafana from secret
func NewGrafanaFromSecret(secret *corev1.Secret) (*Grafana, error) {
	secret = secret.DeepCopy()
	grafana := &Grafana{}
	grafana.ObjectMeta = secret.ObjectMeta
	if !strings.HasPrefix(secret.GetName(), grafanaSecretNamePrefix) {
		return nil, NewInvalidGrafanaSecretNameError()
	}
	grafana.SetName(strings.TrimPrefix(secret.GetName(), grafanaSecretNamePrefix))
	grafana.SetNamespace("")
	if annotations := secret.GetAnnotations(); annotations != nil {
		grafana.Spec.Endpoint = strings.TrimSpace(annotations[grafanaSecretEndpointAnnotationKey])
		delete(annotations, grafanaSecretEndpointAnnotationKey)
		grafana.SetAnnotations(annotations)
	}
	if grafana.Spec.Endpoint == "" {
		return nil, NewEmptyEndpointGrafanaSecretError()
	}
	if secret.Data[grafanaSecretTokenKey] != nil {
		grafana.Spec.Access.Token = pointer.String(string(secret.Data[grafanaSecretTokenKey]))
	}
	if secret.Data[grafanaSecretUsernameKey] != nil && secret.Data[grafanaSecretPasswordKey] != nil {
		grafana.Spec.Access.BasicAuth = &BasicAuth{
			Username: string(secret.Data[grafanaSecretUsernameKey]),
			Password: string(secret.Data[grafanaSecretPasswordKey]),
		}
	}
	if grafana.Spec.Access.BasicAuth == nil && grafana.Spec.Access.Token == nil {
		return nil, NewEmptyCredentialGrafanaSecretError()
	}
	return grafana, nil
}
