# JSON Server

A simple and fast JSON storage server with a web UI, built in Go using FastHTTP.

## Features

- **Fast and Lightweight**: Built with FastHTTP for high performance
- **Simple API**: Easy-to-use REST endpoints for storing and retrieving JSON data
- **Web UI**: Clean, dark-themed interface for managing JSON data
- **Password Protection**: Secure POST endpoints with password authentication
- **CORS Support**: Built-in CORS support for domains ending with "noxchat.in"
- **SQLite Storage**: Persistent storage with in-memory caching

## API Endpoints

### GET /{key}
Retrieves JSON data for the specified key.

- **Response**: 
  - 200: JSON data
  - 404: Key not found
  - 500: Server error

### POST /{key}
Updates or creates JSON data for the specified key.

- **Headers**:
  - `Authorization`: Required password
  - `Content-Type`: application/json
- **Response**:
  - 200: Success
  - 400: Invalid JSON or key format
  - 403: Invalid password
  - 413: Request body too large
  - 500: Server error

### GET /ui?key={key}&pass={password}
Web interface for viewing and editing JSON data.

## Setup

1. Create a `.env` file:
```env
PORT=8080
PASSWORD=your_secure_password
```

2. Run the server:
```sh
go run .
```


## Key Format Rules

- Maximum length: 256 characters
- Allowed characters:
  - Lowercase letters (a-z)
  - Uppercase letters (A-Z)
  - Numbers (0-9)
  - Special characters: `-`, `_`, `.`

## Technical Details

- **Storage**: SQLite with in-memory caching
- **Max Request Size**: 1MB

## Security Considerations

- POST requests require password authentication
- JSON validation before storage
- Input sanitization for keys
- CORS restrictions

[All code and documentation written by Claude]

[All code is written by Claude, including this README]
