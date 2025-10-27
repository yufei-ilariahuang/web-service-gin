#### generate api.go byO PENAPI

# Product API

This is a simple RESTful API for managing a catalog of products. It allows for retrieving products, creating new ones, and deleting existing ones. The API is designed to simulate a realistic memory footprint by storing a large number of products in memory and uses a fixed-time computation approach for searches.

## Features

*   **Product Management:**
    *   `GET /products`: Retrieve a list of products. Supports filtering by brand and category.
    *   `POST /products`: Create a new product.
    *   `GET /products/{id}`: Retrieve a specific product by its unique ID.
    *   `DELETE /products/{id}`: Delete a specific product by its unique ID.
*   **Data Generation:**
    *   Starts by generating 100,000 products in memory upon startup.
    *   Product names follow the pattern "Product [AppName] [ID]".
    *   Brands are randomly selected from a predefined list: `["Apple", "Samsung", "Google", "Microsoft", "Sony", "Logitech"]`.
    *   Categories are randomly selected from a predefined list: `["Electronics", "Books", "Home", "Clothing", "Groceries", "Toys"]`.
    *   Descriptions are generated using `gofakeit`.
    *   Prices are random floating-point numbers.
*   **Search Logic:**
    *   **Fixed Computation Time:** Each search operation checks a maximum of 100 products, regardless of the total number of products available. This simulates a fixed-time computation for search operations.
    *   **Filtering:** Searches support case-insensitive partial matching for `brands` and case-insensitive exact matching for `categories`.
    *   **Result Limiting:** Returns a maximum of 20 matching products per search request.
*   **Response Format:**
    *   All responses are in JSON format.
    *   Search results include:
        *   `products`: An array of matching product objects (max 20).
        *   `total_found`: The total count of products that matched the search criteria (within the searched limit).
        *   `search_time`: The duration taken for the search operation (optional, but included for performance insights).

## Getting Started

### Running the API
```bash
oapi-codegen \
  --package api \
  -o src/api/server.gen.go \
  --generate types,server \
  ecommerce.yaml
```
* --package api:sets the package name.
* -o src/api/server.gen.go: The -o flag takes the full path and filename where the generated code should be saved.
* --generate types,server:  Echo Framework
* ecommerce.yaml: OpenAPI specification file.
```bash
docker build -t api .
docker run -p 8080:8080 api   
```


## API Endpoints

All API requests should be made to `http://localhost:8080`.

### `GET /products`

Retrieves a list of products.

**Query Parameters:**

*   `brands` (string, optional): Filter products by brand name (case-insensitive, partial match).
    *   Example: `/products?brands=App`
*   `categories` (string, optional): Filter products by category name (case-insensitive, exact match).
    *   Example: `/products?categories=Electronics`

**Example Request:**

```bash
curl "http://localhost:8080/products?brands=Sam&categories=Home"
```

**Example Response:**

```json
{
  "products": [
    {
      "id": "5",
      "name": "Product Go 5",
      "category": "Home",
      "brand": "Samsung",
      "description": "Reprehenderit ut aut et ut voluptatem.",
      "price": 789.12
    }
    // ... up to 19 more products
  ],
  "total_found": 12,
  "search_time": "1.2ms"
}
```

### `POST /products`

Creates a new product.

**Request Body:**

The request body should be a JSON object representing the `ProductInput` structure.

```json
{
  "name": "New Gadget Pro",
  "category": "Electronics",
  "price": 999.99,
  "brand": "TechCo",
  "description": "The latest innovation in tech."
}
```

**Example Request:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "name": "New Gadget Pro",
  "category": "Electronics",
  "price": 999.99,
  "brand": "TechCo",
  "description": "The latest innovation in tech."
}' http://localhost:8080/products
```

**Example Response (Status Code: 201 Created):**

```json
{
  "id": "100001",
  "name": "New Gadget Pro",
  "category": "Electronics",
  "brand": "TechCo",
  "description": "The latest innovation in tech.",
  "price": 999.99
}
```

### `GET /products/{id}`

Retrieves a single product by its ID.

**Path Parameters:**

*   `id` (string, required): The unique identifier of the product.

**Example Request:**

```bash
curl http://localhost:8080/products/1
```

**Example Response (Status Code: 200 OK):**

```json
{
  "id": "1",
  "name": "Product Apple 1",
  "category": "Electronics",
  "brand": "Apple",
  "description": "Nisi rerum ad.",
  "price": 123.45
}
```

If the product is not found, a `404 Not Found` status will be returned.

### `DELETE /products/{id}`

Deletes a single product by its ID.

**Path Parameters:**

*   `id` (string, required): The unique identifier of the product to delete.

**Example Request:**

```bash
curl -X DELETE http://localhost:8080/products/1
```

**Example Response (Status Code: 204 No Content):**

A successful deletion will return a `204 No Content` status. If the product is not found, a `404 Not Found` status will be returned.

