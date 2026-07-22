---
inclusion: always
---

# Security Policies

- NEVER hardcode secrets, API keys, passwords, or tokens. Use environment variables / Spring config with placeholders; add `.env` files to `.gitignore`.
- Validate and sanitize all external input at the boundary (Bean Validation on DTOs in Spring, zod or equivalent on the frontend/API layer).
- Every new backend endpoint must declare its authentication/authorization requirements explicitly — no accidentally-public endpoints.
- Use parameterized queries / JPA repositories only; never concatenate user input into queries.
- Don't log sensitive data (credentials, tokens, personal data).
- Keep dependencies current; flag any newly added dependency with known vulnerabilities instead of silently adding it.
- Frontend: never trust client-side checks alone; escape/encode rendered user content (React default escaping — avoid `dangerouslySetInnerHTML`).
