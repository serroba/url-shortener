# URL Shortener

[![CI/CD Pipeline](https://github.com/serroba/web-demo-go/actions/workflows/cy.yml/badge.svg)](https://github.com/serroba/web-demo-go/actions/workflows/cy.yml)
[![codecov](https://codecov.io/gh/serroba/web-demo-go/branch/main/graph/badge.svg)](https://codecov.io/gh/serroba/web-demo-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/serroba/web-demo-go)](https://goreportcard.com/report/github.com/serroba/web-demo-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/serroba/web-demo-go.svg)](https://pkg.go.dev/github.com/serroba/web-demo-go)

A URL shortening service with two strategies for generating short codes.

## API

### Create Short URL

```
POST /shorten
```

**Request:**
```json
{
  "url": "https://example.com/very/long/path",
  "strategy": "token"
}
```

**Strategies:**
- `token` (default): Generates a unique short code for every request
- `hash`: Returns the same short code for identical URLs (deduplication)

**Response:**
```json
{
  "code": "abc123",
  "shortUrl": "http://localhost:8888/abc123",
  "originalUrl": "https://example.com/very/long/path"
}
```

### Redirect

```
GET /{code}
```

Redirects (301) to the original URL.
