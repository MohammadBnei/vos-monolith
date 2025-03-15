# API Documentation

This directory contains the OpenAPI/Swagger specifications for the Voc on Steroid API.

## Overview

The API provides endpoints for:

- Searching for words
- Retrieving recent words
- Health checking

## Swagger UI

When the application is running in development mode (LogLevel=debug), you can access the Swagger UI at:

```
http://localhost:8080/swagger/index.html
```

Note: Swagger UI is only available in development mode for security reasons.

## Swagger YAML

The raw Swagger YAML file is available at:

```
http://localhost:8080/swagger.yaml
```

## API Endpoints

### Word Search

```
POST /api/v1/words/search
```

Search for a word by text and language.

### Recent Words

```
GET /api/v1/words/recent
```

Retrieve recently searched words.

### Health Check

```
GET /health
```

Check if the API is running.

## Authentication

Currently, the API does not require authentication.

## Rate Limiting

There are no rate limits currently implemented.
