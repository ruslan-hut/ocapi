package core

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
	"os"
	"path/filepath"
)

func (c *Core) FindProduct(uid string) (interface{}, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.ProductSearch(uid)
}

func (c *Core) LoadProducts(products []*entity.ProductData) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}
	err := c.repo.SaveProducts(products)
	if err != nil {
		return err
	}
	return nil
}

func (c *Core) LoadProductDescriptions(products []*entity.ProductDescription) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}
	err := c.repo.SaveProductsDescription(products)
	if err != nil {
		return err
	}
	return nil
}

func (c *Core) LoadProductImages(products []*entity.ProductImage) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	for _, product := range products {
		// Decode base64 image data
		fileData, err := base64.StdEncoding.DecodeString(product.FileData)
		if err != nil {
			return fmt.Errorf("decode base64 %s: %v", product.ProductUid, err)
		}

		fileName := fmt.Sprintf("%s%s", product.FileUid, product.FileExt)
		imagePath := filepath.Join(c.imagePath, fileName)

		// Save image file
		err = os.WriteFile(imagePath, fileData, 0644)
		if err != nil {
			return fmt.Errorf("save image %s: %v", product.ProductUid, err)
		}

		imageUrl := fmt.Sprintf("%s%s%s", c.imageUrl, product.FileUid, product.FileExt)

		imageData := entity.NewFromProductImage(product)
		imageData.ImageUrl = imageUrl

		logger := c.log.With(
			slog.String("product_uid", product.ProductUid),
			slog.String("image_url", imageUrl),
			slog.Bool("is_main", product.IsMain),
		)

		err = c.repo.UpdateProductImage(imageData)
		if err != nil {
			logger.Error("update product image", sl.Err(err))
			return fmt.Errorf("product %s: %v", product.ProductUid, err)
		}
		logger.Debug("image loaded")
	}

	return nil
}

func (c *Core) SetProductImages(products []*entity.ProductData) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}
	for _, product := range products {
		err := c.repo.CleanUpProductImages(product.Uid, product.Images)
		if err != nil {
			return fmt.Errorf("product %s: %v", product.Uid, err)
		}
	}
	return nil
}

func (c *Core) LoadProductSpecial(products []*entity.ProductSpecial) error {
	if c.repo == nil {
		return fmt.Errorf("repository not initialized")
	}
	err := c.repo.SaveProductSpecial(products)
	if err != nil {
		return err
	}
	return nil
}
