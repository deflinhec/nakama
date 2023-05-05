package server

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	lua "github.com/heroiclabs/nakama/v3/internal/gopher-lua"
	"go.uber.org/zap"
)

var (
	luaRuntimeModulePatchHistory = LocalRuntimeLuaPatchHistory{
		LocalRuntimePatchHistory: LocalRuntimePatchHistory{
			histories: make([]*RuntimeModulePatchInfo, 0, 32),
		},
		refreshTime: time.Now(),
	}
)

type RuntimeLuaModulePatchRegistry interface {
	Subscribe(uuid.UUID, chan *RuntimeLuaModule)

	Unsubscribe(id uuid.UUID)
}

type LocalRuntimeLuaModulePatchRegistry struct {
	*MapOf[uuid.UUID, chan *RuntimeLuaModule]

	logger *zap.Logger
}

func (mp *LocalRuntimeLuaModulePatchRegistry) Post(module *RuntimeLuaModule) {
	mp.Range(func(key uuid.UUID, ch chan *RuntimeLuaModule) bool {
		// Captuer error in case the channel is closed
		defer func() {
			if x := recover(); x != nil {
				mp.Delete(key)
				mp.logger.Info("Removed Lua module hotfix channel",
					zap.String("mid", key.String()))
			}
		}()
		// Send module to the channel
		select {
		case ch <- module:
			mp.logger.Info("Hotfixing Lua module",
				zap.String("module", module.Name),
				zap.String("mid", key.String()))
		default:
			mp.logger.Warn("Failed to send module to Lua module hotfix channel",
				zap.String("module", module.Name),
				zap.String("mid", key.String()))
		}
		return true
	})
}

func (mp *LocalRuntimeLuaModulePatchRegistry) Subscribe(id uuid.UUID, ch chan *RuntimeLuaModule) {
	mp.Store(id, ch)
	mp.logger.Info("Subscribed to Lua module hotfixes", zap.String("mid", id.String()))
}

func (mp *LocalRuntimeLuaModulePatchRegistry) Unsubscribe(id uuid.UUID) {
	mp.Delete(id)
	mp.logger.Info("Unsubscribed Lua module hotfix channel", zap.String("mid", id.String()))
}

type LocalRuntimeLuaPatchHistory struct {
	LocalRuntimePatchHistory

	refreshTime time.Time
}

func (lh *LocalRuntimeLuaPatchHistory) Refresh(infos []*moduleInfo) {
	lh.RLock()
	defer lh.RUnlock()
	if len(lh.histories) > 0 {
		tail := lh.histories[len(lh.histories)-1]
		if lh.refreshTime.After(tail.UpdateTime) {
			return
		}
	}

	for _, history := range lh.histories {
		if history.UpdateTime.Before(lh.refreshTime) {
			continue
		}
		found := false
		for _, info := range infos {
			if info.path == history.Path {
				if fileInfo, err := os.Stat(history.Path); err == nil {
					info.modTime = fileInfo.ModTime()
				}
				found = true
				break
			}
		}
		if !found {
			if fileInfo, err := os.Stat(history.Path); err == nil {
				infos = append(infos, &moduleInfo{
					path:    history.Path,
					modTime: fileInfo.ModTime(),
				})
			}
		}
	}
	lh.refreshTime = time.Now()
}

func (lh *LocalRuntimeLuaPatchHistory) Add(path string) {
	lh.Lock()
	defer lh.Unlock()
	if len(lh.histories) > 32 {
		lh.histories = lh.histories[1:]
	}
	var name, relPath string
	relPath, _ = filepath.Rel(lua.LuaLDir, path)
	name = strings.TrimSuffix(relPath, filepath.Ext(relPath))
	// Make paths Lua friendly.
	name = strings.ReplaceAll(name, string(os.PathSeparator), ".")
	lh.histories = append(lh.histories, &RuntimeModulePatchInfo{
		Name:       name,
		Path:       path,
		UpdateTime: time.Now(),
	})
}
