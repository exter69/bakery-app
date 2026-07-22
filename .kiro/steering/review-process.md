---
inclusion: always
---

# Mandatory Review Pass

At the END of every prompt that changed code, perform a review pass BEFORE declaring the task done. Act as a separate, critical reviewer — assume the code is wrong until proven otherwise.

## Review checklist

1. **Correctness**: re-read every changed file. Does the code do what was asked? Any edge cases missed (null/empty inputs, error paths, concurrency)?
2. **Consistency**: does it follow existing patterns in this codebase (naming, layering, error handling)?
3. **Tests**: does every new/changed method or class have tests? Do they pass? Run them.
4. **Security**: no secrets committed, inputs validated, no injection risks, authz checks in place on new endpoints.
5. **Regressions**: could this break existing callers? Check usages of modified signatures.
6. **Cleanup**: no dead code, debug logs, commented-out blocks, or unused imports left behind.

## Output

End the response with a short **Review Summary**:

- Issues found and fixed during review
- Remaining risks or follow-ups (or "none")
- Test results (pass/fail counts)

If review finds problems, fix them before finishing — do not report issues without fixing them unless they are out of scope.
