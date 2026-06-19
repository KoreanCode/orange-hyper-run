# Product Plan

## Product

Hyper Run is an AI-delegated project growth runtime. It lets an AI agent take broad ownership of a software project: research what it needs, organize findings, choose the next smallest useful step, implement it, validate it, learn from the result, and repeat until the project reaches service quality.

The product exists for builders who do not want a project to become mediocre just because human attention is inconsistent. Hyper Run should preserve direction, evidence, validation, decisions, and next actions inside the repository so the AI can keep moving with minimal human intervention.

## Target Users

Individual builders and small teams using Codex Desktop, CLI agents, or similar coding assistants to grow projects over many sessions.

The primary user wants to delegate execution control to AI while still keeping enough local evidence, safety boundaries, and review artifacts to understand what happened.

## MVP

The smallest useful version is a repository-local loop that can:

- Read a human-owned product brief.
- Generate one small runtime packet.
- Let the AI execute that packet.
- Require concrete validation or a real blocker.
- Record evidence, decisions, reusable patterns, and next work.
- Learn durable signals from completed packets.
- Continue toward a target stage without the human restating context every turn.

## Current Stage

Sustained Service Quality

## Target Stage

Sustained Service Quality



## Build Style

Native Go CLI with project-local files, SQLite-backed event storage, Codex Desktop routing, and generated runtime packets.

## Non-goals

- Do not become a generic project management app.
- Do not require users to design a harness before the project has repeated evidence that it needs one.
- Do not make static skill files the source of truth.
- Do not silently perform destructive, credential-sensitive, or external-cost actions without an explicit safety boundary.
- Do not optimize for endless activity when validation shows no progress.

## Constraints

- User intervention should be minimized, but not removed for irreversible, destructive, credential, payment, publication, or high-risk operations.
- Human control should sit at policy boundaries: approval, credentials, product ownership, spending, publication, destructive actions, and scope changes. Humans should not need to micromanage ordinary task selection or validation execution.
- The AI must work in small coherent loops so progress is inspectable.
- Every loop needs evidence: command output, smoke proof, browser proof, artifact proof, benchmark proof, or a concrete blocker.
- Repeatable command validation should prefer `hyper verify -- <command>` so exit code, log hashes, commit SHA, worktree status hash, run ID, and goal ID are machine-recorded. Markdown evidence should summarize or cite those records, not replace them.
- Harnesses, validators, skills, agents, and stricter workflows should be generated only after repeated pressure proves they are useful.
- Project knowledge must live in `plan.md`, `.hyper/`, logs, evidence, and generated candidates, not in transient chat memory.
- Auto continuation must include progress guards so repeated non-progress does not look like work.

## Success Criteria

Hyper Run reaches the intended product quality when a user can start with a broad goal, set a target stage, and let AI repeatedly:

- Research missing product, technical, validation, and benchmark context when needed.
- Convert research into a concise local decision or constraint.
- Pick the next smallest coherent implementation step.
- Implement and validate that step.
- Create or promote a project-specific validator, harness, skill, or agent only after repeated evidence supports it.
- Stop only for real safety boundaries, missing credentials, repeated validation failure, unclear product ownership, or target-stage completion.
- Produce enough evidence that a later AI or human can understand what changed, what passed, what failed, and why the next step was chosen.

Service Quality specifically requires repeatable validation, clear setup/update/release paths, operational recovery notes, security and privacy boundaries, reference benchmark evidence, product satisfaction review, and sustained progress without relying on hidden chat context.

## Current Focus

Operate Hyper Run as an autonomous service-quality loop: the user should normally run only `hyper run`, while the AI reads the runtime packet, performs the smallest coherent step, validates it, writes evidence and next notes, runs the finish gate internally, and continues only when the next packet is safe, concrete, and evidence-producing.

## Autonomous Service Quality Loop

The loop exists to keep project direction intact while minimizing human attention. It should prefer project quality over activity volume.

1. Human entrypoint
   - The normal user action is `hyper run` or the equivalent Codex `$hyper-run` route.
   - The user is not expected to run `hyper complete` in the ordinary flow.
   - The user is asked only for approval-required actions, missing credentials/environments, ambiguous product ownership, or a deliberate scope change.

2. Agent responsibility
   - Read `plan.md`, the generated runtime packet, recent evidence, active capabilities, and `.hyper/next-packet.md`.
   - Classify safety before action.
   - Choose one smallest coherent episode.
   - Run active validators through `hyper verify -- <command>` when possible, or record a concrete reason they cannot run.
   - Write evidence, next notes, durable Learn signals, and self review.
   - Run the finish gate internally before starting another packet.

3. Continue automatically when all are true
   - `.hyper/next-packet.md` says `Action: run`.
   - The next command is concrete and scoped to one episode.
   - No destructive, credential-sensitive, publication, deployment, external-cost, production-data, or environment-changing action is required.
   - The previous packet passed the finish gate.
   - The next packet can produce code, validation evidence, readiness evidence, an active capability signal, a clearer blocker, or a changed next step.

4. Stop when any are true
   - Approval is needed for install/update, tag, push, release, deployment, branch deletion, credentials, external spend, production data, or similar high-risk action.
   - A required environment is missing, such as Windows for Windows installer smoke or `cosign` for sigstore verification.
   - Validation fails twice for the same reason.
   - The next recommendation repeats without new evidence or a changed boundary.
   - Product scope, target user, or current stage is unclear.
   - `.hyper/next-packet.md` says `Action: stop`, or `Action: advance` without authorized stage advancement.

5. Service-quality evidence
   - Functional proof: active Go validators, targeted tests, command smoke, or artifact proof. Repeatable command proof should be backed by Verified Evidence records under `.hyper/verified-evidence/`.
   - Operational proof: install/update/release/checksum/signature/rollback/setup evidence when those surfaces are touched.
   - Core UX proof: `hyper run`, `status`, `doctor`, and generated packet guidance keep users on the intended flow.
   - Security proof: secrets are not exposed; release artifacts have checksum proof and signature proof when tooling exists.
   - Maintainability proof: stale branches, dirty state, repeated friction, and unclear handoffs are closed or routed to the next packet.
   - Product satisfaction proof: the result remains useful, coherent, and aligned with delegated autonomy, not merely test-passing.
   - Verified Evidence proof: `hyper verify` records become the source of truth for command execution metadata; `evidence.md` remains the human-readable summary and decision ledger.

6. Validator and harness promotion
   - Active validators are required until evidence says otherwise: `GOCACHE=/private/tmp/hyper-go-cache go test ./...`, `go test ./...`, and `git diff --check`.
   - Promotable validators such as `hyper status --short`, fresh `init/run/status/doctor`, and stale wording guards should become active only after the activation threshold or explicit maintainer acceptance.
   - A release-verification helper should wait until Windows smoke and sigstore-tooling decisions add one more stable pressure cycle.
   - Do not create a harness from a single packet or from planning alone.

7. Near-term service-quality order
   - First, publish the autonomous-loop runtime fix as the next patch release using current-environment validation and release checks.
   - Defer Windows installer smoke until a Windows-capable environment is available.
   - Next, run optional `cosign` verification if installation is approved.
   - Then promote repeated first-run/status validators only if another independent packet confirms the pressure.
   - After that, audit operations and recovery notes for setup, update, rollback, and failed-packet recovery.
   - Only after repeated release verification pressure should a project-owned release verification helper be created.
