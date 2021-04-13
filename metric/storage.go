package metric

import (
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func newPoint(tags Tags, value int64) Point {
	return &point{
		Tags:      tags,
		Value:     value,
		CreatedAt: time.Now(),
	}
}

type point struct {
	Tags      Tags
	Value     int64
	CreatedAt time.Time
}

func (p *point) GetAggregationKey() string {
	keys := make([]string, len(p.Tags))
	for k, _ := range p.Tags {
		keys = append(keys, k)
	}

	var ret string
	sort.Strings(keys)
	for _, key := range keys {
		v := p.Tags[key]
		ret = ret + key + v
	}

	return ret
}

func (p *point) GetTime() time.Time {
	return p.CreatedAt
}

func (p *point) GetValue() int64 {
	return p.Value
}

func (p *point) Merge(mergePoint Point) {
	atomic.AddInt64(&p.Value, mergePoint.GetValue())
}

type MetricOptions struct {
	MaxTimeNumber   int64
	MaxPointNumber  int64
	MaxMetricNumber int64
}

type metricStorage struct {
	key             string
	maxPointNumber  int64
	maxMetricNumber int64
	lock            sync.RWMutex
	aggregationMap  map[string]*aggregationStorage
}

func (s *metricStorage) Store(metricName string, point Point) {
	ts := s.GetOrCreateKeyValue(metricName)
	ts.Store(point)
}

func (s *metricStorage) Load(metricName string) []Point {
	ts, ok := s.GetValue(metricName)
	if !ok {
		return []Point{}
	}

	return ts.Load()
}

func (s *metricStorage) LoadAll() map[string][]Point {
	s.lock.RLock()
	defer s.lock.RUnlock()
	ret := map[string][]Point{}
	for k, m := range s.aggregationMap {
		ret[k] = m.Load()
	}

	return ret
}

func (s *metricStorage) GetValue(key string) (*aggregationStorage, bool) {
	s.lock.RLock()
	v, ok := s.aggregationMap[key]
	s.lock.RUnlock()
	if ok {
		return v, true
	}

	return nil, false
}

func (s *metricStorage) GetOrCreateKeyValue(key string) *aggregationStorage {
	s.lock.RLock()
	v, ok := s.aggregationMap[key]
	s.lock.RUnlock()
	if ok {
		return v
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	v, ok = s.aggregationMap[key]
	if ok {
		return v
	}

	if len(s.aggregationMap) > int(s.maxMetricNumber) {
		//TODO
	}

	v = newAggregationStoreage(s.maxPointNumber)
	s.aggregationMap[key] = v
	return v
}

func NewPoint() Point {
	return &point{}
}

func newMetricStorage(key string, maxMetricNumber int64, maxPointNumber int64) *metricStorage {
	return &metricStorage{
		key:             key,
		maxPointNumber:  maxPointNumber,
		maxMetricNumber: maxMetricNumber,
		aggregationMap:  make(map[string]*aggregationStorage, maxPointNumber),
	}
}

func NewStorage(opt *MetricOptions) Storage {
	return &storage{
		maxTimeNumber:   opt.MaxTimeNumber,
		maxPointNumber:  opt.MaxPointNumber,
		maxMetricNumber: opt.MaxMetricNumber,
		metricArray:     make([]*metricStorage, 0, opt.MaxTimeNumber),
		metricMap:       make(map[string]*metricStorage, opt.MaxTimeNumber),
	}
}

type storage struct {
	maxTimeNumber   int64
	maxPointNumber  int64
	maxMetricNumber int64
	metricArray     []*metricStorage
	metricMap       map[string]*metricStorage
	lock            sync.RWMutex
}

func (ts *storage) timeSerise(time time.Time) string {
	return strconv.FormatInt(time.Unix(), 10)
}

func (ts *storage) Store(metricName string, point Point) {
	key := ts.timeSerise(point.GetTime())
	as := ts.GetOrCreateKeyValue(key)
	as.Store(metricName, point)
}

func (ts *storage) LoadAll() map[string]map[string][]Point {
	ts.lock.RLock()
	defer ts.lock.RUnlock()
	ret := map[string]map[string][]Point{}
	for k, v := range ts.metricMap {
		ret[k] = v.LoadAll()
	}

	return ret
}

func (ts *storage) LoadAllTimeRange(metricName string) map[string][]Point {
	ts.lock.RLock()
	defer ts.lock.RUnlock()
	ret := map[string][]Point{}
	for k, v := range ts.metricMap {
		ret[k] = v.Load(metricName)
	}

	return ret
}

func (ts *storage) Load(metricName string, time time.Time) []Point {
	key := ts.timeSerise(time)
	as, ok := ts.GetValue(key)
	if !ok {
		return []Point{}
	}

	return as.Load(metricName)
}

func (ts *storage) GetValue(key string) (*metricStorage, bool) {
	ts.lock.RLock()
	defer ts.lock.RUnlock()
	v, ok := ts.metricMap[key]
	if ok {
		return v, true
	}

	return nil, false
}

func (ts *storage) GetOrCreateKeyValue(key string) *metricStorage {
	ts.lock.RLock()
	v, ok := ts.metricMap[key]
	ts.lock.RUnlock()
	if ok {
		return v
	}

	ts.lock.Lock()
	defer ts.lock.Unlock()
	v, ok = ts.metricMap[key]
	if ok {
		return v
	}

	v = newMetricStorage(key, ts.maxPointNumber, ts.maxPointNumber)
	ts.metricMap[key] = v
	ts.metricArray = append(ts.metricArray, v)
	if len(ts.metricArray) > int(ts.maxTimeNumber) {
		first := ts.metricArray[0]
		ts.metricArray = ts.metricArray[1:]
		delete(ts.metricMap, first.key)
	}
	return v
}

func newAggregationStoreage(maxPointNumber int64) *aggregationStorage {
	return &aggregationStorage{
		maxPointNumber: maxPointNumber,
		pointArray:     make([]Point, 0, maxPointNumber),
		pointMap:       make(map[string]Point, maxPointNumber),
	}
}

type aggregationStorage struct {
	key            string
	maxPointNumber int64
	pointMap       map[string]Point
	pointArray     []Point
	pointNumber    int64
	lock           sync.RWMutex
}

func (as *aggregationStorage) GetOrCreateKeyValue(key string, point Point) (Point, bool) {
	as.lock.RLock()
	v, ok := as.pointMap[key]
	as.lock.RUnlock()
	if ok {
		return v, true
	}

	as.lock.Lock()
	defer as.lock.Unlock()
	v, ok = as.pointMap[key]
	if ok {
		return v, true
	}

	if as.pointNumber > as.maxPointNumber {
		//TODO
	}

	as.pointMap[key] = point
	as.pointArray = append(as.pointArray, point)
	as.pointNumber = as.pointNumber + 1
	return nil, false
}

func (as *aggregationStorage) Store(point Point) {
	key := point.GetAggregationKey()
	findPoint, ok := as.GetOrCreateKeyValue(key, point)
	if ok {
		findPoint.Merge(point)
	}
}

func (as *aggregationStorage) Load() []Point {
	as.lock.RLock()
	defer as.lock.RUnlock()
	return as.pointArray
}
