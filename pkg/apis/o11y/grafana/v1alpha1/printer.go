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
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// GrafanaCredentialType defines the credential type for grafana
type GrafanaCredentialType string

const (
	// GrafanaCredentialTypeNotAvailable not available
	GrafanaCredentialTypeNotAvailable GrafanaCredentialType = "NA"
	// GrafanaCredentialTypeBasicAuth basic auth
	GrafanaCredentialTypeBasicAuth GrafanaCredentialType = "BasicAuth"
	// GrafanaCredentialTypeBearerToken bearer token
	GrafanaCredentialTypeBearerToken GrafanaCredentialType = "BearerToken"
)

// GetCredentialType .
func (in *Grafana) GetCredentialType() GrafanaCredentialType {
	switch {
	case in.Spec.Access.Token != nil:
		return GrafanaCredentialTypeBearerToken
	case in.Spec.Access.BasicAuth != nil:
		return GrafanaCredentialTypeBasicAuth
	default:
		return GrafanaCredentialTypeNotAvailable
	}
}

// ConvertToTable convert resource to table
func (in *Grafana) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	switch obj := object.(type) {
	case *Grafana:
		return printGrafana(obj), nil
	case *GrafanaList:
		return printGrafanaList(obj), nil
	default:
		return nil, fmt.Errorf("unknown type %T", object)
	}
}

var (
	definitions = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: "the name of the Grafana"},
		{Name: "Endpoint", Type: "string", Description: "the endpoint"},
		{Name: "Credential_Type", Type: "string", Description: "the credential type"},
		{Name: "Labels", Type: "string", Description: "the labels of the Grafana", Priority: 10},
		{Name: "Creation_Timestamp", Type: "dateTime", Description: "the creation timestamp of the Grafana", Priority: 10},
	}
)

func printGrafana(in *Grafana) *metav1.Table {
	return &metav1.Table{
		ColumnDefinitions: definitions,
		Rows:              []metav1.TableRow{printGrafanaRow(in)},
	}
}

func printGrafanaList(in *GrafanaList) *metav1.Table {
	t := &metav1.Table{
		ColumnDefinitions: definitions,
	}
	for _, c := range in.Items {
		t.Rows = append(t.Rows, printGrafanaRow(c.DeepCopy()))
	}
	return t
}

func printGrafanaRow(c *Grafana) metav1.TableRow {
	var labels []string
	for k, v := range c.GetLabels() {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: c},
	}
	row.Cells = append(row.Cells,
		c.Name,
		c.Spec.Endpoint,
		c.GetCredentialType(),
		strings.Join(labels, ","),
		c.GetCreationTimestamp())
	return row
}
