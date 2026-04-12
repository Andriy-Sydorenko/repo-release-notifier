# 📩 GitHub Releases Subscription API (Test Task)

## 📌 Overview

Implement an API service that allows users to subscribe to **email notifications** about new releases of a selected GitHub repository.

The service must periodically check for new releases and notify subscribers when a new release appears.

---

## ⚙️ Technical Requirements

### 1. API Contract (Swagger)

- The service **must strictly follow the provided Swagger API specification**
- You are **NOT allowed to modify API contracts**
- You can use:
    - https://editor.swagger.io/ for visualization and testing

---

### 2. Architecture

- The entire system must be implemented as a **single service (monolith)**
- The service must include:
    - API layer
    - Scanner (release checker)
    - Notifier (email sender)

> ❌ Microservices are **NOT allowed**

---

### 3. Data Storage

- All application data must be stored in a **database**
- You must implement:
    - Database schema
    - **Migrations that run on service startup**

---

### 4. Dockerization

- The repository must include:
    - `Dockerfile`
    - `docker-compose.yml`
- The whole system must be runnable via Docker

---

### 5. Release Monitoring Logic

- The service must periodically check for new releases for **all active subscriptions**
- For each repository:
    - Store `last_seen_tag`
    - Send notifications **ONLY if a new release appears**

---

### 6. Subscription Validation

When creating a subscription:

- Validate repository via **GitHub API**
- Expected format:
  ```
  owner/repo
  ```
  Example:
  ```
  golang/go
  ```

- Return:
    - `400` → invalid format
    - `404` → repository not found

---

### 7. External API Error Handling

- Handle GitHub API rate limiting:
    - `429 Too Many Requests`
- Rate limits:
    - Without token → 60 requests/hour
    - With token → 5000 requests/hour

---

### 8. Framework Restrictions

You may use **only lightweight frameworks**

#### Allowed:

- **Go**
    - Gin
    - Chi
    - net/http

- **Node.js**
    - Express
    - Fastify

- **PHP**
    - Slim
    - Built-in capabilities

#### Forbidden:

- Nest.js (Node.js)
- Revel / Fx (Go)
- Laravel (PHP)

---

### 9. Testing

- ✅ Unit tests for business logic — **REQUIRED**
- ➕ Integration tests — optional (bonus)

---

### 10. Documentation

You may include:

- Implementation notes
- Architecture decisions
- Trade-offs

In:

```
README.md
```

> Good explanations can compensate for incomplete implementation

---

## 🚀 Extra (Bonus Points)

You can earn additional points by implementing:

- Deployment:
    - Hosted API
    - HTML page for subscription

- gRPC interface:
    - As alternative or addition to REST API

- Redis caching:
    - Cache GitHub API responses
    - TTL: **10 minutes**

- API Key Authentication:
    - Protect endpoints using token in headers

- Prometheus metrics:
    - `/metrics` endpoint
    - Basic service metrics

- CI Pipeline:
    - GitHub Actions
    - Run linter + tests on every push

## 🛠 Implementation Note

This solution will be implemented in **Golang**
