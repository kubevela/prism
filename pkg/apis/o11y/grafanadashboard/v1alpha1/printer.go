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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kubevela/prism/pkg/util/apiserver"
	"github.com/kubevela/prism/pkg/util/subresource"
)

// ConvertToTable convert resource to table
func (in *GrafanaDashboard) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	switch obj := object.(type) {
	case *GrafanaDashboard:
		return printGrafanaDashboard(obj), nil
	case *GrafanaDashboardList:
		return printGrafanaDashboardList(obj), nil
	default:
		return nil, fmt.Errorf("unknown type %T", object)
	}
}

var (
	definitions = []metav1.TableColumnDefinition{
		{Name: "UID", Type: "string", Format: "name", Description: "the name of the GrafanaDashboard"},
		{Name: "Title", Type: "string", Description: "the title of the GrafanaDashboard"},
		{Name: "FolderId", Type: "string", Description: "the folder id of the grafana dashboard", Priority: 10},
	}
)

func printGrafanaDashboard(in *GrafanaDashboard) *metav1.Table {
	return &metav1.Table{
		ColumnDefinitions: definitions,
		Rows:              []metav1.TableRow{printGrafanaDashboardRow(in)},
	}
}

func printGrafanaDashboardList(in *GrafanaDashboardList) *metav1.Table {
	t := &metav1.Table{
		ColumnDefinitions: definitions,
	}
	for _, c := range in.Items {
		t.Rows = append(t.Rows, printGrafanaDashboardRow(c.DeepCopy()))
	}
	return t
}

func printGrafanaDashboardRow(c *GrafanaDashboard) metav1.TableRow {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: c},
	}
	var folderId string
	if labels := c.GetLabels(); labels != nil {
		folderId = labels[grafanaDashboardFolderIdLabelKey]
	}
	row.Cells = append(row.Cells,
		subresource.NewCompoundName(c.Name).SubResourceName,
		apiserver.GetStringFromRawExtension(&c.Spec, "title"),
		folderId,
	)
	return row
}
