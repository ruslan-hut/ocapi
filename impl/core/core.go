package core

import (
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
)

type Repository interface {
	ProductSearch(model string) ([]*entity.Product, error)
}

type MessageService interface {
	SendEventMessage(msg *entity.EventMessage) error
}

type Core struct {
	repo    Repository
	ms      MessageService
	authKey string
	log     *slog.Logger
}

func (c *Core) FindModel(model string) ([]*entity.Product, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.ProductSearch(model)
}

func New(repo Repository, key string, log *slog.Logger) *Core {
	return &Core{
		repo:    repo,
		authKey: key,
		log:     log.With(sl.Module("core")),
	}
}

func (c *Core) SetMessageService(ms MessageService) {
	c.ms = ms
}

func (c *Core) SendMail(message *entity.MailMessage) (interface{}, error) {
	return nil, nil
}

func (c *Core) SendEvent(message *entity.EventMessage) (interface{}, error) {
	if c.ms == nil {
		return nil, fmt.Errorf("not set MessageService")
	}
	return nil, c.ms.SendEventMessage(message)
}

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
