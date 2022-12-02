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
	"cuelang.org/go/cue"
	"cuelang.org/go/cue/format"
)

// ToString stringify cue.Value with reference resolved
func ToString(v cue.Value, opts ...cue.Option) (string, error) {
	opts = append([]cue.Option{cue.Final(), cue.Docs(true), cue.All()}, opts...)
	node := v.Syntax(opts...)
	bs, err := format.Node(node)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}
