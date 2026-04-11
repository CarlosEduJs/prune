---
title: AI TESTS
description: Guidelines for generate tests with AI
---

## Write high-quality tests for the provided Go code.

## Context:
- Language: Go
- Testing: standard library ("testing" package)
- Style: idiomatic Go (table-driven tests when appropriate)

## Rules:
- Test only exported/public behavior when possible
- Do NOT test implementation details
- Avoid unnecessary mocks (prefer real structs and simple fakes)
- Use table-driven tests for multiple scenarios
- Cover:
  - happy path
  - edge cases
  - error cases
- Tests must be deterministic (no randomness, no timing issues)
- Use clear and descriptive test names
- Keep tests simple and readable

## Structure:
- Use subtests (t.Run) where it improves clarity
- Reuse setup code when appropriate
- Avoid duplication

## Important:
- Do not assume behavior not present in the code
- If something is hard to test, suggest a small refactor as a comment