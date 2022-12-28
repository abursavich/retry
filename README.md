# Retry
[![License][license-img]][license]
[![GoDev Reference][godev-img]][godev]
[![Go Report Card][goreportcard-img]][goreportcard]

Package retry provides backoff algorithms for retryable processes.

It was inspired by [github.com/cenkalti/backoff/v4][backoff] which is a port of [Google's HTTP Client
Library for Java].


## Why?

It separates state from policy, which reduces allocations and allows a single policy instance to be used
concurrently by all callers, and it uses explicit return values instead of magic sentinel values.

```go
type Policy interface {
    Next(err error, start, now time.Time, attempt int) (backoff time.Duration, retry bool)
}
```

It decomposes features and encourages their composition.

```go
policy := retry.WithRandomJitter(retry.ConstantBackoff(time.Second), 0.5)
```

It makes [context] first-class and improves call ergonomics.

```go
err := retry.Do(ctx, policy, func() error {
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
