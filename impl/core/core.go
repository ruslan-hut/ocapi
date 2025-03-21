package core

import (
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
)

type Repository interface {
	ProductSearch(model string) ([]*entity.Product, error)
	SaveProducts(products []*entity.ProductData) error
	SaveProductsDescription(productsDescData []*entity.ProductDescription) error
	SaveCategories(categoriesData []*entity.CategoryData) error
	SaveCategoriesDescription(categoriesDescData []*entity.CategoryDescriptionData) error
	ReadTable(table, filter string, limit int) (interface{}, error)
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
