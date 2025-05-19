# Document Administration System

A knowledge library management system for organizing, storing, and managing documents in a hierarchical structure.

## Overview

This application provides a backend service for managing document libraries, with the following features:

- Create and manage multiple knowledge libraries
- Organize documents in a tree-like hierarchical structure
- Upload and manage images associated with documents
- RESTful API for document operations

## Project Structure

```
doc_admin/
├── config.yaml         # Configuration file
├── db/                 # Database directory
├── handlers/           # HTTP request handlers
│   ├── document.go     # Document management handlers
│   ├── library.go      # Library management handlers
│   └── upload.go       # File upload handlers
├── models/             # Data models
│   └── document.go     # Document model
├── router/             # API routing
│   └── router.go       # Router setup
├── storage/            # Document storage directory
└── main.go             # Application entry point
```

## Setup and Configuration

1. Ensure you have Go installed (version 1.16+ recommended)
2. Configure the document root directory in `config.yaml`
3. Run the application:

```bash
go run main.go
```

The server will start on port 8080 by default.

## API Endpoints

### Library Management

- `POST /library/create` - Create a new knowledge library
  - Request body: `{"name": "library_name", "base_path": "./storage"}`

### Document Management

- `POST /document` - Create a new document
  - Request body: `{"title": "Document Title", "content": "Document content", "parent_id": 0}`
- `GET /document/tree` - Get the document tree structure
- `POST /document/update-parent` - Update a document's parent
  - Request body: `{"id": 1, "parent_id": 2}`

### File Management

- `POST /upload/:id` - Upload an image for a document
  - Form data: `file` - The image file to upload

## Database Structure

The application uses SQLite for data storage. Each knowledge library has its own database file with the following structure:

### Documents Table

| Column    | Type    | Description                   |
|-----------|---------|-------------------------------|
| id        | INTEGER | Primary key                   |
| title     | TEXT    | Document title                |
| content   | TEXT    | Document content              |
| parent_id | INTEGER | Parent document ID (for tree) |

## Storage Structure

Documents and associated files are stored in the configured document root directory:

```
storage/
├── library_name1/
│   ├── blog.db        # SQLite database
│   └── pic/           # Image storage
│       └── doc_id/    # Images for specific document
├── library_name2/
    ├── blog.db
    └── pic/
```

## License

[MIT License](LICENSE)
