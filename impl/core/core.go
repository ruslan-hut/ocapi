package core

import (
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
)

type Repository interface {
	ProductSearch(model string) (interface{}, error)
	SaveProducts(products []*entity.ProductData) error
	SaveProductsDescription(productsDescData []*entity.ProductDescription) error
	UpdateProductImage(productUid, image string, isMain bool) error

	SaveCategories(categoriesData []*entity.CategoryData) error
	SaveCategoriesDescription(categoriesDescData []*entity.CategoryDescriptionData) error

	SaveAttributes(attributes []*entity.Attribute) error

	ReadTable(table, filter string, limit int, plain bool) (interface{}, error)
	Stats() string
	CheckApiKey(key string) (string, error)

	FinalizeProductBatch(batchUid string) (int, error)
}

type MessageService interface {
	SendEventMessage(msg *entity.EventMessage) error
}

type Core struct {
	repo      Repository
	ms        MessageService
	authKey   string
	imagePath string
	imageUrl  string
	keys      map[string]string
	log       *slog.Logger
}

func New(log *slog.Logger) *Core {
	return &Core{
		log:  log.With(sl.Module("core")),
		keys: make(map[string]string),
	}
}

func (c *Core) SetRepository(repo Repository) {
	c.repo = repo
}

func (c *Core) SetAuthKey(key string) {
	c.authKey = key
}

func (c *Core) SetImageParameters(imagePath, imageUrl string) {
	c.imagePath = imagePath
	c.imageUrl = imageUrl
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

func (c *Core) FinishBatch(batchUid string) (*entity.BatchResult, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not set")
	}
	if batchUid == "" {
		return nil, fmt.Errorf("batch_uid not set")
	}
	productCount, err := c.repo.FinalizeProductBatch(batchUid)
	if err != nil {
		return entity.NewBatchResult(batchUid, err), nil
	}
	result := entity.NewBatchResult(batchUid, nil)
	result.Products = productCount
	return result, nil
}
