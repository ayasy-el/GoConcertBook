package metrics

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

var (
	requestMu       sync.Mutex
	httpTotal       = map[string]uint64{}
	httpDurationSum = map[string]float64{}

	reservationSuccess atomic.Uint64
	reservationFailed  atomic.Uint64
	failedReservation  atomic.Uint64
	kafkaLagGauge      atomic.Int64
	redisMemoryGauge   atomic.Int64
	dbOpenConnGauge    atomic.Int64
)

func ObserveHTTP(method, path string, status int, duration time.Duration) {
	key := method + "|" + path + "|" + strconv.Itoa(status)
	durKey := method + "|" + path
	requestMu.Lock()
	httpTotal[key]++
	httpDurationSum[durKey] += duration.Seconds()
	requestMu.Unlock()
}

func IncReservationSuccess() { reservationSuccess.Add(1) }
func IncReservationFailed() {
	reservationFailed.Add(1)
	failedReservation.Add(1)
}
func SetKafkaLag(v int64)    { kafkaLagGauge.Store(v) }
func SetRedisMemory(v int64) { redisMemoryGauge.Store(v) }
func SetDBOpenConn(v int64)  { dbOpenConnGauge.Store(v) }

func Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	write(w,
		"# HELP reservation_total Total reservation requests\n",
		"# TYPE reservation_total counter\n",
		fmt.Sprintf("reservation_total{result=\"success\"} %d\n", reservationSuccess.Load()),
		fmt.Sprintf("reservation_total{result=\"failed\"} %d\n", reservationFailed.Load()),
		"# HELP failed_reservation_total Total failed reservation requests\n",
		"# TYPE failed_reservation_total counter\n",
		fmt.Sprintf("failed_reservation_total %d\n", failedReservation.Load()),
		"# HELP kafka_consumer_lag Kafka consumer lag\n",
		"# TYPE kafka_consumer_lag gauge\n",
		fmt.Sprintf("kafka_consumer_lag %d\n", kafkaLagGauge.Load()),
		"# HELP redis_memory_bytes Redis used memory in bytes\n",
		"# TYPE redis_memory_bytes gauge\n",
		fmt.Sprintf("redis_memory_bytes %d\n", redisMemoryGauge.Load()),
		"# HELP db_open_connections Database open connections\n",
		"# TYPE db_open_connections gauge\n",
		fmt.Sprintf("db_open_connections %d\n", dbOpenConnGauge.Load()),
	)

	requestMu.Lock()
	keys := make([]string, 0, len(httpTotal))
	durationKeys := make([]string, 0, len(httpDurationSum))
	for k := range httpTotal {
		keys = append(keys, k)
	}
	for k := range httpDurationSum {
		durationKeys = append(durationKeys, k)
	}
	sort.Strings(keys)
	sort.Strings(durationKeys)
	write(w, "# HELP http_requests_total Total HTTP requests\n", "# TYPE http_requests_total counter\n")
	for _, k := range keys {
		parts := strings.Split(k, "|")
		write(w, fmt.Sprintf("http_requests_total{method=\"%s\",path=\"%s\",status=\"%s\"} %d\n", parts[0], parts[1], parts[2], httpTotal[k]))
	}
	write(w, "# HELP http_request_duration_seconds HTTP request duration sum in seconds\n", "# TYPE http_request_duration_seconds counter\n")
	for _, k := range durationKeys {
		parts := strings.Split(k, "|")
		write(w, fmt.Sprintf("http_request_duration_seconds{method=\"%s\",path=\"%s\"} %.6f\n", parts[0], parts[1], httpDurationSum[k]))
	}
	requestMu.Unlock()
}

func write(w io.Writer, lines ...string) {
	for _, line := range lines {
		_, _ = io.WriteString(w, line)
	}
}

func StartInfraCollectors(db *sql.DB, redisClient *goredis.Client, interval time.Duration, stop <-chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			if db != nil {
				SetDBOpenConn(int64(db.Stats().OpenConnections))
			}
			if redisClient != nil {
				if info, err := redisClient.Info(context.Background(), "memory").Result(); err == nil {
					for _, line := range strings.Split(info, "\n") {
						if strings.HasPrefix(line, "used_memory:") {
							v := strings.TrimSpace(strings.TrimPrefix(line, "used_memory:"))
							n, _ := strconv.ParseInt(v, 10, 64)
							SetRedisMemory(n)
							break
						}
					}
				}
			}
		}
	}
}
