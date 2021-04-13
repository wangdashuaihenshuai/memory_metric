package metric

import "time"

const (
	TimerSumEnd   = "_sum"
	TimerCountEnd = "_count"
)

type Tags map[string]string

type Counter interface {
	WithTag(key string, value string) Counter
	Value(value int64)
	Count()
}

type Timer interface {
	WithTag(key string, value string) Timer
	Start()
	End()
	Value(value int64)
}

type Metric interface {
	GetStorage() Storage
	Count(name string, tags Tags, value int64)
	Timer(name string, tags Tags, value int64)
	NewCounter(name string) Counter
	NewTimer(name string) Timer
}

type Storage interface {
	Store(metricName string, point Point)
	Load(metricName string, time time.Time) []Point
	LoadAllTimeRange(metricName string) map[string][]Point
	LoadAll() map[string]map[string][]Point
}

type Point interface {
	GetAggregationKey() string
	GetTime() time.Time
	GetValue() int64
	Merge(p Point)
}
