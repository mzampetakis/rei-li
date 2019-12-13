# reili
A simple but general purpose http Rate Limiter

## Installing
```
$ go get -u github.com/mzampetakis/reili
```

## Usage

Define your own identifier function
```go
IdentifyVisitor(r *http.Request) (string, error)
```
following the interface that package `reili` provides
```go
type VisitorIdentifier interface {
	IdentifyVisitor(r *http.Request) (string, error)
}
```

Create a rate limiter using `reili.NewRateLimiter(reqPerSec, burstReq, visitorIdentifier)` and then limit your server's mux using your limiter.


## Examples
[Limit by request IP](examples/limitByIp)
