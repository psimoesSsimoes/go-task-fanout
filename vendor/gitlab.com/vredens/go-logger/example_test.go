package logger

func Example() {
	// logging with the default global logger
	Debugf("This is a debug of %s", "hello world")

	// create a new logger
	var log = SpawnSimple(WithFields(map[string]interface{}{"component": "mypackage"}), WithTags("user", "account"))

	// log something
	log.Write("this is a message and we use the tag info instead of a log level", Tags{"bootstrap", "info"})

	// log with extra data
	log.Dump("A message with structured data", nil, KV{"id": 10, "age": 50, "name": "john smith"})

	// create a new logger from another logger
	userLog := log.SpawnSimple(WithFields(KV{"user_type": "me", "user_id": 1234}))

	// logging with the new logger
	userLog.Dump("here's a log with some extra data", Tags{"moardata"}, KV{"kiss": "is simple"})

	// change the logging format and min log level
	Reconfigure(
		WithLevel(DEBUG),
		WithFormat(FormatJSON),
	)

	// do a debug log
	userLog.Dbg("this will show up now", nil, nil)
}
