package core

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"ocapi/entity"
	"os"
	"path/filepath"
)

func (c *Core) FindModel(model string) (interface{}, error) {
	if c.repo == nil {
		return nil, fmt.Errorf("repository not initialized")
	}
	return c.repo.ProductSearch(model)
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
	fileExt := ".png"

	for _, product := range products {
		// Decode base64 image data
		imageData, err := base64.StdEncoding.DecodeString(product.FileData)
		if err != nil {
			return fmt.Errorf("decode base64 %s: %v", product.ProductUid, err)
		}

		fileName := fmt.Sprintf("%s%s", product.FileUid, fileExt)
		imagePath := filepath.Join(c.imagePath, fileName)

		// Save image file
		err = os.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			return fmt.Errorf("save image %s: %v", product.ProductUid, err)
		}

		imageUrl := fmt.Sprintf("%s%s%s", c.imageUrl, product.FileUid, fileExt)

		logger := c.log.With(
			slog.String("product_uid", product.ProductUid),
			slog.String("image_url", imageUrl),
			slog.Bool("is_main", product.IsMain),
		)

		if product.IsMain {
			// Update image path in repository
			err = c.repo.UpdateProductImage(product.ProductUid, imageUrl)
			if err != nil {
				return fmt.Errorf("product %s: %v", product.ProductUid, err)
			}
		} else {
			err = c.repo.UpdateProductNotMainImage(product.ProductUid, imageUrl)
			if err != nil {
				return fmt.Errorf("product %s: %v", product.ProductUid, err)
			}
		}
		logger.Debug("image loaded")
	}

	return nil
}
