# Release Checklist

Use this checklist when publishing a new Hyper Run release.

For AI-assisted release preparation, stop before replacing the active `hyper` executable, creating tags, pushing tags, or publishing GitHub releases unless the maintainer explicitly approves that action.

## 1. Prepare The Release Branch

- Confirm the branch is based on the latest `main`.
- Confirm `docs/CHANGELOG.md` and `docs/CHANGELOG_ko.md` have a section for the new version.
- Confirm README install/update instructions still match the release assets.
- For runtime-packet or generator-template changes, confirm the changelog names the generated `goal.md`, `tasks.md`, and `evidence.md` surface that users will receive through `hyper update`.
- Confirm no generated local test projects or temporary binaries are staged.

## 2. Run Local Validation

```bash
go test -count=1 ./...
go vet ./...
staticcheck ./...
govulncheck ./...
git diff --check
```

If the local Go cache is not writable in a sandboxed environment, use a writable cache path:

```bash
GOCACHE=/private/tmp/hyper-go-cache go test -count=1 ./...
```

For runtime-packet or generator-template changes, build the local binary and verify packet generation in a disposable project before tagging:

```bash
GOCACHE=/private/tmp/hyper-go-cache go build -o /private/tmp/hyper-local ./cmd/hyper
/private/tmp/hyper-local version
```

The disposable project check must confirm that generated packets include any new user-visible sections across:

- `.hyper/goals/<GOAL-ID>/goal.md`
- `.hyper/goals/<GOAL-ID>/tasks.md`
- `.hyper/goals/<GOAL-ID>/evidence.md`

For autonomous runtime-template releases, confirm these generated sections before publishing:

- `AI Control Charter`
- `External Reference Evolution`
- `Decision Hierarchy`
- `Autonomous Work Plan`
- `Autonomous Safety Policy`
- `Capability Expansion Policy`
- `Research Evidence Policy`
- `Loop Progress Policy`
- `Product Satisfaction Policy`

## 3. Merge The PR

- Push the branch.
- Open a PR.
- Wait for CI on Linux, macOS, and Windows.
- Merge only after CI passes.

## 4. Tag The Release

```bash
git switch main
git pull --ff-only origin main
git tag -a vX.Y.Z -m "vX.Y.Z"
git push origin vX.Y.Z
```

## 5. Verify GitHub Actions

Confirm both workflows pass:

- CI
- Release

The release workflow should publish:

- `hyper-darwin-arm64`
- `hyper-darwin-amd64`
- `hyper-linux-arm64`
- `hyper-linux-amd64`
- `hyper-windows-amd64.exe`
- `checksums.txt`
- one `.sigstore.json` bundle for each binary and for `checksums.txt`

## 6. Smoke Test Install And Update

On macOS or Linux:

```bash
hyper update
hyper version
hyper doctor
```

For an existing project, refresh project state after the binary update:

```bash
hyper migrate
hyper doctor
hyper status --short
```

On Windows PowerShell:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/KoreanCode/orange-hyper-run/main/install.ps1 | iex"
hyper version
hyper doctor
```

## 7. Verify Checksums And Optional Signatures

The default user path should verify checksums automatically.

If `cosign` is installed, install/update should also verify the sigstore bundle. To require signature verification:

```bash
HYPER_RUN_VERIFY_SIGNATURE=required hyper update
```

PowerShell:

```powershell
$env:HYPER_RUN_VERIFY_SIGNATURE="required"
hyper update
```

## 8. Release Notes Check

Before announcing the release, confirm the release page has:

- the expected tag
- all assets
- generated notes or manual notes that match the changelog
- no older asset overwritten by mistake

## 9. If Something Is Wrong

- If the tag was pushed but the release failed, inspect the Release workflow first.
- If assets are missing, rerun the Release workflow or re-push only after understanding the failure.
- If users have not consumed the release yet and the tag is clearly wrong, delete the GitHub release and tag, then publish a corrected tag.
- If users may have consumed it, publish a patch release instead.
