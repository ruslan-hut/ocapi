package core

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"ocapi/entity"
	"ocapi/internal/lib/sl"
	"os"
	"path/filepath"
	"strings"
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
		// Filter out the main image UID — it belongs in the product table, not product_image
		mainImageUid := c.mainImageUid(product.Uid)
		additionalImages := product.Images
		if mainImageUid != "" {
			additionalImages = make([]string, 0, len(product.Images))
			for _, uid := range product.Images {
				if uid != mainImageUid {
					additionalImages = append(additionalImages, uid)
				}
			}
		}

		// Clean up images that are no longer in the list and get existing UIDs
		existing, err := c.repo.CleanUpProductImages(product.Uid, additionalImages)
		if err != nil {
			return fmt.Errorf("product %s: %v", product.Uid, err)
		}

		// Insert images that are in the request but not yet in the database
		for i, fileUid := range additionalImages {
			if existing[fileUid] {
				continue
			}

			// Resolve image file on disk by mask {imagePath}/{fileUid}*
			imageUrl, err := c.resolveImageUrl(fileUid)
			if err != nil {
				return fmt.Errorf("product %s, image %s: %v", product.Uid, fileUid, err)
			}

			err = c.repo.InsertProductImage(product.Uid, fileUid, imageUrl, i)
			if err != nil {
				return fmt.Errorf("product %s, insert image %s: %v", product.Uid, fileUid, err)
			}
		}
	}
	return nil
}

// mainImageUid extracts the file UID from the product's main image path.
// Returns empty string if the main image is not set or on error.
func (c *Core) mainImageUid(productUid string) string {
	mainImage, err := c.repo.GetProductMainImage(productUid)
	if err != nil || mainImage == "" {
		return ""
	}
	// Extract UID from path like "catalog/products/798b00f4-d767-11f0-9d8e-0cc47a39a0b2.jpg"
	base := filepath.Base(mainImage)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// resolveImageUrl finds the image file on disk by file_uid glob pattern
// and returns the relative URL for the database (e.g. "catalog/product/uid.jpg").
func (c *Core) resolveImageUrl(fileUid string) (string, error) {
	pattern := filepath.Join(c.imagePath, fileUid+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob %s: %v", pattern, err)
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("image file not found: %s", pattern)
	}

	// Use the first match; extract the file extension
	fileName := filepath.Base(matches[0])
	imageUrl := c.imageUrl + fileName
	return imageUrl, nil
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
