# Mode - Image Generation and Management Service

## Overview

Mode is a Go-based web service that provides user authentication, image storage, and image generation capabilities. It uses PostgreSQL for data storage, S3 compatible services for image storage, and integrates with a separate gRPC-based image generation service.

## Features

- User authentication (signup, signin)
- Secure password hashing
- JWT-based authentication for protected routes
- Image storage and retrieval using S3
- Integration with a gRPC-based image generation service
- Logging middleware for request tracking

## Prerequisites

- Go 1.16+
- PostgreSQL
- AWS S3 account and credentials
- Protocol Buffers compiler (protoc)

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/alvarofc/mode.git
   cd mode
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Set up environment variables (create a `.env` file in the project root):
   ```
   DB_HOST=your_db_host
   DB_PORT=your_db_port
   DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_NAME=your_db_name
   JWT_SECRET=your_jwt_secret
   AWS_ACCESS_KEY_ID=your_aws_access_key
   AWS_SECRET_ACCESS_KEY=your_aws_secret_key
   AWS_REGION=your_aws_region
   S3_BUCKET_NAME=your_s3_bucket_name
   ```

4. Generate gRPC code:
   ```
   protoc --go_out=. --go-grpc_out=. proto/imagegen/v1/imagegen.proto
   ```

## Usage

1. Start the server:
   ```
   go run main.go
   ```

2. The server will start on `localhost:8080` (or the port specified in your configuration).

## API Endpoints

- `POST /signup`: Create a new user account
- `POST /signin`: Authenticate and receive a JWT token
- `GET /user`: Get user information (protected route)
- `GET /photo/{key}`: Retrieve a photo by its key (protected route)
- `GET /user/{user_id}/photos`: Get the last X photos for a user (protected route)
- `GET /user/{user_id}/photo`: Get the last photo for a user (protected route)
- `POST /generate-image`: Generate a new image (protected route, requires image generation service)

## Project Structure

- `api/`: Contains the main server logic and handlers
- `storage/`: Interfaces and implementations for data storage (PostgreSQL) and file storage (S3)
- `proto/`: Protocol Buffer definitions for the image generation service
- `types/`: Common type definitions used across the project

