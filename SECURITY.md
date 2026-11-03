# Security Policy

## Supported versions

The latest released image (`ghcr.io/st0o0/mjolnir:latest`) receives security
updates. Older tags are not maintained, so pin to a `MAJOR.MINOR` tag and update
regularly.

## Reporting a vulnerability

Please report security issues **privately** via GitHub's
[private vulnerability reporting](https://github.com/st0o0/mjolnir/security/advisories/new)
(Security > Advisories > *Report a vulnerability*). Do **not** open a public
issue for security problems.

You can expect an initial response within a few days. Once a fix is available it
is released and the advisory is published.

## Scope

mjolnir is a Go entrypoint that manages NUT daemons and exports Prometheus
metrics. It handles UPS credentials via environment variables and Docker secrets.
Vulnerabilities in NUT itself should be reported to the NUT project; Go module
dependency issues should be reported upstream. The image is scanned weekly with
Trivy and results appear in the repository's Security tab.
