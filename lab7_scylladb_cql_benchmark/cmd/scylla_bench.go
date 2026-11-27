package main

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/gocql/gocql"
)

func percentileUs(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}

func main() {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Port = 9042
	cluster.Keyspace = "unefa_telco_5g"
	cluster.Consistency = gocql.LocalQuorum
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	cluster.Timeout = 5 * time.Second

	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Error conectando a ScyllaDB: %v", err)
	}
	defer session.Close()

	const workers = 16
	const opsPerWorker = 250
	totalOps := workers * opsPerWorker

	latenciasUs := make([]float64, 0, totalOps)
	var mu sync.Mutex
	var wg sync.WaitGroup

	query := `INSERT INTO cdr_por_abonado_dia
		(msisdn, bucket_fecha, timestamp_evento, id_cdr, celda_5g, bytes_subida, bytes_bajada, latencia_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	inicioGlobal := time.Now()

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			localLatencias := make([]float64, 0, opsPerWorker)
			msisdn := fmt.Sprintf("+58414000%04d", workerID)
			bucket := time.Now().UTC().Format("2006-01-02")

			for i := 0; i < opsPerWorker; i++ {
				t0 := time.Now()
				err := session.Query(
					query,
					msisdn,
					bucket,
					time.Now().UTC(),
					gocql.TimeUUID(),
					fmt.Sprintf("CELL-5G-%d", workerID%8),
					int64(1024*(i+1)),
					int64(4096*(i+1)),
					12+(i%25),
				).Exec()
				elapsedUs := float64(time.Since(t0).Microseconds())
				if err != nil {
					log.Printf("Worker %d error en op %d: %v", workerID, i, err)
					continue
				}
				localLatencias = append(localLatencias, elapsedUs)
			}

			mu.Lock()
			latenciasUs = append(latenciasUs, localLatencias...)
			mu.Unlock()
		}(w)
	}

	wg.Wait()
	duracionTotal := time.Since(inicioGlobal).Seconds()

	sort.Float64s(latenciasUs)
	throughput := float64(len(latenciasUs)) / duracionTotal

	p50Us := percentileUs(latenciasUs, 0.50)
	p95Us := percentileUs(latenciasUs, 0.95)
	p99Us := percentileUs(latenciasUs, 0.99)

	fmt.Printf("=== BENCHMARK SCYLLADB (W=%d, Ops=%d) ===\n", workers, len(latenciasUs))
	fmt.Printf("Throughput : %.2f ops/s\n", throughput)
	fmt.Printf("P50        : %.2f us (%.3f ms)\n", p50Us, p50Us/1000.0)
	fmt.Printf("P95        : %.2f us (%.3f ms)\n", p95Us, p95Us/1000.0)
	fmt.Printf("P99        : %.2f us (%.3f ms)\n", p99Us, p99Us/1000.0)
}