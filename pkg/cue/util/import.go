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

package util

import (
	"path/filepath"

	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/parser"
)

func BuildImport(path string, template string) (*build.Instance, error) {
	pkg := &build.Instance{
		PkgName:    filepath.Base(path),
		ImportPath: path,
	}
	file, err := parser.ParseFile("-", template)
	if err == nil {
		err = pkg.AddSyntax(file)
	}
	return pkg, err
}
