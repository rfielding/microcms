package handler

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const MetricsLoad = 0
const MetricsStop = 1

var MetricsGet = NewMetricsCollector("get")
var MetricsPost = NewMetricsCollector("post")
var MetricsDelete = NewMetricsCollector("delete")

type MetricStats struct {
	Load       int
	TimeMS     int64
	BytesRead  int64
	BytesWrite int64
	Requests   int64
}

// No actual state. Just communication
type MetricsCollector struct {
	Name  string
	load1 chan int
	end   chan Metric
	read  chan chan MetricStats
}

// The client struct for bumping counters
type Metric struct {
	c          MetricsCollector
	beginMS    int64
	endMS      int64
	BytesRead  int64
	BytesWrite int64
	Requests   int64
}

// A new metrics collector
func NewMetricsCollector(name string) MetricsCollector {
	load1 := make(chan int)
	end := make(chan Metric)
	read := make(chan chan MetricStats)
	go func() {
		bytesRead := int64(0)
		bytesWrite := int64(0)
		observedTime := int64(0)
		requests := int64(0)
		load := 0
		for {
			select {
			case v := <-load1:
				if v == MetricsLoad {
					load++
				}
				if v == MetricsStop {
					break
				}
			case m := <-end:
				bytesRead += m.BytesRead
				bytesWrite += m.BytesWrite
				observedTime += m.endMS - m.beginMS
				load--
			case c := <-read:
				c <- MetricStats{
					Load:       load,
					TimeMS:     observedTime,
					BytesRead:  bytesRead,
					BytesWrite: bytesWrite,
					Requests:   requests,
				}
			}
		}
	}()
	return MetricsCollector{
		Name:  name,
		load1: load1,
		end:   end,
		read:  read,
	}
}

func (c MetricsCollector) Load() {
	c.load1 <- MetricsLoad
}

func (c MetricsCollector) Stop() {
	c.load1 <- MetricsStop
}

func (c MetricsCollector) NowMS() int64 {
	return time.Now().UnixNano() / 1000000
}

func (c MetricsCollector) Task() Metric {
	c.Load()
	t := c.NowMS()
	return Metric{
		c:       c,
		beginMS: t,
	}
}

func (m Metric) End() {
	m.endMS = m.c.NowMS()
	m.c.end <- m
}

func (c MetricsCollector) Stats() MetricStats {
	r := make(chan MetricStats)
	c.read <- r
	return <-r
}

func GetMetricsHandler(w http.ResponseWriter, r *http.Request, pathTokens []string) bool {
	GetMetricsWriter(w)
	return true
}

func GetMetricsWriter(w io.Writer) {
	writeObject := func(c MetricsCollector, s MetricStats, n string, t string, h string) {
		w.Write([]byte(fmt.Sprintf("# HELP microms_%s_%s %s\n", c.Name, n, h)))
		w.Write([]byte(fmt.Sprintf("# TYPE microms_%s_%s %s\n", c.Name, n, t)))
		w.Write([]byte(fmt.Sprintf("microms_%s_%s %d\n", c.Name, n, s.Requests)))
	}
	for _, c := range []MetricsCollector{MetricsGet, MetricsPost, MetricsDelete} {
		s := c.Stats()
		writeObject(
			c, s, "requests", "counter", "Total number of requests",
		)

		writeObject(
			c, s, "bytes_read", "counter", "Total number of bytes read",
		)

		writeObject(
			c, s, "bytes_write", "counter", "Total number of bytes write",
		)

		writeObject(
			c, s, "load", "gauge", "Current load",
		)

		writeObject(
			c, s, "time_ms", "counter", "Total time in milliseconds",
		)
	}
}
