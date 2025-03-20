package core

import (
	"fmt"
	"ocapi/entity"
)

func (c *Core) AuthenticateByToken(token string) (*entity.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token not provided")
	}
	if c.authKey == token {
		return &entity.User{
			Username: "internal",
		}, nil
	}
	return nil, fmt.Errorf("invalid token")
}
