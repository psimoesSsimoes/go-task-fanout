package app_test

import (
	"sync"
	"testing"
	"time"

	. "gitlab.com/mandalore/go-app/app"
)

func checkAverage(t *testing.T, stat Average, min, avg, max float64, total int64) {
	if stat.Total != total {
		t.Errorf("expected %d to be %d", stat.Total, total)
	}
	if stat.Min != min {
		t.Errorf("expected %f to be %f", stat.Min, min)
	}
	if stat.Average != avg {
		t.Errorf("expected %f to be %f", stat.Average, avg)
	}
	if stat.Max != max {
		t.Errorf("expected %f to be %f", stat.Max, max)
	}
}

func TestStatsAverage(t *testing.T) {
	StatsAverageAdd("test-avg-01", 10)
	StatsAverageAdd("test-avg-01", 20)
	StatsAverageAdd("test-avg-01", 30)

	stats := StatsGet()
	checkAverage(t, stats.Averages["test-avg-01"], 10.0, 20.0, 30.0, 3)

	stats = StatsGetAndReset()
	checkAverage(t, stats.Averages["test-avg-01"], 10.0, 20.0, 30.0, 3)

	stats = StatsGet()
	checkAverage(t, stats.Averages["test-avg-01"], 0.0, 0.0, 0.0, 0)
}

func TestStatsTimeAverage(t *testing.T) {
	ts := time.Now()
	<-time.After(time.Millisecond)
	StatsTimeAverageAdd("test-tavg-01", ts)
	<-time.After(time.Millisecond)
	StatsTimeAverageAdd("test-tavg-01", ts)
	<-time.After(time.Millisecond)
	StatsTimeAverageAdd("test-tavg-01", ts)

	stats := StatsGet()
	stat := stats.Averages["test-tavg-01"]
	if stat.Min < 1 || stat.Min >= 2 {
		t.Errorf("expected min time to be around the 1-2 microseconds instead of %.2f", stat.Min)
	}
	if stat.Average < 2 || stat.Average >= 4 {
		t.Errorf("expected average time to be around the 2-4 microseconds instead of %.2f", stat.Average)
	}
	if stat.Max < 3 || stat.Max >= 5 {
		t.Errorf("expected max time to be around 3-5 microseconds instead of %.2f", stat.Max)
	}
	if stat.Total != 3 {
		t.Errorf("expected total to be 3 instead of %d", stat.Total)
	}
}

func TestStatsCounter(t *testing.T) {
	StatsCounterIncr("test-cnt-01", 1)
	StatsCounterIncr("test-cnt-01", 5)
	StatsCounterIncr("test-cnt-01", -2)

	stats := StatsGet()
	if stats.Counters["test-cnt-01"] != 4 {
		t.Errorf("expected %d to be 4", stats.Counters["test-cnt-01"])
	}
}

func TestStatsConcurrency(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			ts := time.Now()
			StatsCounterIncr("test-cnt", 1)
			StatsAverageAdd("test-avg", 10.0)
			StatsTimeAverageAdd("test-tavg", ts)
			wg.Done()
		}()
	}
	wg.Wait()

	stats := StatsGet()
	checkAverage(t, stats.Averages["test-avg"], 10.0, 10.0, 10.0, 4)
	if stats.Counters["test-cnt"] != 4 {
		t.Errorf("expected concurrent counter value %d to be 4", stats.Counters["test-cnt"])
	}
}
