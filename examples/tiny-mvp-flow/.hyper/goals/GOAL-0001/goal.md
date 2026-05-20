# GOAL-0001 Runtime Packet

## Continue From

No similar Hyper Run context found yet.

## Current Episode

Build the smallest local task list MVP.

## Why Now

- Product: Pocket Tasks, a local-first browser task list for a solo builder.
- Stage: Tiny MVP
- Target users: Solo builders who need a tiny task board while shaping an MVP.
- This packet exists to continue from observed project state, not to freeze a long-lived SPEC.

## Runtime Inputs

- Build style: Web app
- Current focus: Build the smallest local task list MVP.

## Stage Gate

- Current gate: Tiny MVP -> Usable MVP
- Gate status: not_ready
- Next readiness pressure: Core UX (emerging)
- Pressure reason: Core UX is emerging for the Tiny MVP -> Usable MVP gate.
- Gate gap: Core UX: The primary user flow is not yet proven usable.
- Gate gap: Validation coverage: The primary behavior does not have repeatable validation evidence.
- Gate evidence: Product and MVP slice are measurable.
- Gate evidence: One core user flow works locally.
- Gate evidence: Minimal validation evidence exists.

## Work Boundary

- Do the smallest coherent implementation step that advances: Build the smallest local task list MVP.
- Keep the work inside the current stage: Tiny MVP
- Use the product MVP brief as the boundary: Users can add a task, mark it complete, delete it, and keep tasks after a browser reload.
- Avoid plan non-goals: No auth, backend, sync, collaboration, analytics, or paid services.
- Respect constraints: Use local storage only. Keep the UI simple enough to validate in one browser smoke test.
- Prioritize the smallest service-readiness step for Core UX: The primary user flow is not yet proven usable.

## Validation Signals

- Detect and run the safest available build, test, or smoke command.
- If a browser UI exists, capture screenshot evidence and check console errors.
- If a dev server is required, document the URL and verification steps.
- Capture readiness evidence for Core UX in evidence.md.

## Evidence Required

- Command output or reason validation could not run
- Readiness evidence for the selected pressure
- Active capability evidence when required validators are present
- Changed file summary
- Decisions that should persist into future runs
- Reusable patterns that should guide similar future work
- Blocker, constraint, or failure signal when applicable
- Screenshot path when applicable

## Stop When

- One core user flow works locally.
- The project can be run from documented commands.
- Minimal validation has passed or blockers are documented.
- evidence.md and next.md are updated.
