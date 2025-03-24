package core

import (
	"fmt"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
)

func (c *Core) AuthenticateByToken(token string) (*entity.User, error) {
	if token == "" {
		return nil, fmt.Errorf("token not provided")
	}

	// encode token to base64
	// tokenB64 := base64.StdEncoding.EncodeToString([]byte(token))

	userName, err := c.repo.CheckApiKey(token)
	if err == nil {
		return &entity.User{
			Username: userName,
		}, nil
	}
	c.log.With(sl.Err(err)).Debug("check api key failed")

	if c.authKey == token {
		return &entity.User{
			Username: "internal",
		}, nil
	}
	return nil, fmt.Errorf("invalid token")
}
