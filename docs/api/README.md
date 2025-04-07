# API Documentation

## Overview

This document provides detailed API specifications for all microservices in the e-commerce platform.

## Product Service API

### List Products
```http
GET /api/v1/products
```

Response:
```json
{
  "products": [
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "price": number,
      "category": "string",
      "created_at": "ISO8601 datetime",
      "updated_at": "ISO8601 datetime"
    }
  ]
}
```

### Get Product
```http
GET /api/v1/products/{id}
```

Response:
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "price": number,
  "category": "string",
  "created_at": "ISO8601 datetime",
  "updated_at": "ISO8601 datetime"
}
```

### Create Product
```http
POST /api/v1/products
```

Request:
```json
{
  "name": "string",
  "description": "string",
  "price": number,
  "category": "string"
}
```

Response: 201 Created with product object

### Create Product (Async)
```http
POST /api/v1/products/async
```

Request: Same as Create Product
Response: 202 Accepted with event ID

## Order Service API

### List Orders
```http
GET /api/v1/orders
```

Response:
```json
{
  "orders": [
    {
      "id": "string",
      "user_id": "string",
      "total_amount": number,
      "status": "string",
      "created_at": "ISO8601 datetime",
      "updated_at": "ISO8601 datetime",
      "items": [
        {
          "id": "string",
          "product_id": "string",
          "quantity": integer,
          "price": number
        }
      ]
    }
  ]
}
```

### Get Order
```http
GET /api/v1/orders/{id}
```

Response: Same as order object in List Orders

### Create Order
```http
POST /api/v1/orders
```

Request:
```json
{
  "user_id": "string",
  "items": [
    {
      "product_id": "string",
      "quantity": integer
    }
  ]
}
```

Response: 201 Created with order object

### Create Order (Async)
```http
POST /api/v1/orders/async
```

Request: Same as Create Order
Response: 202 Accepted with event ID

## Error Responses

All endpoints may return the following error responses:

```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": object
  }
}
```

Common HTTP status codes:
- 400 Bad Request
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- 500 Internal Server Error 