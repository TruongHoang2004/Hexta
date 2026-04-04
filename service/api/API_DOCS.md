# CommerceHub API Documentation

This document provides a comprehensive overview of the available API endpoints for the CommerceHub Backend (API Service). All endpoints are prefixed with `/api/v1`.

## Authentication

### Register
Create a new user account.

- **URL:** `/auth/register`
- **Method:** `POST`
- **Auth required:** No
- **Request Body:**
```json
{
  "user_name": "johndoe",
  "full_name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword123",
  "gender": "male", // "male", "female", "other"
  "phone": "+1234567890",
  "date_of_birth": "1990-01-01T00:00:00Z"
}
```
- **Success Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "id": "uuid-string",
    "user_name": "johndoe",
    "full_name": "John Doe",
    "email": "john@example.com",
    "gender": "male",
    "phone": "+1234567890",
    "date_of_birth": "1990-01-01T00:00:00Z",
    "created_at": "2024-04-04T12:00:00Z",
    "updated_at": "2024-04-04T12:00:00Z"
  }
}
```

### Login
Authenticate a user and receive access/refresh tokens.

- **URL:** `/auth/login`
- **Method:** `POST`
- **Auth required:** No
- **Request Body:**
```json
{
  "user_name": "johndoe",
  "password": "securepassword123"
}
```
- **Success Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "access_token": "jwt-access-token",
    "refresh_token": "jwt-refresh-token"
  }
}
```

### Validate Token
Checks if the provided access token is valid.

- **URL:** `/auth/validate-token`
- **Method:** `GET`
- **Headers:** `Authorization: Bearer <access_token>`
- **Auth required:** Yes
- **Success Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "valid": true,
    "user_id": "uuid-string"
  }
}
```

### Refresh Token
Obtain a new access token using a refresh token.

- **URL:** `/auth/refresh`
- **Method:** `POST`
- **Auth required:** No
- **Request Body:**
```json
{
  "refresh_token": "jwt-refresh-token"
}
```
- **Success Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "access_token": "new-jwt-access-token",
    "refresh_token": "new-jwt-refresh-token"
  }
}
```

---

## User Management

### Get Current User Profile
Retrieve the profile of the currently authenticated user.

- **URL:** `/users/me`
- **Method:** `GET`
- **Headers:** `Authorization: Bearer <access_token>`
- **Auth required:** Yes
- **Success Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "id": "uuid-string",
    "user_name": "johndoe",
    "full_name": "John Doe",
    "email": "john@example.com",
    "gender": "male",
    "phone": "+1234567890",
    "date_of_birth": "1990-01-01T00:00:00Z"
  }
}
```

### List Users
Retrieve a paginated list of users (Admin/Merchant only).

- **URL:** `/users/`
- **Method:** `GET`
- **Headers:** `Authorization: Bearer <access_token>`
- **Auth required:** Yes
- **Query Parameters:**
  - `page`: Page number (default: 1)
  - `size`: Page size (default: 20, max: 100)
  - `search`: Global search string
  - `user_name`: Filter by username
  - `full_name`: Filter by full name
  - `email`: Filter by email
  - `gender`: Filter by gender (male/female/other)
- **Success Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "data": [...],
    "total": 100,
    "page": 1,
    "size": 20,
    "total_pages": 5
  }
}
```

---

## Merchant Management

### Create Merchant
Register a new merchant profile for the authenticated user.

- **URL:** `/merchants/`
- **Method:** `POST`
- **Headers:** `Authorization: Bearer <access_token>`
- **Auth required:** Yes
- **Request Body:**
```json
{
  "name": "My Store",
  "logo": "https://example.com/logo.png",
  "description": "Store description",
  "phone": "+1234567890",
  "email": "store@example.com"
}
```
- **Success Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "id": "uuid-string",
    "name": "My Store",
    "status": "pending",
    "created_at": "...",
    ...
  }
}
```

### List Merchants
Retrieve a paginated list of merchants.

- **URL:** `/merchants/`
- **Method:** `GET`
- **Headers:** `Authorization: Bearer <access_token>`
- **Auth required:** Yes
- **Query Parameters:**
  - `page`, `size`, `search`
  - `status`, `name`, `email`
- **Success Response (200 OK):**
```json
{
  "code": "SUCCESS",
  "message": "Success",
  "data": {
    "data": [...],
    "total": 50,
    "page": 1,
    "size": 20,
    "total_pages": 3
  }
}
```

### Update Merchant
Update existing merchant information.

- **URL:** `/merchants/`
- **Method:** `PUT`
- **Headers:** `Authorization: Bearer <access_token>`
- **Auth required:** Yes
- **Request Body:**
```json
{
  "id": "uuid-string",
  "name": "Updated Store Name",
  "logo": "...",
  "description": "...",
  "phone": "...",
  "email": "..."
}
```

---

## System

### Ping
Check if the API is active.

- **URL:** `/ping`
- **Method:** `GET`
- **Auth required:** No
- **Response:** `{"message": "pong"}`
