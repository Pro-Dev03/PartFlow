---
description: "Investigate a bug or change request in PartFlow, fix the root cause, and verify the result with the smallest relevant check."
name: "Fix and Verify"
argument-hint: "Describe the bug, failing behavior, or requested change"
agent: "agent"
---

Investigate and resolve the issue described below in this repository.

Requirements:
- Start with one targeted search or symbol lookup to narrow to the likely implementation area.
- Identify the root cause before making changes.
- Prefer the smallest fix that addresses the actual defect without unrelated cleanup.
- If the behavior is currently untested, add a focused regression test or validation check.
- Keep the change consistent with the project conventions for the Go backend and the frontend TypeScript app.
- Validate with the smallest relevant command and report the exact result.
- If the issue crosses both backend and frontend layers, verify both affected surfaces.

Context:
- Backend services live under the backend folder.
- Frontend code lives under the frontend folder.
- This repo includes Go, TypeScript, Vite, and Docker-based tooling.
- Keep fixes pragmatic and production-safe.

Desired output:
1. Root cause summary
2. What changed
3. Files touched
4. Verification command(s) and result
5. Risks, edge cases, or follow-up actions

Issue or task:
<describe the bug, regression, or requested change here>
