# OCAPI Project

## Overview
OCAPI (OpenCart API) is a project designed to provide a robust and flexible API for the OpenCart site database. It allows users to update site products and retrieve orders seamlessly.

## Features
- Product management
- Order retrieval (coming soon)

## Prerequisites
- Go 1.16 or higher
- Installed OpenCart site

The project includes GitHub actions scripts for CI/CD, you can use as a template for your environment.

## Setup

### Clone the repository
Clone repository to your local machine or server, build application, and deploy it alongside your OpenCart site.

### Configure the database
Update the `config.yaml` file with your OpenCart database credentials. Provide port, on which the OCAPI service will run, and the API key.

### Run the application
Run the application manually or set up a service to run it in the background.

## API Description

To make requests to the API, you need to provide the API key in the `Authorization` header as a Bearer token. All methods return a JSON response with success status, status message, and timestamp. If data fails validation, the response will include an error message.

### Product Management

#### Update or Create Product
- **Endpoint:** `/product`
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
            "category_uid": "6666bc6a-a487-11e9-b6d3-00155d010d00"        }
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
- **Endpoint:** `/product/description`
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
- **Endpoint:** `/category`
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
- **Endpoint:** `/category/description`
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
- **Endpoint:** `/product/{uid}`
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
            "model": "02bc1ea8-70d3-11ef-b7f7-00155d018000",
            "mpn": "",
            "points": 0,
            "price": "0.0000",
            "product_id": 5970,
            "quantity": 354,
            "sku": "doilon3",
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

### Order Retrieval (Coming Soon)

## License
This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact
For any questions or inquiries, please contact developer at [dev@nomadus.net].

`