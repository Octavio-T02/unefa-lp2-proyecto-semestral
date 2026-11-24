package adapters

import (
	"context"
	"fmt"
	"sync"

	"unefa-lp2-proyecto-semestral/lab2_arquitectura_hexagonal/internal/domain"
)

type RepositorioMemoria struct {
	mu          sync.RWMutex
	porID       map[string]domain.TransaccionRecarga
	porIdempKey map[string]string
}

func NuevoRepositorioMemoria() *RepositorioMemoria {
	return &RepositorioMemoria{
		porID:       make(map[string]domain.TransaccionRecarga),
		porIdempKey: make(map[string]string),
	}
}

func (r *RepositorioMemoria) BuscarPorIdempotencyKey(_ context.Context, key string) (domain.TransaccionRecarga, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.porIdempKey[key]
	if !ok {
		return domain.TransaccionRecarga{}, false, nil
	}
	tx, existe := r.porID[id]
	return tx, existe, nil
}

func (r *RepositorioMemoria) Guardar(_ context.Context, tx domain.TransaccionRecarga) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.porID[tx.ID] = tx
	r.porIdempKey[tx.IdempotencyKey] = tx.ID
	return nil
}

type PasarelaBancariaSimulada struct {
	mu           sync.Mutex
	totalCobros  int
	montoCobrado int64
}

func NuevaPasarelaBancariaSimulada() *PasarelaBancariaSimulada {
	return &PasarelaBancariaSimulada{}
}

func (p *PasarelaBancariaSimulada) Debitar(_ context.Context, msisdn string, montoCentavos int64, idempotencyKey string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.totalCobros++
	p.montoCobrado += montoCentavos
	return fmt.Sprintf("BANCO-REF-%s-%04d", msisdn[len(msisdn)-4:], p.totalCobros), nil
}

func (p *PasarelaBancariaSimulada) ResumenCobros() (int, int64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.totalCobros, p.montoCobrado
}

type AuditoriaMemoria struct {
	mu      sync.Mutex
	eventos []string
}

func NuevaAuditoriaMemoria() *AuditoriaMemoria {
	return &AuditoriaMemoria{
		eventos: make([]string, 0),
	}
}

func (a *AuditoriaMemoria) RegistrarEvento(_ context.Context, evento string, tx domain.TransaccionRecarga) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	linea := fmt.Sprintf("[%s] id=%s key=%s msisdn=%s monto=%d estado=%s",
		evento, tx.ID, tx.IdempotencyKey, tx.MSISDN, tx.MontoCentavos, tx.Estado)
	a.eventos = append(a.eventos, linea)
	return nil
}
