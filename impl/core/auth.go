package core

import (
	"fmt"
	"ocapi/entity"
)

func (c *Core) AuthenticateByToken(token string) (*entity.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token not provided")
	}

	userName, err := c.repo.CheckApiKey(token)
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
