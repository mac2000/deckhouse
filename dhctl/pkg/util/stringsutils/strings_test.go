// Copyright 2021 Flant JSC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package stringsutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDifference(t *testing.T) {
	cases := []struct {
		name   string
		first  []string
		second []string
		result []string
	}{
		{
			name:   "Two empty array return empty array",
			first:  make([]string, 0),
			second: make([]string, 0),
			result: make([]string, 0),
		},

		{
			name:   "Second array is empty returns first array",
			first:  []string{"a"},
			second: make([]string, 0),
			result: []string{"a"},
		},

		{
			name:   "First array is empty returns empty array",
			first:  make([]string, 0),
			second: []string{"a"},
			result: make([]string, 0),
		},

		{
			name:   "Two same arrays return empty array",
			first:  []string{"a"},
			second: []string{"a"},
			result: make([]string, 0),
		},

		{
			name:   "Exclude all from first",
			first:  []string{"a", "b"},
			second: []string{"a", "b", "c"},
			result: make([]string, 0),
		},

		{
			name:   "Exclude some from first",
			first:  []string{"a", "b", "c"},
			second: []string{"a", "c"},
			result: []string{"b"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			res := DifferenceSlices(testCase.first, testCase.second)

			require.Equal(t, res, testCase.result)
		})
	}
}
