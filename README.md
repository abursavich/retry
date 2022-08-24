# Retry
[![License][license-img]][license]
[![GoDev Reference][godev-img]][godev]
[![Go Report Card][goreportcard-img]][goreportcard]

Package retry provides backoff algorithms for retryable processes.

It was inspired by [github.com/cenkalti/backoff/v4][backoff] which is a port of [Google's HTTP Client Library for Java].

## Why?

I like [backoff] and I've used it often, but certain things about it have always irked me. I started off trying to write
a little wrapper around it to hide those things, but then I realized it required changes that couldn't retain the same
underlying implementation. So I wrote this from scratch.

This is how it's different:

- It removes retry state from objects, reducing allocations and allowing a single instance
  to be used concurrently by all callers.

    ```go
    type Policy interface {
        Next(start, now time.Time, attempt int) (backoff time.Duration, allow bool)
    }
    ```

- It decomposes features and encourages their composition.

    ```go
    policy := retry.WithRandomJitter(retry.ConstantBackoff(time.Second), rand, 0.5)
    ```

- It changes the primary call ergonomics and moves [context] to the forefront.

    ```go
    err := retry.Do(ctx, policy, func() errror {
        // ...
    })
    ```

[license]: https://raw.githubusercontent.com/abursavich/retry/main/LICENSE
[license-img]: https://img.shields.io/badge/license-mit-blue.svg?style=for-the-badge

[godev]: https://pkg.go.dev/bursavich.dev/retry
[godev-img]: https://img.shields.io/static/v1?logo=go&logoColor=white&color=00ADD8&label=dev&message=reference&style=for-the-badge

[goreportcard]: https://goreportcard.com/report/bursavich.dev/retry
[goreportcard-img]: https://goreportcard.com/badge/bursavich.dev/retry?style=for-the-badge

[backoff]: https://pkg.go.dev/github.com/cenkalti/backoff/v4
[Google's HTTP Client Library for Java]: https://github.com/google/google-http-java-client/blob/da1aa993e90285ec18579f1553339b00e19b3ab5/google-http-client/src/main/java/com/google/api/client/util/ExponentialBackOff.java
[context]: https://pkg.go.dev/context#Context
