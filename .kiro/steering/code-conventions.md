---
inclusion: always
---

# Code Conventions

## General

- Prefer small, focused changes; don't refactor unrelated code in the same prompt.
- Match the existing style of the file you're editing over any general rule.
- No TODOs without a ticket/issue reference.
- Explain non-obvious decisions with a brief comment on the "why", not the "what".

## Spring Boot

- Layering: Controller → Service → Repository. No business logic in controllers.
- Use constructor injection (no `@Autowired` on fields).
- DTOs at the API boundary; never expose JPA entities directly.
- Centralized exception handling via `@ControllerAdvice`; return structured error responses.
- Use `Optional` returns from repositories; no null returns from services.

## TypeScript / React

- `strict` mode TypeScript; no `any` unless justified with a comment.
- Functional components + hooks only; no class components.
- Derive state where possible; lift state only when shared.
- Async data via a query library (e.g., TanStack Query) rather than hand-rolled useEffect fetching, if the project already uses one.
- Named exports preferred; default exports only for pages/routes if the framework requires it.

## Naming

- Java: `PascalCase` classes, `camelCase` methods/fields, packages by domain not by layer where feasible.
- TS/React: `PascalCase` components, `camelCase` functions/variables, `kebab-case` file names for non-components.
