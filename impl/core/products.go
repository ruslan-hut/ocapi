package core

import (
	"encoding/base64"
	"fmt"
	"ocapi/entity"
	"os"
	"path/filepath"
)

func (c *Core) FindModel(model string) ([]*entity.Product, error) {
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
			return fmt.Errorf("decode base64 image %s: %v", product.ProductUid, err)
		}

		// Define the image file path
		imagePath := filepath.Join(c.imagePath, product.FileUid, fileExt)

		// Save image file
		err = os.WriteFile(imagePath, imageData, 0644)
		if err != nil {
			return fmt.Errorf("save product image %s: %v", product.ProductUid, err)
		}

		if product.IsMain {
			imageUrl := fmt.Sprintf("%s%s%s", c.imageUrl, product.FileUid, fileExt)
			// Update image path in repository
			err = c.repo.UpdateProductImage(product.ProductUid, imageUrl)

			if err != nil {
				return fmt.Errorf("update product image %s: %v", product.ProductUid, err)
			}
		}
	}

	return nil
}
