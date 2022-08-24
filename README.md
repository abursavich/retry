# Retry
[![License][license-img]][license]
[![GoDev Reference][godev-img]][godev]
[![Go Report Card][goreportcard-img]][goreportcard]

Package retry provides backoff algorithms for retryable processes.

It was inspired by [github.com/cenkalti/backoff/v4] which is a port of [Google's HTTP Client Library for Java].



[license]: https://raw.githubusercontent.com/abursavich/retry/main/LICENSE
[license-img]: https://img.shields.io/badge/license-mit-blue.svg?style=for-the-badge

[godev]: https://pkg.go.dev/bursavich.dev/retry
[godev-img]: https://img.shields.io/static/v1?logo=go&logoColor=white&color=00ADD8&label=dev&message=reference&style=for-the-badge

[goreportcard]: https://goreportcard.com/report/bursavich.dev/retry
[goreportcard-img]: https://goreportcard.com/badge/bursavich.dev/retry?style=for-the-badge

[github.com/cenkalti/backoff/v4]: https://github.com/cenkalti/backoff
[Google's HTTP Client Library for Java]: https://github.com/google/google-http-java-client/blob/da1aa993e90285ec18579f1553339b00e19b3ab5/google-http-client/src/main/java/com/google/api/client/util/ExponentialBackOff.java
