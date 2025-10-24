## API v1 Description

### Authentication
To make requests to the API, you need to provide Bearer token in the `Authorization` header.
OCAPI supports two ways of token storage: in the configuration file, `listen` section, and in the OpenCart API section inside the admin panel.
```yaml
listen:
  bind_ip: 127.0.0.1
  port: 9800
  key: api-key        # API key for the OCAPI service
```

### Product Management

#### Update or Create Product
- **Endpoint:** `/api/v1/product`
- **Method:** `POST`
- **Description:** Updates the details of a specific product. If product is not found, it will be created.
- **Request Body:**
  ```json
  {
    "data": [
        {
            "product_uid": "28ac4a2c-6f4c-11ef-b7f7-00155d018000",
            "article": "scMUSE",
            "quantity": 6,
            "price": 25,
            "active": true,
            "categories": ["29b666d4-bc22-11ee-b7b4-00155d018000"]
        }
    ]
  }
  ```
- **Response:**
  ```json
  {
    "success": true,
    "status_message": "Success",
    "timestamp": "2025-03-24T11:22:39Z"
  }
  ```

#### Update or Add Product Description
- **Endpoint:** `/api/v1/product/description`
- **Method:** `POST`
- **Description:** Updates the description of a specific product.
- **Request Body:**
  ```json
    {
    "data": [
            {
                "language_id": 1,
                "product_uid": "28ac4a2c-6f4c-11ef-b7f7-00155d018000",
                "name": "Spa candle MUSE, 30 g",
                "description": "The candle is made of natural soy wax. The aroma of the candle is a combination of the scents of the forest and the sea. The candle is packed in a beautiful gift box."  
            }
        ]
    }
  ```

### Categories. Products Hierarchy

#### Update or Create Category
- **Endpoint:** `/api/v1/category`
- **Method:** `POST`
- **Description:** Updates category. If category is not found, it will be created.
- **Request Body:**
  ```json
  {
    "data": [
        {
            "sort_order": 0,
            "active": true,
            "parent_uid": "",
            "menu": true,
            "category_uid": "6666bc6a-a487-11e9-b6d3-00155d010d00",
            "article": ""
        }
    ]
  }
  ```

#### Update or Add Category Description
- **Endpoint:** `/api/v1/category/description`
- **Method:** `POST`
- **Description:** Updates category description.
- **Request Body:**
  ```json
  {
    "data": [
        {
            "language_id": 1,
            "category_uid": "6666bc6a-a487-11e9-b6d3-00155d010d00",
            "name": "ALL FOR EXTENSION",
            "description": "The category includes all the necessary materials for hair extension."
        }
    ]
  }
  ```

#### Get Products
- **Endpoint:** `/api/v1/product/{uid}`
- **Method:** `GET`
- **Description:** Retrieves a product data, a record from the database.
- **Response:**
  ```json
  {
    "data": [
        {
            "batch_uid": "",
            "date_added": "2024-10-24T11:52:25Z",
            "date_available": "2024-10-21T00:00:00Z",
            "date_modified": "2025-03-24T09:33:42Z",
            "ean": "",
            "height": "0.00000000",
            "image": "import/563235c5-8ab8-11ef-b7fb-00155d018000.png",
            "isbn": "",
            "jan": "",
            "length": "0.00000000",
            "length_class_id": 1,
            "location": "",
            "manufacturer_id": 0,
            "max_discount": "0.00",
            "meta_robots": "",
            "minimum": 1,
            "model": "doilon3",
            "mpn": "",
            "points": 0,
            "price": "0.0000",
            "product_id": 5970,
            "product_uid": "02bc1ea8-70d3-11ef-b7f7-00155d018000",
            "quantity": 354,
            "sku": "",
            "sort_order": 0,
            "status": 1,
            "stock_status_id": 7,
            "subtract": 1,
            "tax_class_id": 9,
            "upc": "",
            "viewed": 0,
            "weight": "0.00000000",
            "weight_class_id": 1,
            "width": "0.00000000"
        }
    ],
    "success": true,
    "status_message": "Success",
    "timestamp": "2025-03-24T09:36:34Z"
    }
    ```

### Order Management

