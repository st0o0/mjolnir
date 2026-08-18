# Changelog

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
