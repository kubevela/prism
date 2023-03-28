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

	"github.com/kubevela/pkg/util/apiserver"

	"github.com/kubevela/prism/pkg/util/subresource"
)

// ConvertToTable convert resource to table
func (in *GrafanaDatasource) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	switch obj := object.(type) {
	case *GrafanaDatasource:
		return printGrafanaDatasource(obj), nil
	case *GrafanaDatasourceList:
		return printGrafanaDatasourceList(obj), nil
	default:
		return nil, fmt.Errorf("unknown type %T", object)
	}
}

var (
	definitions = []metav1.TableColumnDefinition{
		{Name: "UID", Type: "string", Format: "name", Description: "the uid of the GrafanaDatasource"},
		{Name: "Name", Type: "string", Format: "name", Description: "the name of the GrafanaDatasource"},
		{Name: "Type", Type: "string", Description: "the type of the GrafanaDatasource"},
		{Name: "URL", Type: "string", Description: "the url of the GrafanaDatasource"},
	}
)

func printGrafanaDatasource(in *GrafanaDatasource) *metav1.Table {
	return &metav1.Table{
		ColumnDefinitions: definitions,
		Rows:              []metav1.TableRow{printGrafanaDatasourceRow(in)},
	}
}

func printGrafanaDatasourceList(in *GrafanaDatasourceList) *metav1.Table {
	t := &metav1.Table{
		ColumnDefinitions: definitions,
	}
	for _, c := range in.Items {
		t.Rows = append(t.Rows, printGrafanaDatasourceRow(c.DeepCopy()))
	}
	return t
}

func printGrafanaDatasourceRow(c *GrafanaDatasource) metav1.TableRow {
	row := metav1.TableRow{
		Object: runtime.RawExtension{Object: c},
	}
	row.Cells = append(row.Cells,
		subresource.NewCompoundName(c.Name).SubResourceName,
		apiserver.GetStringFromRawExtension(&c.Spec, "name"),
		apiserver.GetStringFromRawExtension(&c.Spec, "type"),
		apiserver.GetStringFromRawExtension(&c.Spec, "url"),
	)
	return row
}
