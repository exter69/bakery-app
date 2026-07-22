---
inclusion: fileMatch
fileMatchPattern: ["**/controller/**/*.java", "**/api/**/*", "**/*Controller.java"]
---

# API Standards

- RESTful resource naming: plural nouns, kebab-case paths (`/api/v1/order-items`), no verbs in paths.
- Version the API in the path (`/api/v1/...`).
- HTTP status codes: 200/201/204 for success, 400 validation, 401/403 auth, 404 not found, 409 conflict, 500 unexpected. Never return 200 with an error body.
- Consistent error response shape:

```json
{
  "timestamp": "...",
  "status": 400,
  "error": "VALIDATION_ERROR",
  "message": "human readable",
  "details": []
}
```

- Pagination on all collection endpoints (`page`, `size`, sorted response metadata).
- Request/response bodies are DTOs with Bean Validation annotations; document with OpenAPI/springdoc annotations.
- Breaking changes require a new version, never silent changes to existing contracts.
