package core

import (
	"encoding/base64"
	"fmt"
	"ocapi/entity"
)

func (c *Core) AuthenticateByToken(token string) (*entity.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token not provided")
	}

	// encode token to base64
	tokenB64 := base64.StdEncoding.EncodeToString([]byte(token))

	userName, err := c.repo.CheckApiKey(tokenB64)
	if err == nil {
		return &entity.User{
			Username: userName,
		}, nil
	}

	if c.authKey == token {
		return &entity.User{
			Username: "internal",
		}, nil
	}
	return nil, fmt.Errorf("invalid token")
}
