# Go App

An opinionated Go Application package with everything you need:

- multi-service/worker controller for launching multiple processes and keeping central control over them
- creating and handling application errors using error codes and error messages for exposing errors.
- structured logging
- configuration finder and parser
- multi-protocol API (HTTP, gRPC and NSQ)
- event emitter (NSQ)
- database connector (PostgreSQL)

## Requirements

* At least Go 1.9

## Installing the lib

### Global install

```
go get -u gitlab.com/mandalore/go-app/app
```

### Dep (vendoring)

```
dep ensure -add github.com/pkg/errors
```

## Samples

This project has a `samples` branch with sample applications. We recommend you take a look at that branch.

## TODO

- configuration 
  - add support for etcd


