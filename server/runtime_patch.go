// Copyright 2023 Deflinhec
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"sync"
	"time"
)

type RuntimeModulePatchInfo struct {
	Name       string
	Path       string
	UpdateTime time.Time
}

type RuntimeModulePatchHistory interface {
	Add(path string)

	Refresh(infos []*moduleInfo)

	List() []*RuntimeModulePatchInfo
}

type LocalRuntimePatchHistory struct {
	sync.RWMutex

	histories []*RuntimeModulePatchInfo
}

func (lh *LocalRuntimePatchHistory) Add(path string) {
}

func (lh *LocalRuntimePatchHistory) Refresh(infos []*moduleInfo) {
}

func (lh *LocalRuntimePatchHistory) List() []*RuntimeModulePatchInfo {
	lh.RLock()
	defer lh.RUnlock()
	clone := make([]*RuntimeModulePatchInfo, len(lh.histories))
	copy(clone, lh.histories)
	return clone
}
