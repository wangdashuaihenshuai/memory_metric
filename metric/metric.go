package metric

type metric struct {
	storage Storage
}

func NewMetric(opt *MetricOptions) Metric {
	return &metric{
		storage: NewStorage(opt),
	}
}

func (m *metric) GetStorage() Storage {
	return m.storage
}

func (m *metric) Count(name string, tags Tags, value int64) {
	point := newPoint(tags, value)
	m.storage.Store(name, point)
}

func (m *metric) Timer(name string, tags Tags, value int64) {
	point := newPoint(tags, value)
	m.storage.Store(name+TimerSumEnd, point)

	point = newPoint(tags, 1)
	m.storage.Store(name+TimerCountEnd, point)
}

func (m *metric) NewCounter(name string) Counter {
	return newCounter(name, m.storage)
}

func (m *metric) NewTimer(name string) Timer {
	return newTimer(name, m.storage)
}
