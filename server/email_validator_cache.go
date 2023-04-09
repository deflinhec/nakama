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

type EmailValidatorCache interface {
	Stop()
	// Allow checks whether email is locked out or should be allowed to attempt to reset password.
	Allow(email, ip string) (LockoutType, time.Duration)
	// Validate if the latest expiry time matched.
	Validate(email string, secret string) bool
	// Add a attempt.
	Add(email, ip, secret string, expiry time.Time)
	// Reset email attempts on successful password reset.
	Reset(email string)
}

type LocalEmailValidatorCache struct {
	sync.RWMutex
	ctx         context.Context
	ctxCancelFn context.CancelFunc

	emailCache map[string]*expireLockoutStatus
	ipCache    map[string]*lockoutStatus
}

type expireLockoutStatus struct {
	lockoutStatus
	secret       string
	epxiredUntil time.Time
}

func NewLocalEmailValidatorCache() *LocalEmailValidatorCache {
	ctx, ctxCancelFn := context.WithCancel(context.Background())

	c := &LocalEmailValidatorCache{
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

func (c *LocalEmailValidatorCache) Stop() {
	c.ctxCancelFn()
}

func (c *LocalEmailValidatorCache) Allow(email, ip string) (LockoutType, time.Duration) {
	now := time.Now()
	c.RLock()
	defer c.RUnlock()
	if status, found := c.emailCache[email]; found {
		// If the email is locked out one second ago, don't allow.
		if len(status.attempts) > 0 {
			if now.Sub(status.attempts[len(status.attempts)-1]) < time.Second {
				return LockoutTypeFrequency, time.Second
			}
		}
		if !status.lockedUntil.IsZero() && status.lockedUntil.After(now) {
			return LockoutTypeEmail, status.lockedUntil.Sub(now)
		}
	}
	if status, found := c.ipCache[ip]; found {
		if !status.lockedUntil.IsZero() && status.lockedUntil.After(now) {
			return LockoutTypeIp, status.lockedUntil.Sub(now)
		}
	}
	return LockoutTypeNone, time.Duration(0)
}

func (c *LocalEmailValidatorCache) Validate(email string, secret string) bool {
	c.RLock()
	defer c.RUnlock()
	if status, found := c.emailCache[email]; found {
		return status.secret == secret
	}
	return false
}

func (c *LocalEmailValidatorCache) Reset(email string) {
	c.Lock()
	delete(c.emailCache, email)
	c.Unlock()
}

func (c *LocalEmailValidatorCache) Add(email, ip, secret string, expiry time.Time) {
	now := time.Now().UTC()
	c.Lock()
	defer c.Unlock()
	if email != "" {
		status, found := c.emailCache[email]
		if !found {
			status = &expireLockoutStatus{}
			c.emailCache[email] = status
		}
		status.secret = secret
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
