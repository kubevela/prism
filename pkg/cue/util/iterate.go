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

import "cuelang.org/go/cue"

// Iterate over all fields of the cue.Value with fn, if fn returns true,
// iteration stops
func Iterate(value cue.Value, fn func(v cue.Value) (stop bool)) (stop bool) {
	var it *cue.Iterator
	switch value.Kind() {
	case cue.ListKind:
		_it, _ := value.List()
		it = &_it
	default:
		it, _ = value.Fields(cue.Optional(true), cue.Hidden(true), cue.Definitions(true))
	}
	for it != nil && it.Next() {
		if Iterate(it.Value(), fn) {
			return true
		}
	}
	return fn(value)
}
