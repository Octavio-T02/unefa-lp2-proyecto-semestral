package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Transaccion struct {
	ID            string
	MSISDN        string
	MontoCentavos int64
	PaisOrigen    string
}

type ResultadoAuditoria struct {
	IDTransaccion string
	Aprobada      bool
	RiesgoScore   int
	WorkerID      int
}

type ProcesadorTransaccion interface {
	Procesar(tx Transaccion, workerID int) ResultadoAuditoria
}

type MotorReglasFraude struct {
	LimiteCentavos int64
}

func (m MotorReglasFraude) Procesar(tx Transaccion, workerID int) ResultadoAuditoria {
	riesgo := 10
	if tx.MontoCentavos > m.LimiteCentavos {
		riesgo += 65
	}
	if tx.PaisOrigen != "VE" {
		riesgo += 20
	}
	return ResultadoAuditoria{
		IDTransaccion: tx.ID,
		Aprobada:      riesgo < 70,
		RiesgoScore:   riesgo,
		WorkerID:      workerID,
	}
}

func EjecutarWorkerPool(
	numWorkers int,
	procesador ProcesadorTransaccion,
	jobs <-chan Transaccion,
	results chan<- ResultadoAuditoria,
	aprobadas *atomic.Int64,
	rechazadas *atomic.Int64,
) {
	var wg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for tx := range jobs {
				res := procesador.Procesar(tx, workerID)
				if res.Aprobada {
					aprobadas.Add(1)
				} else {
					rechazadas.Add(1)
				}
				results <- res
			}
		}(w)
	}

	go func() {
		wg.Wait()
		close(results)
	}()
}

func main() {
	const workers = 4
	const totalTx = 24

	jobs := make(chan Transaccion, totalTx)
	results := make(chan ResultadoAuditoria, totalTx)

	var aprobadas atomic.Int64
	var rechazadas atomic.Int64

	motor := MotorReglasFraude{LimiteCentavos: 25_000}
	EjecutarWorkerPool(workers, motor, jobs, results, &aprobadas, &rechazadas)

	go func() {
		for i := 1; i <= totalTx; i++ {
			pais := "VE"
			if i%6 == 0 {
				pais = "US"
			}
			monto := int64(2_000 * i)
			jobs <- Transaccion{
				ID:            fmt.Sprintf("OP-%03d", i),
				MSISDN:        fmt.Sprintf("+58412%07d", 500_000+i),
				MontoCentavos: monto,
				PaisOrigen:    pais,
			}
		}
		close(jobs)
	}()

	procesadas := 0
	for res := range results {
		procesadas++
		if procesadas <= 6 || !res.Aprobada {
			fmt.Printf("Worker #%d -> tx=%s aprobada=%v score=%d\n",
				res.WorkerID, res.IDTransaccion, res.Aprobada, res.RiesgoScore)
		}
	}

	fmt.Printf("\nResumen Worker Pool -> total=%d aprobadas=%d rechazadas=%d\n",
		procesadas, aprobadas.Load(), rechazadas.Load())
}
