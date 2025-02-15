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

package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/deckhouse/deckhouse/dhctl/pkg/app"
	"github.com/deckhouse/deckhouse/dhctl/pkg/kubernetes/client"
	"github.com/deckhouse/deckhouse/dhctl/pkg/log"
	"github.com/deckhouse/deckhouse/dhctl/pkg/state"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util"
	"github.com/deckhouse/deckhouse/dhctl/pkg/util/cache"
)

var once sync.Once

var (
	_ state.Cache = &cache.StateCache{}
	_ state.Cache = &cache.DummyCache{}
	_ state.Cache = &client.StateCache{}
)

var globalCache state.Cache = &cache.DummyCache{}

func choiceCache(identity string) (state.Cache, error) {
	tmpDir := filepath.Join(app.CacheDir, util.Sha256Encode(identity))
	log.DebugF("Cache dir %s\n", tmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, fmt.Errorf("can't create cache directory: %w", err)
	}

	if app.CacheKubeNamespace == "" {
		return cache.NewStateCache(tmpDir)
	}

	kubeCl := client.NewKubernetesClient()
	err := kubeCl.Init(&client.KubernetesInitParams{
		KubeConfig:          app.CacheKubeConfig,
		KubeConfigContext:   app.CacheKubeConfigContext,
		KubeConfigInCluster: app.CacheKubeConfigInCluster,
	})
	if err != nil {
		return nil, err
	}

	secretName := identity
	if app.CacheKubeName != "" {
		secretName = app.CacheKubeName
	}

	k8sCache := client.NewK8sStateCache(kubeCl, app.CacheKubeNamespace, secretName, tmpDir).
		WithLabels(app.CacheKubeLabels)

	err = k8sCache.Init()
	if err != nil {
		return nil, err
	}

	hasTombstone, err := k8sCache.InCache(state.TombstoneKey)
	if err != nil {
		return nil, err
	}

	if hasTombstone {
		return nil, fmt.Errorf("Cache exchaused")
	}

	return k8sCache, nil
}

func initCache(identity string) error {
	var err error
	once.Do(func() {
		globalCache, err = choiceCache(identity)
	})
	return err
}

func Init(identity string) error {
	return initCache(identity)
}

func Global() state.Cache {
	return globalCache
}

func Dummy() state.Cache {
	return &cache.DummyCache{}
}
