# reili Example - Limit by request IP
A simple http Rate Limiter by requester's IP

## Run the example
```
$ go run main.go
```

## Usage
The only endpoint tha is supported is the `/` which is limited by 1 request per second and up to 3 burst requests per IP.
