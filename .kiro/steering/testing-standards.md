---
inclusion: fileMatch
fileMatchPattern: ["**/*.java", "**/*.kt", "**/*.ts", "**/*.tsx"]
---

# Testing Standards

Every new method, class, component, or endpoint MUST ship with tests in the same prompt. No exceptions unless explicitly told to skip.

## General rules

- Follow Arrange–Act–Assert (Given–When–Then) structure.
- One behavior per test; name tests after the behavior, not the method (`returnsEmptyListWhenNoOrdersExist`, not `testGetOrders`).
- Cover the happy path, at least one error path, and boundary cases (null, empty, invalid input).
- When fixing a bug, first write a failing test reproducing it.
- Run the relevant test suite before finishing and report results.

## Spring Boot (Java)

- Framework: JUnit 5 + Mockito + AssertJ.
- Unit tests for services/domain logic — mock repositories and external clients.
- `@WebMvcTest` for controllers, `@DataJpaTest` for repositories; reserve `@SpringBootTest` for true integration tests.
- Test location mirrors source: `src/test/java/...` matching the package of the class under test.
- Never call real external services; use mocks or Testcontainers.

## TypeScript / React

- Framework: Vitest (or Jest) + React Testing Library.
- Test files co-located: `Component.test.tsx` / `module.test.ts` next to the source.
- Test components through user-visible behavior (queries by role/text), not implementation details or snapshots.
- Mock network calls with MSW or module mocks; never hit real APIs.
- For hooks and utilities, plain unit tests with edge cases.

## Coverage expectation

New code should not lower overall coverage. Aim for meaningful assertions over coverage numbers — no assertion-free tests.
