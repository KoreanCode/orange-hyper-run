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
- The AI must work in small coherent loops so progress is inspectable.
- Every loop needs evidence: command output, smoke proof, browser proof, artifact proof, benchmark proof, or a concrete blocker.
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

Before executing the next six product steps, add an auditable decision hierarchy that helps the AI choose work without exposing hidden chain-of-thought: safety boundary, product intent, evidence gap, smallest step, validation proof, learning/promotion signal. Then proceed step by step through autonomous packet format, safety policy, validator/harness expansion, research evidence, loop guard, and product satisfaction.
