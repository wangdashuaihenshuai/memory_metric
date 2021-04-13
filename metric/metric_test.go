package metric_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/wangdashuaihenshuai/memory_metric/metric"
)

var opt = &metric.MetricOptions{
	MaxTimeNumber:   2,
	MaxPointNumber:  10000,
	MaxMetricNumber: 5,
}

var m = metric.NewMetric(opt)

func TestMetric(t *testing.T) {
	wg := sync.WaitGroup{}
	loopTimes := 10000
	wg.Add(loopTimes)
	for i := 0; i < loopTimes; i++ {
		go func() {
			id := strconv.Itoa(i % 10)
			m.Count("test", metric.Tags(map[string]string{"tag1": id}), 1)
			m.Timer("test", metric.Tags(map[string]string{"tag1": id, "tag2": "2"}), 100)
			wg.Done()
		}()
	}
	wg.Wait()
	m.GetStorage().LoadAll()
}

func TestMetricTimeLimit(t *testing.T) {
	m.Count("test", metric.Tags(map[string]string{"tag1": "hello1"}), 1)
	time.Sleep(2 * time.Second)
	m.Count("test", metric.Tags(map[string]string{"tag1": "hello2"}), 1)
	time.Sleep(2 * time.Second)
	m.Count("test", metric.Tags(map[string]string{"tag1": "hello3"}), 1)
	ps := m.GetStorage().LoadAll()
	if len(ps) > 2 {
		t.Fatal("time limit error")
	}
}

func BenchmarkCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i % 10)
		m.Count("test"+id, metric.Tags(map[string]string{"tag1": "1"}), 1)
	}
}

func BenchmarkTimer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		id := strconv.Itoa(i % 10)
		m.Timer("test"+id, metric.Tags(map[string]string{"tag1": "1"}), 10)
	}
}
