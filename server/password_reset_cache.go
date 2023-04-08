// Copyright 2022 The Nakama Authors
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
	"context"
	"sync"
	"time"
)

const (
	LockoutTypeEmail LockoutType = iota + 8
	LockoutTypeFrequency
)

type PasswordResetCache interface {
	Stop()
	// Allow checks whether email is locked out or should be allowed to attempt to reset password.
	Allow(email, ip string) LockoutType
	// Validate if the latest expiry time matched.
	Validate(email string, expiry int64) bool
	// Add a attempt.
	Add(email, ip string, expiry time.Time)
	// Reset email attempts on successful password reset.
	Reset(email string)
}

type LocalPasswordResetCache struct {
	sync.RWMutex
	ctx         context.Context
	ctxCancelFn context.CancelFunc

	emailCache map[string]*expireLockoutStatus
	ipCache    map[string]*lockoutStatus
}

type expireLockoutStatus struct {
	lockoutStatus
	epxiredUntil time.Time
}

func NewLocalPasswordResetCache() *LocalPasswordResetCache {
	ctx, ctxCancelFn := context.WithCancel(context.Background())

	c := &LocalPasswordResetCache{
		emailCache: make(map[string]*expireLockoutStatus),
		ipCache:    make(map[string]*lockoutStatus),

		ctx:         ctx,
		ctxCancelFn: ctxCancelFn,
	}

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-c.ctx.Done():
				ticker.Stop()
				return
			case t := <-ticker.C:
				now := t.UTC()
				c.Lock()
				for email, status := range c.emailCache {
					if status.trim(now, lockoutPeriodAccount) {
						if status.epxiredUntil.After(now) {
							delete(c.emailCache, email)
						}
					}
				}
				for ip, status := range c.ipCache {
					if status.trim(now, lockoutPeriodIp) {
						delete(c.ipCache, ip)
					}
				}
				c.Unlock()
			}
		}
	}()

	return c
}

func (c *LocalPasswordResetCache) Stop() {
	c.ctxCancelFn()
}

func (c *LocalPasswordResetCache) Allow(email, ip string) LockoutType {
	now := time.Now()
	c.RLock()
	defer c.RUnlock()
	if status, found := c.emailCache[email]; found {
		// If the email is locked out one second ago, don't allow.
		if len(status.attempts) > 0 {
			if now.Sub(status.attempts[len(status.attempts)-1]) < time.Second {
				return LockoutTypeFrequency
			}
		}
		if !status.lockedUntil.IsZero() && status.lockedUntil.After(now) {
			return LockoutTypeEmail
		}
	}
	if status, found := c.ipCache[ip]; found {
		if !status.lockedUntil.IsZero() && status.lockedUntil.After(now) {
			return LockoutTypeIp
		}
	}
	return LockoutTypeNone
}

func (c *LocalPasswordResetCache) Validate(email string, expiry int64) bool {
	c.RLock()
	defer c.RUnlock()
	if status, found := c.emailCache[email]; found {
		return status.epxiredUntil.Unix() == expiry
	}
	return false
}

func (c *LocalPasswordResetCache) Reset(email string) {
	c.Lock()
	delete(c.emailCache, email)
	c.Unlock()
}

func (c *LocalPasswordResetCache) Add(email, ip string, expiry time.Time) {
	now := time.Now().UTC()
	c.Lock()
	defer c.Unlock()
	if email != "" {
		status, found := c.emailCache[email]
		if !found {
			status = &expireLockoutStatus{}
			c.emailCache[email] = status
		}
		status.epxiredUntil = expiry
		status.attempts = append(status.attempts, now)
		_ = status.trim(now, lockoutPeriodAccount)
		if len(status.attempts) >= maxAttemptsAccount {
			status.lockedUntil = now.Add(lockoutPeriodAccount)
		}
	}
	if ip != "" {
		status, found := c.ipCache[ip]
		if !found {
			status = &lockoutStatus{}
			c.ipCache[ip] = status
		}
		status.attempts = append(status.attempts, now)
		_ = status.trim(now, lockoutPeriodIp)
		if len(status.attempts) >= maxAttemptsIp {
			status.lockedUntil = now.Add(lockoutPeriodIp)
		}
	}
}
