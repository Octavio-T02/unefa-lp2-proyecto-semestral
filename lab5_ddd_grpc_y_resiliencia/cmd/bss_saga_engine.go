package main

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

// --- VALUE OBJECT INMUTABLE: DINERO ---

type Dinero struct {
	MontoCentavos int64
	Moneda        string
}

func NuevoDinero(centavos int64, moneda string) (Dinero, error) {
	if moneda == "" {
		return Dinero{}, errors.New("la moneda es obligatoria")
	}
	if centavos < 0 {
		return Dinero{}, errors.New("un monto inicial de Dinero no puede ser negativo")
	}
	return Dinero{MontoCentavos: centavos, Moneda: moneda}, nil
}

func (d Dinero) Sumar(otro Dinero) (Dinero, error) {
	if d.Moneda != otro.Moneda {
		return Dinero{}, errors.New("no se pueden sumar monedas distintas")
	}
	if otro.MontoCentavos < 0 {
		return Dinero{}, errors.New("el monto a sumar no puede ser negativo")
	}
	return Dinero{
		MontoCentavos: d.MontoCentavos + otro.MontoCentavos,
		Moneda:        d.Moneda,
	}, nil
}

func (d Dinero) Restar(otro Dinero) (Dinero, error) {
	if d.Moneda != otro.Moneda {
		return Dinero{}, errors.New("no se pueden restar monedas distintas")
	}
	if otro.MontoCentavos < 0 {
		return Dinero{}, errors.New("el monto a restar no puede ser negativo")
	}
	nuevoSaldo := d.MontoCentavos - otro.MontoCentavos
	if nuevoSaldo < 0 {
		return Dinero{}, errors.New("saldo insuficiente: no se permiten saldos negativos")
	}
	return Dinero{
		MontoCentavos: nuevoSaldo,
		Moneda:        d.Moneda,
	}, nil
}

// --- EVENTO DE DOMINIO INMUTABLE ---

type CargoAplicadoEvent struct {
	EventoID       string
	SuscripcionID  string
	IdempotencyKey string
	MontoCentavos  int64
	Moneda         string
	SaldoRestante  int64
	OcurridoEn     time.Time
}

// --- RAÍZ DE AGREGADO: SUSCRIPCION 5G ---

type Suscripcion5G struct {
	ID      string
	MSISDN  string
	Activa  bool
	Saldo   Dinero
	Eventos []CargoAplicadoEvent
}

func (s *Suscripcion5G) AplicarCargo(idempotencyKey string, cargo Dinero) (CargoAplicadoEvent, error) {
	if !s.Activa {
		return CargoAplicadoEvent{}, errors.New("la suscripción 5G no está activa")
	}
	nuevoSaldo, err := s.Saldo.Restar(cargo)
	if err != nil {
		return CargoAplicadoEvent{}, err
	}
	s.Saldo = nuevoSaldo

	evento := CargoAplicadoEvent{
		EventoID:       fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		SuscripcionID:  s.ID,
		IdempotencyKey: idempotencyKey,
		MontoCentavos:  cargo.MontoCentavos,
		Moneda:         cargo.Moneda,
		SaldoRestante:  s.Saldo.MontoCentavos,
		OcurridoEn:     time.Now().UTC(),
	}
	s.Eventos = append(s.Eventos, evento)
	return evento, nil
}

func (s *Suscripcion5G) ReembolsarCargo(reembolso Dinero) error {
	nuevoSaldo, err := s.Saldo.Sumar(reembolso)
	if err != nil {
		return err
	}
	s.Saldo = nuevoSaldo
	return nil
}

// --- CIRCUIT BREAKER THREAD-SAFE ---

type CBState string

const (
	StateClosed   CBState = "CLOSED"
	StateOpen     CBState = "OPEN"
	StateHalfOpen CBState = "HALF-OPEN"
)

type CircuitBreaker struct {
	mu           sync.Mutex
	state        CBState
	fallos       int
	maxFallos    int
	timeoutReset time.Duration
	abiertoEn    time.Time
}

func NuevoCircuitBreaker(maxFallos int, timeoutReset time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:        StateClosed,
		maxFallos:    maxFallos,
		timeoutReset: timeoutReset,
	}
}

func (cb *CircuitBreaker) Ejecutar(fn func() error) error {
	cb.mu.Lock()
	switch cb.state {
	case StateOpen:
		if time.Since(cb.abiertoEn) >= cb.timeoutReset {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker OPEN: petición rechazada")
		}
	}
	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.fallos++
		if cb.state == StateHalfOpen || cb.fallos >= cb.maxFallos {
			cb.state = StateOpen
			cb.abiertoEn = time.Now()
		}
		return err
	}

	cb.fallos = 0
	cb.state = StateClosed
	return nil
}