#### Get Order by ID
- Endpoint: `/api/v1/order/{orderId}`
- Method: `GET`
- Description: Retrieves a single order with customer, shipping/payment info, products, and totals.
- Response:
  ```json
  {
    "data": [
      {
        "order_id": 10234,
        "invoice_no": "INV-000123",
        "invoice_prefix": "INV-",
        "store_id": 0,
        "store_name": "Default",
        "store_url": "https://example.com/",
        "customer_id": 501,
        "customer_group_id": 1,
        "firstname": "Jane",
        "lastname": "Doe",
        "email": "jane.doe@example.com",
        "telephone": "+1-555-1234",
        "custom_field": "",
        "payment_firstname": "Jane",
        "payment_lastname": "Doe",
        "payment_company": "",
        "payment_address_1": "123 Main St",
        "payment_address_2": "Apt 4B",
        "payment_city": "Springfield",
        "payment_postcode": "12345",
        "payment_country": "USA",
        "payment_country_id": 223,
        "payment_zone": "IL",
        "payment_zone_id": 3613,
        "payment_address_format": "",
        "payment_custom_field": "",
        "payment_method": "Credit Card",
        "payment_code": "cc",
        "shipping_firstname": "Jane",
        "shipping_lastname": "Doe",
        "shipping_company": "",
        "shipping_address_1": "123 Main St",
        "shipping_address_2": "Apt 4B",
        "shipping_city": "Springfield",
        "shipping_postcode": "12345",
        "shipping_country": "USA",
        "shipping_country_id": 223,
        "shipping_zone": "IL",
        "shipping_zone_id": 3613,
        "shipping_address_format": "",
        "shipping_custom_field": "",
        "shipping_method": "Flat Shipping Rate",
        "shipping_code": "flat.flat",
        "comment": "Please deliver after 6 PM",
        "total": 149.9,
        "order_status_id": 2,
        "affiliate_id": 0,
        "commission": 0,
        "marketing_id": 0,
        "tracking": "",
        "language_id": 1,
        "currency_id": 1,
        "currency_code": "USD",
        "currency_value": 1,
        "ip": "203.0.113.10",
        "forwarded_ip": "",
        "user_agent": "Mozilla/5.0",
        "accept_language": "en-US,en;q=0.9",
        "date_added": "2025-03-01T10:15:30Z",
        "date_modified": "2025-03-01T12:05:00Z",
        "products": [
          {
            "discount_amount": 0,
            "discount_type": "",
            "ean": "",
            "isbn": "",
            "jan": "",
            "location": "",
            "model": "SKU-001",
            "mpn": "",
            "name": "Sample Product A",
            "order_id": 10234,
            "price": 49.95,
            "product_id": 5970,
            "product_uid": "02bc1ea8-70d3-11ef-b7f7-00155d018000",
            "quantity": 1,
            "reward": 0,
            "sku": "SKU-001",
            "tax": 0,
            "total": 49.95,
            "upc": "",
            "weight": 0
          }
        ],
        "totals": [
          { "code": "sub_total", "title": "Sub-Total", "value": "129.90" },
          { "code": "shipping",  "title": "Flat Shipping Rate", "value": "20.00" },
          { "code": "total",     "title": "Total", "value": "149.90" }
        ]
      }
    ],
    "success": true,
    "status_message": "Success",
    "timestamp": "2025-03-24T11:22:39Z"
  }
  ```

#### Get Products of Order
- Endpoint: `/api/v1/order/{orderId}/products`
- Method: `GET`
- Description: Retrieves products belonging to the given order.
- Response:
  ```json
  {
    "data": [
      {
        "discount_amount": 0,
        "discount_type": "",
        "ean": "",
        "isbn": "",
        "jan": "",
        "location": "",
        "model": "SKU-001",
        "mpn": "",
        "name": "Sample Product A",
        "order_id": 10234,
        "price": 49.95,
        "product_id": 5970,
        "product_uid": "02bc1ea8-70d3-11ef-b7f7-00155d018000",
        "quantity": 1,
        "reward": 0,
        "sku": "SKU-001",
        "tax": 0,
        "total": 49.95,
        "upc": "",
        "weight": 0
      }
    ],
    "success": true,
    "status_message": "Success",
    "timestamp": "2025-03-24T11:22:39Z"
  }
  ```

#### Change Order Status
- Endpoint: `/api/v1/order`
- Method: `POST`
- Description: Sets a new status and optional comment for one or more orders.
- Request Body:
  ```json
  {
    "data": [
      {
        "order_id": 10234,
        "order_status_id": 5,
        "comment": "Status updated by OCAPI"
      }
    ]
  }
  ```
- Response:
  ```json
  {
    "success": true,
    "status_message": "Success",
    "timestamp": "2025-03-24T11:22:39Z"
  }
  ```

#### Get Orders by Status
- Endpoint: `/api/v1/orders/{orderStatusId}`
- Method: `GET`
- Query Parameters:
  - `from_date` (optional, RFC3339). If omitted, defaulted to 30 days ago.
- Description: Returns a list of order IDs that are in the given status and were modified since `from_date`.
- Response:
  ```json
  {
    "data": [10234, 10235, 10250],
    "success": true,
    "status_message": "Success",
    "timestamp": "2025-03-24T11:22:39Z"
  }
  ```