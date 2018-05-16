# Go-Logger v1.0

A logging library for Go.

My view on application logging

* Should not be where your CPU spends most of the time.
* Should be done to STDERR/STDOUT. Log rotation, logging to log dispatchers or centralized log aggregators, etc should be the job of operations not development.
* Should not be confused with auditing, although it can be rudementarily used for it.
* Should not cause application failures, even if nothing is being logged due to an error in the logger.
* Should be simple to use in development and powerful enough to be used in production without convoluted and complex configurations. Id est, it should permit structured logging while keeping a simple interface.

## Logging scopes and package usage.

In an opinionated view of logging, one can consider a logger to have 4 hierarchical levels

* Application level is from where all loggers should be created.
* Component level is associated with only a part of the application. This is typically a domain package, library, "class", etc.
* Instance level should be associated with a data structure and respective methods, in the OOP world that would be a class instance. These loggers should be derived from the component level logger.
* Function level is specific to a certain function, with parts of the input or state included in the fields and/or tags.

## Fields or Data?

Fields should represent data relevant to filtering or selecting log entries. Data, on the other hand, is only relevant to the entry itself and has little value in terms of finding the log entry.

Field examples: a user's ID, action, application component.
Data examples: a user's email or name

## Possible Future Features

> TODO

## Quick Usage

### Download the package

If you are using [Dep](https://github.com/golang/dep) (recommended) you can simply add the package to your code and run `dep ensure` or install it in your project using

```Bash
dep ensure -add gitlab.com/vredens/go-logger
```

Or install it globally using plain old `go get`

```Bash
go get -u gitlab.com/vredens/go-logger
```

### Using it in your code (Simple, the recommended interface)

```Go
import logger "gitlab.com/vredens/go-logger"

// create a package level logger
var log = logger.SpawnSimple(
	logger.WithFields(map[string]interface{}{"component": "mypackage"}),
	logger.WithTags("user", "account"),
)

func main() {
	// log with our package level logger
	log.Write("this is a message and we use the tag info instead of a log level", logger.Tags{"bootstrap", "info"})

	// logging with extra data
	log.Dump("A message with structured data", nil, logger.KV{"id": 10, "age": 50, "name": "john smith"})

	// create a new logger with some fields
	userLog := log.SpawnSimple(logger.WithFields(logger.KV{"user_type": "me", "user_id": 1234}))

	// logging with the new logger
	userLog.Dump("here's a log with some extra data", logger.Tags{"moardata"}, logger.KV{"kiss": "is simple"})

	// change the logging format and min log level
	logger.Reconfigure(logger.WithLevel(logger.DEBUG), logger.WithFormat(logger.FormatJSON))

	// Log a debug message
	userLog.Dbg("this will show up now", nil, nil)
}
```

### Using it in your code (Logger, the typical interface)

```Go
import logger "gitlab.com/vredens/go-logger"

// create a package level logger
var log = logger.Spawn(
	logger.WithFields(map[string]interface{}{"component": "mypackage"}),
	logger.WithTags("user", "account"),
)

func main() {
	// logging with the default global logger
	logger.Debugf("This is a debug of %s", "hello world")

	// log with our package level logger
	log.Info("this is a message")

	// logging with extra data
	log.WarnData("A message with structured data", logger.KV{"id": 10, "age": 50, "name": "john smith"})

	// create a new small scope logger with some fields
	userLog := log.WithFields(logger.KV{"user_type": "me", "user_id": 1234})

	// logging with the new logger
	userLog.WithData(logger.KV{"kiss": "is simple"}).Debug("here's a log with some extra data")

	// change the logging format and min log level
	logger.Reconfigure(logger.WithLevel(logger.DEBUG), logger.WithFormat(logger.FormatJSON))

	// Log an error message with extra tags and fields
	userLog.With(logger.Tags{"critical"}, logger.KV{"ctx_id":666}).Error("the devil is around here!")
}
```