func (cb *CircuitBreaker) Estado() CBState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// --- EXPONENTIAL BACKOFF + FULL JITTER ---

func BackoffFullJitter(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	baseMs := 50.0
	maxMs := 2000.0
	exp := math.Pow(2, float64(attempt))
	capMs := math.Min(maxMs, baseMs*exp)
	jitterMs := rand.Float64() * capMs
	return time.Duration(jitterMs) * time.Millisecond
}

// --- SAGA ORCHESTRATOR (T1 -> T2 -> T3 Y COMPENSACIONES C2, C1) ---

type SagaStep struct {
	Nombre     string
	Accion     func(idempotencyKey string) error
	Compensar  func(idempotencyKey string) error
	Ejecutado  bool
}

type SagaOrchestrator struct {
	pasos          []SagaStep
	compensaciones map[string]bool
	mu             sync.Mutex
}

func NuevoSagaOrchestrator(pasos []SagaStep) *SagaOrchestrator {
	return &SagaOrchestrator{
		pasos:          pasos,
		compensaciones: make(map[string]bool),
	}
}

func (s *SagaOrchestrator) Ejecutar(idempotencyKey string) error {
	ejecutados := make([]int, 0, len(s.pasos))

	for i := range s.pasos {
		stepKey := fmt.Sprintf("%s:step:%d", idempotencyKey, i)
		if err := s.pasos[i].Accion(stepKey); err != nil {
			fmt.Printf("[SAGA] Fallo en %s (%s): %v\n", s.pasos[i].Nombre, stepKey, err)
			s.compensarEnReversa(idempotencyKey, ejecutados)
			return err
		}
		ejecutados = append(ejecutados, i)
	}
	return nil
}

func (s *SagaOrchestrator) compensarEnReversa(idempotencyKey string, ejecutados []int) {
	for idx := len(ejecutados) - 1; idx >= 0; idx-- {
		stepIdx := ejecutados[idx]
		compKey := fmt.Sprintf("%s:compensate:%d", idempotencyKey, stepIdx)

		s.mu.Lock()
		if s.compensaciones[compKey] {
			s.mu.Unlock()
			continue
		}
		s.compensaciones[compKey] = true
		s.mu.Unlock()

		if s.pasos[stepIdx].Compensar != nil {
			if err := s.pasos[stepIdx].Compensar(compKey); err != nil {
				fmt.Printf("[SAGA] Error compensando %s (%s): %v\n", s.pasos[stepIdx].Nombre, compKey, err)
			}
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	saldoInicial, _ := NuevoDinero(5000, "USD") // $50.00 en centavos
	sub := &Suscripcion5G{
		ID:     "sub-5g-001",
		MSISDN: "+584141234567",
		Activa: true,
		Saldo:  saldoInicial,
	}

	cb := NuevoCircuitBreaker(2, 200*time.Millisecond)
	for i := 0; i < 3; i++ {
		_ = cb.Ejecutar(func() error { return errors.New("timeout OCS 5G") })
		fmt.Printf("Intento %d -> Backoff Full Jitter: %v | Estado CB: %s\n", i, BackoffFullJitter(i), cb.Estado())
	}

	// Simulamos fallo en T2 (Aprovisionar QoS 5G) para activar C1 (Reembolsar Billetera)
	montoCargo, _ := NuevoDinero(1500, "USD") // $15.00
	saga := NuevoSagaOrchestrator([]SagaStep{
		{
			Nombre: "T1: Debitar Billetera",
			Accion: func(key string) error {
				ev, err := sub.AplicarCargo(key, montoCargo)
				if err != nil {
					return err
				}
				fmt.Printf("[T1 OK] Key=%s | Cargo=%d centavos | Saldo actual=%d\n", key, ev.MontoCentavos, sub.Saldo.MontoCentavos)
				return nil
			},
			Compensar: func(key string) error {
				if err := sub.ReembolsarCargo(montoCargo); err != nil {
					return err
				}
				fmt.Printf("[C1 COMPENSADO] Key=%s | Saldo restaurado=%d centavos\n", key, sub.Saldo.MontoCentavos)
				return nil
			},
		},
		{
			Nombre: "T2: Aprovisionar QoS 5G",
			Accion: func(key string) error {
				return errors.New("fallo de red en función PCF/UPF 5G")
			},
			Compensar: func(key string) error {
				fmt.Printf("[C2 COMPENSADO] Key=%s | Liberando perfil QoS 5G\n", key)
				return nil
			},
		},
		{
			Nombre: "T3: Emitir Factura",
			Accion: func(key string) error {
				fmt.Printf("[T3 OK] Key=%s | Factura emitida\n", key)
				return nil
			},
		},
	})

	_ = saga.Ejecutar("idem-tx-9001")
}