package app

import "time"

var appStats = NewStatsCollector()

// StatsCounterReset resets the counter value back to 0.
func StatsCounterReset(namespace string) {
	appStats.CounterReset(namespace)
}

// StatsCounterIncr increment a counter stat.
func StatsCounterIncr(namespace string, value int64) {
	appStats.CounterIncr(namespace, value)
}

// StatsCounterGet returns the current value of the provided stat namespace.
func StatsCounterGet(namespace string) int64 {
	return appStats.CounterGet(namespace)
}

// StatsAverageAdd recalcs the average using the incremental average algorythm.
func StatsAverageAdd(namespace string, value float64) {
	appStats.AverageAdd(namespace, value)
}

// StatsTimeAverageAdd will take the provided time and calculate the time diference to the present time and add that to the named average. Time unit is miliseconds.
func StatsTimeAverageAdd(namespace string, start time.Time) {
	appStats.TimeAverageAdd(namespace, start)
}

// StatsDump converts stats into a JSON file.
func StatsDump() ([]byte, error) {
	return appStats.Dump()
}

// StatsDumpAndReset converts statistics to a JSON byte array and resets them.
func StatsDumpAndReset() ([]byte, error) {
	return appStats.DumpAndReset()
}

// StatsGet returns a copy of the current application statistics.
func StatsGet() Stats {
	return appStats.Get()
}

// StatsGetAndReset returns a copy of the current application statistics and resets them.
func StatsGetAndReset() Stats {
	return appStats.GetAndReset()
}
