package core

import (
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
	"os"
	"path/filepath"
	"time"
)

type Repository interface {
	ProductSearch(uid string) (interface{}, error)
	SaveProducts(products []*entity.ProductData) error
	SaveProductsDescription(productsDescData []*entity.ProductDescription) error
	UpdateProductImage(imageData *entity.ProductImageData) error
	CleanUpProductImages(productUid string, images []string) error
	SaveProductAttributes(attributes []*entity.ProductAttribute) error
	SaveProductSpecial(products []*entity.ProductSpecial) error

	SaveCategories(categoriesData []*entity.CategoryData) error
	SaveCategoriesDescription(categoriesDescData []*entity.CategoryDescriptionData) error

	SaveAttributes(attributes []*entity.Attribute) error

	OrderSearchId(orderId int64) (*entity.Order, error)
	OrderSearchStatus(statusId int64, from time.Time) ([]int64, error)
	OrderProducts(orderId int64) ([]*entity.ProductOrder, error)
	OrderTotals(orderId int64) ([]*entity.OrderTotal, error)
	UpdateOrderStatus(orderId int64, statusId int, comment string) error

	UpdateCurrencyValue(currencyCode string, value float64) error

	ReadTable(table, filter string, limit int, plain bool) (interface{}, error)
	DeleteRecords(table, filter string) (int64, error)
	Stats() string
	CheckApiKey(key string) (string, error)

	FinalizeProductBatch(batchUid string) (int, error)
	GetAllImages() ([]string, error)
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
	count, err := c.checkImageFiles()
	if err != nil {
		c.log.Warn("check files", sl.Err(err))
	}
	result.DeletedFiles = count
	return result, nil
}

func (c *Core) checkImageFiles() (int, error) {
	if c.repo == nil {
		return 0, fmt.Errorf("repository not set")
	}

	images, err := c.repo.GetAllImages()
	if err != nil {
		return 0, err
	}
	//extract file names from a path
	validImages := make(map[string]bool)
	for _, image := range images {
		validImages[filepath.Base(image)] = true
	}

	var count int
	err = filepath.Walk(c.imagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Check if the current file is in the validImages list
		file := filepath.Base(path)
		if validImages[file] {
			return nil
		}

		// If the file is not in the validImages list, delete it
		if err = os.Remove(path); err != nil {
			c.log.Error("removing image", sl.Err(err))
			return err
		}
		c.log.With(slog.String("image", path)).Info("file removed")
		count++

		return nil
	})

	return count, err
}

func (c *Core) UpdateRates(data []*entity.Currency) error {
	if c.repo == nil {
		return fmt.Errorf("repository not set")
	}
	if len(data) == 0 {
		return fmt.Errorf("currency data is empty")
	}

	for _, currency := range data {
		err := c.repo.UpdateCurrencyValue(currency.Code, currency.Rate)
		if err != nil {
			c.log.Error("updating currency rate", sl.Err(err), slog.String("currency", currency.Code))
			return fmt.Errorf("failed to update currency %s: %w", currency.Code, err)
		}
	}
	return nil
}
