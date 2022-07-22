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

// ConvertToTable convert resource to table
func (in *Cluster) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	switch obj := object.(type) {
	case *Cluster:
		return printCluster(obj), nil
	case *ClusterList:
		return printClusterList(obj), nil
	default:
		return nil, fmt.Errorf("unknown type %T", object)
	}
}

var (
	definitions = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string", Format: "name", Description: "the name of the cluster"},
		{Name: "Alias", Type: "string", Description: "the cluster provider type"},
		{Name: "Credential_Type", Type: "string", Description: "the credential type"},
		{Name: "Endpoint", Type: "string", Description: "the endpoint"},
		{Name: "Accepted", Type: "boolean", Description: "the acceptance of the cluster"},
		{Name: "Labels", Type: "string", Description: "the labels of the cluster"},
		{Name: "Creation_Timestamp", Type: "dateTime", Description: "the creation timestamp of the cluster", Priority: 10},
	}
)

func printCluster(in *Cluster) *metav1.Table {
	return &metav1.Table{
		ColumnDefinitions: definitions,
		Rows:              []metav1.TableRow{printClusterRow(in)},
	}
}

func printClusterList(in *ClusterList) *metav1.Table {
	t := &metav1.Table{
		ColumnDefinitions: definitions,
	}
	for _, c := range in.Items {
		t.Rows = append(t.Rows, printClusterRow(c.DeepCopy()))
	}
	return t
}

func printClusterRow(c *Cluster) metav1.TableRow {
	var labels []string
	for k, v := range c.GetLabels() {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: c},
	}
	row.Cells = append(row.Cells,
		c.Name,
		c.Spec.Alias,
		c.Spec.CredentialType,
		c.Spec.Endpoint,
		c.Spec.Accepted,
		strings.Join(labels, ","),
		c.GetCreationTimestamp())
	return row
}
