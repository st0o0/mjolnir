## Purpose

Define the CI/CD pipeline: multi-arch Docker image builds via GitHub Actions with buildx, publishing to GHCR, semver tagging via release-please, layer caching, PR validation, image signing, security scanning, conventional commit enforcement, and automated dependency updates.

## Requirements

### Requirement: Multi-arch image build

Built for `linux/amd64` and `linux/arm64` via Docker buildx + QEMU.

#### Scenario: Multi-arch build on release
- **WHEN** a release is created by release-please
- **THEN** Docker images are built for `linux/amd64` and `linux/arm64` and pushed as a multi-arch manifest to GHCR

---

### Requirement: GHCR publishing

On release: tags `<version>`, `<major>.<minor>`, `latest`. On PR: build only, no push. Version is determined by release-please, not git tags.

#### Scenario: Release publish
- **WHEN** release-please creates a release with version `1.2.3`
- **THEN** images are pushed to GHCR with tags `1.2.3`, `1.2`, and `latest`

#### Scenario: PR build
- **WHEN** a pull request is opened
- **THEN** the image is built but not pushed to GHCR

---

### Requirement: Build caching

Docker layer cache via GitHub Actions cache (`type=gha`) on all build workflows.

#### Scenario: Cached build
- **WHEN** a build runs with unchanged Go dependencies
- **THEN** the Go module download and build layers are served from GHA cache

---

### Requirement: PR validation pipeline

Pull requests SHALL run four jobs in sequence: unit tests (`go test -race`), lint (golangci-lint + hadolint), Docker build + smoke test, and e2e tests. All four jobs SHALL be required checks for merge.

#### Scenario: Unit test failure blocks merge
- **WHEN** a PR has a failing unit test
- **THEN** the `unit` job fails and the PR cannot be merged

#### Scenario: Lint failure blocks merge
- **WHEN** a PR introduces a golangci-lint violation
- **THEN** the `lint` job fails and the PR cannot be merged

#### Scenario: E2e runs after cheap checks pass
- **WHEN** unit, lint, and build jobs all pass
- **THEN** the e2e job runs (it depends on the first three)

---

### Requirement: Release automation via release-please

Pushes to main SHALL trigger release-please to manage semver versioning and changelogs. When a release is created, the Docker publish pipeline SHALL run automatically.

#### Scenario: Conventional commit triggers release PR
- **WHEN** a `feat:` commit is pushed to main
- **THEN** release-please creates or updates a release PR with a minor version bump

#### Scenario: Release PR merge triggers publish
- **WHEN** the release PR is merged
- **THEN** release-please creates a GitHub release and the Docker publish job runs

---

### Requirement: Image signing and attestation

Released images SHALL be signed with cosign (keyless via OIDC) and include SLSA build provenance and SBOM attestations.

#### Scenario: Signed release image
- **WHEN** a release image is pushed to GHCR
- **THEN** cosign signs the image digest using keyless OIDC signing

#### Scenario: Provenance attestation
- **WHEN** a release image is built
- **THEN** SLSA provenance and SBOM attestations are attached to the image

---

### Requirement: Dev build workflow

Pull requests with the `dev-build` label SHALL trigger an environment-gated dev image push to GHCR. The workflow SHALL comment the pull command on the PR.

#### Scenario: Dev build triggered
- **WHEN** a PR is labeled `dev-build` and the `dev` environment is approved
- **THEN** an amd64-only image is pushed to GHCR with tags `pr-<number>` and `dev-<sha>`

#### Scenario: Dev build comment
- **WHEN** a dev build completes
- **THEN** a comment is posted (or updated) on the PR with the `docker pull` command

---

### Requirement: Security scanning with Trivy

Trivy SHALL scan the built Docker image on PRs (when Dockerfile/go.mod/go.sum change) and on a weekly schedule. Results SHALL be uploaded as SARIF to the GitHub Security tab.

#### Scenario: Trivy scan on PR
- **WHEN** a PR modifies `Dockerfile`, `go.mod`, or `go.sum`
- **THEN** Trivy scans the built image for CRITICAL and HIGH vulnerabilities

#### Scenario: Weekly Trivy scan
- **WHEN** the weekly cron fires
- **THEN** Trivy scans the current main image and uploads SARIF results

---

### Requirement: Conventional commit enforcement

All PR commits SHALL follow the Conventional Commits specification. Enforced via commitlint with the `@commitlint/config-conventional` preset.

#### Scenario: Non-conventional commit rejected
- **WHEN** a PR contains a commit message like `fixed stuff`
- **THEN** the commitlint check fails

#### Scenario: Conventional commit accepted
- **WHEN** all PR commits follow `type(scope): description` format
- **THEN** the commitlint check passes

---

### Requirement: Automated dependency updates

Dependabot SHALL be configured for weekly updates across three ecosystems: gomod, github-actions, and docker. Updates SHALL be grouped by ecosystem.

#### Scenario: Go dependency update
- **WHEN** a new version of a Go dependency is available
- **THEN** Dependabot creates a grouped PR with the update, using `deps` commit prefix
