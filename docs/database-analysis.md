# Database Analysis - OCAPI

This document describes all database tables used by OCAPI and the conditions under which data is created or updated.

## Table Index

| # | Table | Category | Description |
|---|-------|----------|-------------|
| 1 | [product](#1-product) | Products | Main product catalog |
| 2 | [product_description](#2-product_description) | Products | Multi-language product names/descriptions |
| 3 | [product_image](#3-product_image) | Products | Additional product images |
| 4 | [product_special](#4-product_special) | Products | Special/promotional pricing |
| 5 | [product_to_category](#5-product_to_category) | Products | Product-category relationships |
| 6 | [product_to_store](#6-product_to_store) | Products | Product store visibility |
| 7 | [product_to_layout](#7-product_to_layout) | Products | Product layout assignments |
| 8 | [product_attribute](#8-product_attribute) | Products | Product attribute values |
| 9 | [category](#9-category) | Categories | Product category hierarchy |
| 10 | [category_description](#10-category_description) | Categories | Multi-language category names |
| 11 | [category_to_store](#11-category_to_store) | Categories | Category store visibility |
| 12 | [attribute](#12-attribute) | Attributes | Attribute definitions |
| 13 | [attribute_description](#13-attribute_description) | Attributes | Multi-language attribute names |
| 14 | [manufacturer](#14-manufacturer) | Other | Product manufacturers/brands |
| 15 | [manufacturer_to_store](#15-manufacturer_to_store) | Other | Manufacturer store visibility |
| 16 | [order](#16-order-read-only-except-status) | Orders | Customer orders |
| 17 | [order_product](#17-order_product-read-only) | Orders | Products within orders |
| 18 | [order_total](#18-order_total-read-only) | Orders | Order totals |
| 19 | [order_history](#19-order_history) | Orders | Order status history |
| 20 | [currency](#20-currency) | Other | Currency exchange rates |
| 21 | [api](#21-api) | Other | API key authentication |

---

## Overview

OCAPI works with an OpenCart database using a configurable table prefix (default: `prefix_`). The API adds custom UID columns to support external system integration.

## Custom Columns Added by OCAPI

On startup, the application automatically adds these columns if they don't exist:

| Table | Column | Type | Purpose |
|-------|--------|------|---------|
| `product` | `product_uid` | VARCHAR(64) | External unique identifier |
| `product` | `batch_uid` | VARCHAR(64) | Batch processing identifier |
| `category` | `category_uid` | VARCHAR(64) | External unique identifier |
| `category` | `parent_uid` | VARCHAR(64) | Parent category external ID |
| `attribute` | `attribute_uid` | VARCHAR(64) | External unique identifier |
| `product_image` | `file_uid` | VARCHAR(64) | External file identifier |

---

## Database Tables and Update Conditions

### 1. `product`

**Purpose:** Main product catalog table

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `product_id` | x | | Auto-increment PK |
| `product_uid` | x | x | External unique identifier (lookup key) |
| `batch_uid` | | x | Batch processing marker |
| `model` | | x | Article/model number |
| `sku` | | | Stock keeping unit (use CustomFields) |
| `upc` | | | Universal product code (use CustomFields) |
| `ean` | | | European article number (use CustomFields) |
| `jan` | | | Japanese article number (use CustomFields) |
| `isbn` | | | Book identifier (use CustomFields) |
| `mpn` | | | Manufacturer part number (use CustomFields) |
| `location` | | | Storage location (use CustomFields) |
| `quantity` | | x | Stock quantity |
| `stock_status_id` | | x | Stock status reference |
| `price` | | x | Base price |
| `manufacturer_id` | | x | Manufacturer reference |
| `status` | | x | Active/inactive flag |
| `weight` | | x | Product weight |
| `weight_class_id` | | x | Weight unit reference |
| `image` | x | x | Main product image path |
| `minimum` | | x | Minimum order quantity (default: 1) |
| `subtract` | | x | Subtract from stock (default: 1) |
| `shipping` | | x | Requires shipping (default: 1) |
| `tax_class_id` | | x | Tax class (default: 9) |
| `length_class_id` | | x | Length unit (default: 0) |
| `date_available` | | x | Availability date |
| `date_added` | | x | Creation timestamp |
| `date_modified` | | x | Last update timestamp |

**INSERT Condition:**
- When `SaveProducts()` is called and no product exists with the given `product_uid`

**UPDATE Condition:**
- When `SaveProducts()` is called and a product already exists with the given `product_uid`

**Additional Operations:**
- `UpdateProductImage()`: Updates `image` column by `product_uid`
- `FinalizeProductBatch()`: Sets `status=0` for products not in batch, clears `batch_uid`
- Custom fields can be updated via `CustomFields` array in request

---

### 2. `product_description`

**Purpose:** Multi-language product names and descriptions

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `product_id` | x | x | Product reference (composite PK) |
| `language_id` | x | x | Language reference (composite PK) |
| `name` | x | x | Product name |
| `description` | x | x | Product description (HTML) |
| `meta_title` | | x | SEO title (set to name on insert) |

**INSERT Condition:**
- When no record exists for the given `product_id` + `language_id` combination

**UPDATE Condition:**
- When a record exists for the given `product_id` + `language_id`
- If `UpdateDescription=false` or description is empty: only `name` is updated
- If `UpdateDescription=true` and description provided: both `name` and `description` updated

---

### 3. `product_image`

**Purpose:** Additional product images (non-main images)

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `product_image_id` | x | | Auto-increment PK |
| `product_id` | x | x | Product reference |
| `file_uid` | x | x | External file identifier (lookup key) |
| `image` | | x | Image file path |
| `sort_order` | | x | Display order |

**INSERT Condition:**
- When `UpdateProductImage()` is called with `IsMain=false`
- And no record exists for the given `product_id` + `file_uid`

**UPDATE Condition:**
- When a record exists for the given `product_id` + `file_uid`
- Only `sort_order` is updated

**DELETE Condition:**
- `CleanUpProductImages()` removes images where:
  - `file_uid` is not in the provided valid images list
  - OR `file_uid` is a duplicate (already seen for this product)

---

### 4. `product_special`

**Purpose:** Special/promotional pricing per customer group

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `product_special_id` | x | | Auto-increment PK (for lookup) |
| `product_id` | x | x | Product reference (composite key) |
| `customer_group_id` | x | x | Customer group (composite key) |
| `price` | | x | Special price |
| `priority` | | x | Priority order |
| `date_start` | | x | Start date |
| `date_end` | | x | End date |

**INSERT Condition:**
- When no record exists for the given `product_id` + `customer_group_id`

**UPDATE Condition:**
- When a record exists for the given `product_id` + `customer_group_id`

---

### 5. `product_to_category`

**Purpose:** Product-to-category relationships

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `product_id` | x | x | Product reference (composite PK) |
| `category_id` | | x | Category reference (composite PK) |

**DELETE + INSERT (Replace):**
- When `setProductCategories()` is called during product save
- All existing category relationships for the product are deleted
- New relationships are inserted for each category UID

---

### 6. `product_to_store`

**Purpose:** Product visibility in stores

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `product_id` | | x | Product reference |
| `store_id` | | x | Store reference (always 0) |

**INSERT Condition:**
- Only when a new product is created
- Always inserts with `store_id=0` (default store)

---

### 7. `product_to_layout`

**Purpose:** Product layout assignments

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `product_id` | | x | Product reference |
| `store_id` | | x | Store reference (always 0) |
| `layout_id` | | x | Layout reference (always 0) |

**INSERT Condition:**
- Only when a new product is created
- Always inserts with `store_id=0`, `layout_id=0`

---

### 8. `product_attribute`

**Purpose:** Product attribute values (multi-language)

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `product_id` | x | x | Product reference (composite PK) |
| `attribute_id` | x | x | Attribute reference (composite PK) |
| `language_id` | x | x | Language reference (composite PK) |
| `text` | x | x | Attribute value |

**INSERT Condition:**
- When no record exists for `product_id` + `attribute_id` + `language_id`

**UPDATE Condition:**
- When a record exists for the combination above

---

### 9. `category`

**Purpose:** Product category hierarchy

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `category_id` | x | | Auto-increment PK |
| `category_uid` | x | x | External unique identifier (lookup key) |
| `parent_id` | | x | Parent category reference |
| `parent_uid` | | x | Parent external identifier |
| `top` | | x | Show in top menu |
| `sort_order` | | x | Display order |
| `status` | | x | Active/inactive flag |
| `date_added` | | x | Creation timestamp |
| `date_modified` | | x | Last update timestamp |

**INSERT Condition:**
- When `getCategoryByUID()` is called and no category exists with the given UID
- Auto-creates minimal category record

**UPDATE Condition:**
- When `SaveCategories()` is called with existing category UIDs

---

### 10. `category_description`

**Purpose:** Multi-language category names and descriptions

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `category_id` | x | x | Category reference (composite PK) |
| `language_id` | x | x | Language reference (composite PK) |
| `name` | x | x | Category name |
| `description` | x | x | Category description (HTML) |
| `meta_title` | | x | SEO title (set to name on insert) |
| `meta_description` | | x | SEO description (set to name on insert) |

**INSERT Condition:**
- When no record exists for the given `category_id` + `language_id`

**UPDATE Condition:**
- When a record exists for the given `category_id` + `language_id`
- If description is empty: only `name` is updated
- If description provided: both `name` and `description` updated

---

### 11. `category_to_store`

**Purpose:** Category visibility in stores

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `category_id` | | x | Category reference |
| `store_id` | | x | Store reference (always 0) |

**INSERT Condition:**
- When a new category is auto-created
- Always inserts with `store_id=0`

---

### 12. `attribute`

**Purpose:** Attribute definitions

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `attribute_id` | x | | Auto-increment PK |
| `attribute_uid` | x | x | External unique identifier (lookup key) |
| `attribute_group_id` | | x | Attribute group reference |
| `sort_order` | | x | Display order |

**INSERT Condition:**
- When `SaveAttributes()` is called and no attribute exists with the given UID

**UPDATE Condition:**
- When an attribute exists with the given UID

---

### 13. `attribute_description`

**Purpose:** Multi-language attribute names

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `attribute_id` | x | x | Attribute reference (composite PK) |
| `language_id` | x | x | Language reference (composite PK) |
| `name` | x | x | Attribute name |

**INSERT Condition:**
- When no record exists for the given `attribute_id` + `language_id`

**UPDATE Condition:**
- When a record exists for the combination above

---

### 14. `manufacturer`

**Purpose:** Product manufacturers/brands

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `manufacturer_id` | x | | Auto-increment PK |
| `name` | x | x | Manufacturer name (lookup key) |

**INSERT Condition:**
- When `getManufacturerId()` is called with a manufacturer name not in the database

*Note: Lookup is by `name`, not by external UID*

---

### 15. `manufacturer_to_store`

**Purpose:** Manufacturer visibility in stores

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `manufacturer_id` | | x | Manufacturer reference |
| `store_id` | | x | Store reference (always 0) |

**INSERT Condition:**
- When a new manufacturer is created
- Always inserts with `store_id=0`

---

### 16. `order` (Read-Only except status)

**Purpose:** Customer orders

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `order_id` | x | | PK, used for lookup |
| `invoice_no` | x | | Invoice number |
| `invoice_prefix` | x | | Invoice prefix |
| `store_id` | x | | Store reference |
| `store_name` | x | | Store name |
| `store_url` | x | | Store URL |
| `customer_id` | x | | Customer reference |
| `customer_group_id` | x | | Customer group |
| `firstname` | x | | Customer first name |
| `lastname` | x | | Customer last name |
| `email` | x | | Customer email |
| `telephone` | x | | Customer phone |
| `custom_field` | x | | Custom fields JSON |
| `payment_firstname` | x | | Billing first name |
| `payment_lastname` | x | | Billing last name |
| `payment_company` | x | | Billing company |
| `payment_address_1` | x | | Billing address line 1 |
| `payment_address_2` | x | | Billing address line 2 |
| `payment_city` | x | | Billing city |
| `payment_postcode` | x | | Billing postal code |
| `payment_country` | x | | Billing country name |
| `payment_country_id` | x | | Billing country ID |
| `payment_zone` | x | | Billing zone/state |
| `payment_zone_id` | x | | Billing zone ID |
| `payment_address_format` | x | | Address format template |
| `payment_custom_field` | x | | Payment custom fields |
| `payment_method` | x | | Payment method name |
| `payment_code` | x | | Payment method code |
| `shipping_firstname` | x | | Shipping first name |
| `shipping_lastname` | x | | Shipping last name |
| `shipping_company` | x | | Shipping company |
| `shipping_address_1` | x | | Shipping address line 1 |
| `shipping_address_2` | x | | Shipping address line 2 |
| `shipping_city` | x | | Shipping city |
| `shipping_postcode` | x | | Shipping postal code |
| `shipping_country` | x | | Shipping country name |
| `shipping_country_id` | x | | Shipping country ID |
| `shipping_zone` | x | | Shipping zone/state |
| `shipping_zone_id` | x | | Shipping zone ID |
| `shipping_address_format` | x | | Address format template |
| `shipping_custom_field` | x | | Shipping custom fields |
| `shipping_method` | x | | Shipping method name |
| `shipping_code` | x | | Shipping method code |
| `comment` | x | | Order comment |
| `total` | x | | Order total |
| `order_status_id` | x | x | Order status (R/W) |
| `affiliate_id` | x | | Affiliate reference |
| `commission` | x | | Affiliate commission |
| `marketing_id` | x | | Marketing campaign ID |
| `tracking` | x | | Tracking code |
| `language_id` | x | | Language reference |
| `currency_id` | x | | Currency reference |
| `currency_code` | x | | Currency code |
| `currency_value` | x | | Currency exchange rate |
| `ip` | x | | Customer IP |
| `forwarded_ip` | x | | Forwarded IP |
| `user_agent` | x | | Browser user agent |
| `accept_language` | x | | Browser language |
| `date_added` | x | | Order creation date |
| `date_modified` | x | x | Last update (R/W) |

**READ Operations:**
- `OrderSearchId()`: Fetch single order by `order_id`
- `OrderSearchStatus()`: List order IDs by `order_status_id` after a given date

**UPDATE Condition:**
- `UpdateOrderStatus()`: Updates `order_status_id` and `date_modified`

---

### 17. `order_product` (Read-Only)

**Purpose:** Products within orders

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `order_id` | x | | Order reference (lookup key) |
| `product_id` | x | | Product reference |
| `name` | x | | Product name at order time |
| `model` | x | | Product model |
| `sku` | x | | Stock keeping unit |
| `upc` | x | | Universal product code |
| `ean` | x | | European article number |
| `jan` | x | | Japanese article number |
| `isbn` | x | | Book identifier |
| `mpn` | x | | Manufacturer part number |
| `location` | x | | Storage location |
| `quantity` | x | | Ordered quantity |
| `price` | x | | Unit price |
| `total` | x | | Line total |
| `tax` | x | | Tax amount |
| `reward` | x | | Reward points |
| `weight` | x | | Product weight |
| `discount_amount` | x | | Discount amount |
| `discount_type` | x | | Discount type |

*Note: Joins with `product` table to include `product_uid` in response*

**READ Operation:**
- `OrderProducts()`: Fetches products for a given `order_id`

---

### 18. `order_total` (Read-Only)

**Purpose:** Order totals (subtotal, tax, shipping, etc.)

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `order_id` | x | | Order reference (lookup key) |
| `code` | x | | Total type code (sub_total, tax, shipping, total) |
| `title` | x | | Display title |
| `value` | x | | Amount value |

**READ Operation:**
- `OrderTotals()`: Fetches totals for a given `order_id`

---

### 19. `order_history`

**Purpose:** Order status change history

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `order_id` | | x | Order reference |
| `order_status_id` | | x | New status value |
| `notify` | | x | Customer notification flag (always 0) |
| `comment` | | x | Status change comment |
| `date_added` | | x | Timestamp of change |

**INSERT Condition:**
- When `UpdateOrderStatus()` successfully changes the order status
- Creates a history record with timestamp

---

### 20. `currency`

**Purpose:** Currency definitions and exchange rates

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `code` | x | | Currency code (lookup key, e.g., USD, EUR) |
| `value` | | x | Exchange rate value |
| `date_modified` | | x | Last update timestamp (set to NOW()) |

**UPDATE Condition:**
- `UpdateCurrencyValue()`: Updates `value` and `date_modified` by `code`

---

### 21. `api`

**Purpose:** API key authentication

**Fields Used:**

| Field | R | W | Notes |
|-------|---|---|-------|
| `key` | x | | API key (lookup key) |
| `username` | x | | Associated username (returned on success) |
| `status` | x | | Must be 1 for active API keys |

**READ Operation:**
- `CheckApiKey()`: Validates API key, returns username
- Queries by `key` where `status=1`

---

## Summary: Upsert Logic Patterns

| Entity | Lookup Key | Strategy |
|--------|------------|----------|
| Product | `product_uid` | Upsert (insert if not exists) |
| Product Description | `product_id` + `language_id` | Upsert |
| Product Image | `product_id` + `file_uid` | Upsert (update sort_order only) |
| Product Special | `product_id` + `customer_group_id` | Upsert |
| Product Attribute | `product_id` + `attribute_id` + `language_id` | Upsert |
| Product Categories | `product_id` | Replace all |
| Category | `category_uid` | Auto-create if not exists |
| Category Description | `category_id` + `language_id` | Upsert |
| Attribute | `attribute_uid` | Upsert |
| Attribute Description | `attribute_id` + `language_id` | Upsert |
| Manufacturer | `name` | Auto-create if not exists |
| Order Status | `order_id` | Update only |
| Currency | `code` | Update only |

## Batch Processing

The `FinalizeProductBatch()` function implements batch synchronization:

1. Counts products with the given `batch_uid`
2. Fails if batch is empty (safety check)
3. Sets `status=0` for all products NOT in the batch
4. Clears `batch_uid` from all products
5. Returns count of active products (`status=1`)

This enables full catalog synchronization where products not in the import batch are deactivated.

## Image Management

`CleanUpProductImages()` removes orphaned product images:

1. Fetches all images for a product from `product_image`
2. Compares against provided list of valid `file_uid` values
3. Deletes images not in the valid list
4. Removes duplicate entries (same `file_uid` appearing multiple times)

`GetAllImages()` retrieves all image paths from both `product.image` and `product_image.image` for filesystem cleanup.
