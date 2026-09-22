# Changelog

## [0.1.5](https://github.com/st0o0/mjolnir/compare/v0.1.4...v0.1.5) (2026-09-22)


### Features

* extend docker preset for base image digest pinning + automerge ([41c9842](https://github.com/st0o0/mjolnir/commit/41c9842059a865c7c5e9fa6388144bee366c4363))
* migrate to modular build and docker workflows ([a5d2241](https://github.com/st0o0/mjolnir/commit/a5d22416c178a184ba260dcda18d892632ebf128))
* migrate to multi-stage Dockerfile ([3956fbc](https://github.com/st0o0/mjolnir/commit/3956fbc05f5e6f5e452ffaa6d2ea2d3c0d1d73c0))


### Bug Fixes

* add id-token permission for cosign signing in dev builds ([11dafdf](https://github.com/st0o0/mjolnir/commit/11dafdfc1e28b15d303b33c2cd173234216d8cdc))
* grant contents:read and drop paths filter in dev-build docker job ([3654a3b](https://github.com/st0o0/mjolnir/commit/3654a3ba082fd953cc2f7dc73f920f7ddeb66b5a))


### Documentation

* add README badges ([7731177](https://github.com/st0o0/mjolnir/commit/773117769ce46f00ed97b49f6b81eb9dd8883167))

## [0.1.4](https://github.com/st0o0/mjolnir/compare/v0.1.3...v0.1.4) (2026-09-12)


### Features

* decouple release-please from build workflow ([0538ba8](https://github.com/st0o0/mjolnir/commit/0538ba8ccf87f1569edc2143b120a0e0263b539c))


### Bug Fixes

* add apk upgrade to resolve OpenSSL CVEs in runtime image ([639ba85](https://github.com/st0o0/mjolnir/commit/639ba85a60ad51f35d83491a8d5793f604a7f3e9))


### Refactoring

* migrate to shared reusable workflows ([3b53083](https://github.com/st0o0/mjolnir/commit/3b53083150aacd2ca7832fab3089a559c256294b))
* rename CI jobs for cleaner GitHub check names ([9196adb](https://github.com/st0o0/mjolnir/commit/9196adb21c5ee28ea955cc51d94152cbb84b92b7))


### Dependencies

* bump golang from 1.26-alpine to 1.27-alpine ([4b32502](https://github.com/st0o0/mjolnir/commit/4b3250293e4eb312d4e85056575e3ebf944be139))
* bump hadolint/hadolint-action in the actions-all group ([fd48648](https://github.com/st0o0/mjolnir/commit/fd486487374a691a172bc651b76d05faf0513977))

## [0.1.3](https://github.com/st0o0/mjolnir/compare/v0.1.2...v0.1.3) (2026-08-18)


### Bug Fixes

* chown /run/nut/ to nut:nut after config generation ([14afb03](https://github.com/st0o0/mjolnir/commit/14afb032b32425fc5d556598131f04793f846764))
* use 0644 permissions for generated NUT config files ([32db7d5](https://github.com/st0o0/mjolnir/commit/32db7d563a134e89fe80a60d95342423416b1d6c))

## [0.1.2](https://github.com/st0o0/mjolnir/compare/v0.1.1...v0.1.2) (2026-08-18)


### Features

* export non-numeric NUT variables as Prometheus metrics and add /diagnostics endpoint ([2370cfb](https://github.com/st0o0/mjolnir/commit/2370cfb697bd45a5c30965d6b914811e5d694773))

## [0.1.1](https://github.com/st0o0/mjolnir/compare/v0.1.0...v0.1.1) (2026-08-17)


### Features

* support multiple NUT users via NUT_USER_&lt;n&gt;_* env vars ([2d55d18](https://github.com/st0o0/mjolnir/commit/2d55d187886806ed3e1b1583170a7db686776ea8))

## [0.1.0](https://github.com/st0o0/mjolnir/compare/v0.1.0...v0.1.0) (2026-08-17)


### Features

* add Go entrypoint with config parsing, NUT daemon management, health server and Prometheus metrics ([a03ce98](https://github.com/st0o0/mjolnir/commit/a03ce98af431b0eef89661ceaec9ecdbc8698c6a))
* dynamic prometheus metrics from all NUT variables with status flags and device info ([0175bc4](https://github.com/st0o0/mjolnir/commit/0175bc4e6a56a8b4f478d13cf8f55a28fe1168f4))
* replace shell entrypoint with Go binary in multi-stage Dockerfile ([0bb873e](https://github.com/st0o0/mjolnir/commit/0bb873e695eb90090a36a8da1b6beb515e4b3290))


### Bug Fixes

* ignore hadolint DL3064 false positive on NUT_USER and NUT_SECRET_NAME ([edbd277](https://github.com/st0o0/mjolnir/commit/edbd277dcf8d8e1ad0eaf1898a32125f12955f26))
* remove NUT_PASSWORD from ENV to satisfy hadolint DL3064 ([c9d1351](https://github.com/st0o0/mjolnir/commit/c9d135131e4ca493a368616d2d696e0f7aa77c84))
* resolve golangci-lint errcheck and staticcheck warnings ([fa2d078](https://github.com/st0o0/mjolnir/commit/fa2d07825862d5b039b1517a1cbad340bbc0bd43))
* update OpenSpec config to reflect Go tech stack ([c05f76b](https://github.com/st0o0/mjolnir/commit/c05f76bba52058ad64a4c0a2e916f9e513c5887f))


### Documentation

* rewrite README for a more natural tone ([56f3f72](https://github.com/st0o0/mjolnir/commit/56f3f7230e58af9f95c78f3acc55a8949aa99daf))
