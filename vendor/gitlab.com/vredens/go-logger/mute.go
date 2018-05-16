package logger

import (
	"io/ioutil"
)

// SpawnMute creates a new logger which implements the Logger interface but blocks all logging. This is simply a helper in order to inject loggers into components to keep them quiet.
func SpawnMute() Logger {
	w := NewWriter(WithOutput(ioutil.Discard), WithLevel(OFF))
	return w.Spawn()
}

// SpawnSimpleMute creates a new logger which implements the Simple interface but blocks all logging. This is simply a helper in order to inject loggers into components to keep them quiet.
func SpawnSimpleMute() Simple {
	w := NewWriter(WithOutput(ioutil.Discard), WithLevel(OFF))
	return w.SpawnSimple()
}

// SpawnCompatibleMute creates a new logger which implements the Compatible interface but blocks all logging. This is simply a helper in order to inject loggers into components to keep them quiet.
func SpawnCompatibleMute() Compatible {
	w := NewWriter(WithOutput(ioutil.Discard), WithLevel(OFF))
	return w.SpawnCompatible()
}
