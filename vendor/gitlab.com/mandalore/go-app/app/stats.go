package app

import (
	"encoding/json"
	"sync"
	"time"
)

// StatsCollector interface for a collector of statistical data such as counters and averages.
type StatsCollector interface {
	// CounterReset will reset the named counter back to 0.
	CounterReset(name string)
	// CounterIncr will increment the named counter by the provided amount (can be a negative value).
	CounterIncr(name string, amount int64)
	// CounterGet will return the current value of the named counter.
	CounterGet(name string) int64

	// TimeAverageAdd will take the provided time and calculate the time diference to the present time and add that to the named average. Time unit is miliseconds.
	TimeAverageAdd(name string, value time.Time)
	// AverageAdd will add the provided value to the named average collector.
	AverageAdd(name string, value float64)
	// AverageGet will return the named average.
	AverageGet(name string) Average
	// AverageReset will reset the named average back to the zero value.
	AverageReset(name string)

	// Dump converts stats into a JSON file.
	Dump() ([]byte, error)
	// DumpAndReset converts statistics to a JSON byte array and resets them.
	DumpAndReset() ([]byte, error)

	// Get returns all statistics.
	Get() Stats
	// GetAndReset returns all statistics and then resets all values.
	GetAndReset() Stats
}

// NewStatsCollector creates a new instance of a StatsCollector.
func NewStatsCollector() StatsCollector {
	return &stats{
		mux:      new(sync.Mutex),
		Counters: make(map[string]int64),
		Averages: make(map[string]*Average),
	}
}

// Average is a data structure containing basic statistical information around the notion of average value.
type Average struct {
	Min     float64 `json:"min,omitempty"`
	Max     float64 `json:"max,omitempty"`
	Average float64 `json:"average,omitempty"`
	Total   int64   `json:"total,omitempty"`
}

// Stats is a data structure aggregating multiple statistical data.
type Stats struct {
	Counters map[string]int64   `json:"counters,omitempty"`
	Averages map[string]Average `json:"averages,omitempty"`
}

type stats struct {
	mux      *sync.Mutex
	Counters map[string]int64    `json:"counters,omitempty"`
	Averages map[string]*Average `json:"averages,omitempty"`
}

// CounterReset resets the counter value back to 0.
func (s *stats) CounterReset(namespace string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.Counters[namespace] = 0
}

// CounterIncr increment a counter stat.
func (s *stats) CounterIncr(namespace string, value int64) {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.Counters[namespace] += value
}

// CounterGet returns the current value of the provided stat namespace.
func (s *stats) CounterGet(namespace string) int64 {
	s.mux.Lock()
	defer s.mux.Unlock()

	if stat, found := s.Counters[namespace]; found {
		return stat
	}

	return 0
}

// AverageAdd recalcs the average using the incremental average algorythm.
func (s *stats) AverageAdd(namespace string, value float64) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if avg, found := s.Averages[namespace]; found {
		if value > avg.Max {
			avg.Max = value
		}
		if value < avg.Min {
			avg.Min = value
		}
		avg.Total++
		avg.Average = avg.Average*float64(avg.Total-int64(1.0))/float64(avg.Total) + value/float64(avg.Total)
	} else {
		s.Averages[namespace] = &Average{
			Average: value,
			Total:   1,
			Max:     value,
			Min:     value,
		}
	}
}

// AverageGet returns the Average associated with the provided namespace.
func (s *stats) AverageGet(namespace string) Average {
	s.mux.Lock()
	defer s.mux.Unlock()

	return *s.Averages[namespace]
}

// AverageReset resets the named average back to its zero value.
func (s *stats) AverageReset(namespace string) {
	s.mux.Lock()
	defer s.mux.Unlock()

	delete(s.Averages, namespace)
}

// TimeAverageAdd will take the provided time and calculate the time diference to the present time and add that to the named average. Time unit is miliseconds.
func (s *stats) TimeAverageAdd(namespace string, start time.Time) {
	now := time.Now()
	elapsed := now.Sub(start)

	s.AverageAdd(namespace, float64(elapsed.Nanoseconds())/1e6)
}

// Dump converts stats into a JSON file.
func (s *stats) Dump() ([]byte, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	return json.Marshal(s)
}

// DumpAndReset converts statistics to a JSON byte array and resets them.
func (s *stats) DumpAndReset() ([]byte, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	data, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	s.Counters = make(map[string]int64)
	s.Averages = make(map[string]*Average)

	return data, nil
}

// Get returns all statistics.
func (s *stats) Get() Stats {
	s.mux.Lock()
	defer s.mux.Unlock()

	out := Stats{
		Counters: make(map[string]int64),
		Averages: make(map[string]Average),
	}
	for k, v := range s.Counters {
		out.Counters[k] = v
	}
	for k, v := range s.Averages {
		out.Averages[k] = *v
	}

	return out
}

// GetAndReset returns all statistics and then resets all values.
func (s *stats) GetAndReset() Stats {
	s.mux.Lock()
	defer s.mux.Unlock()

	out := Stats{
		Counters: make(map[string]int64),
		Averages: make(map[string]Average),
	}
	for k, v := range s.Counters {
		out.Counters[k] = v
	}
	for k, v := range s.Averages {
		out.Averages[k] = *v
	}

	s.Counters = make(map[string]int64)
	s.Averages = make(map[string]*Average)

	return out
}
