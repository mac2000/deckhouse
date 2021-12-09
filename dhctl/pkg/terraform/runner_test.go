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

package terraform

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/state"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/cache"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/input"
)

func TestCheckPlanDestructiveChanges(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		destructive bool
		err         error
	}{
		/*{
			name:        "No Changes",
			path:        "./mock/no_changes.tfplan",
			destructive: false,
			err:         nil,
		},
		{
			name:        "Has changes",
			path:        "./mock/has_changes.tfplan",
			destructive: true,
			err:         nil,
		},*/
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			code, err := checkPlanDestructiveChanges(tc.path)
			if tc.err != nil {
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.destructive, code)
		})
	}
}

func newTestRunnerWithChanges() *Runner {
	r := NewRunner("a", "b", "c", "d", &cache.DummyCache{})
	r.changesInPlan = PlanHasChanges
	return r
}

func TestRunnerCreatesStateSaver(t *testing.T) {
	tests := []struct {
		name         string
		cache        state.Cache
		destinations int
	}{
		{
			name:         "Dummy cache does create saver with empty destinations",
			cache:        &cache.DummyCache{},
			destinations: 0,
		},

		{
			name:         "File cache does create saver with empty destinations",
			cache:        &cache.StateCache{},
			destinations: 0,
		},

		{
			name:         "K8s cache does create saver with one destination",
			cache:        &client.StateCache{},
			destinations: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runner := NewRunner("a", "b", "c", "d", tc.cache)
			require.NotNil(t, runner)
			require.NotNil(t, runner.stateSaver)
			require.Len(t, runner.stateSaver.saversDestinations, tc.destinations)
		})
	}
}

func TestCheckRunnerHandleChanges(t *testing.T) {
	tests := []struct {
		name   string
		runner *Runner
		skip   bool
		err    error
	}{
		{
			name: "Yes and skip must not skip",
			skip: false,
			err:  nil,
			runner: newTestRunnerWithChanges().
				WithSkipChangesOnDeny(true).
				WithConfirm(func() *input.Confirmation {
					return input.NewConfirmation().WithYesByDefault()
				}),
		},
		{
			name: "Yes without skip must not skip",
			skip: false,
			err:  nil,
			runner: newTestRunnerWithChanges().
				WithConfirm(func() *input.Confirmation {
					return input.NewConfirmation().WithYesByDefault()
				}),
		},
		{
			name: "No and skip must skip",
			skip: true,
			err:  nil,
			runner: newTestRunnerWithChanges().
				WithSkipChangesOnDeny(true),
		},
		{
			name:   "No without skip must throw an error",
			skip:   false,
			err:    ErrTerraformApplyAborted,
			runner: newTestRunnerWithChanges(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			skip, err := tc.runner.isSkipChanges()
			require.Equal(t, tc.skip, skip)
			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, tc.err, err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
