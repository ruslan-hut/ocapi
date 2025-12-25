package core

import (
	"fmt"
	"ocapi/entity"
	"time"
)

func (c *Core) AuthenticateByToken(token string) (*entity.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token not provided")
	}

	// Check cache with read lock
	c.keysMu.RLock()
	if cached, ok := c.keys[token]; ok && time.Now().Before(cached.expiresAt) {
		c.keysMu.RUnlock()
		return &entity.User{Username: cached.username}, nil
	}
	c.keysMu.RUnlock()

	// Try database lookup
	userName, err := c.repo.CheckApiKey(token)
	if err == nil {
		c.log.With("username", userName).Debug("user authenticated from database")
		c.keysMu.Lock()
		c.keys[token] = cachedToken{
			username:  userName,
			expiresAt: time.Now().Add(tokenCacheTTL),
		}
		c.keysMu.Unlock()
		return &entity.User{Username: userName}, nil
	}

	// Try config key
	if c.authKey == token {
		userName = "internal"
		c.log.With("username", userName).Debug("user authenticated from config")
		c.keysMu.Lock()
		c.keys[token] = cachedToken{
			username:  userName,
			expiresAt: time.Now().Add(tokenCacheTTL),
		}
		c.keysMu.Unlock()
		return &entity.User{Username: userName}, nil
	}

	return nil, fmt.Errorf("invalid token")
}
