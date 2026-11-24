package application

import (
	"context"
	"errors"
	"fmt"

	"unefa-lp2-proyecto-semestral/lab2_arquitectura_hexagonal/internal/domain"
	"unefa-lp2-proyecto-semestral/lab2_arquitectura_hexagonal/internal/ports"
)

var ErrDependenciaNula = errors.New("las dependencias del servicio no pueden ser nulas")

type ComandoRecarga struct {
	IDTransaccion  string
	IdempotencyKey string
	MSISDN         string
	MontoCentavos  int64
}

type ServicioRecarga struct {
	repo      ports.RepositorioTransacciones
	pasarela  ports.PasarelaPagoBancaria
	auditoria ports.NotificadorAuditoria
}

func NuevoServicioRecarga(
	repo ports.RepositorioTransacciones,
	pasarela ports.PasarelaPagoBancaria,
	auditoria ports.NotificadorAuditoria,
) (*ServicioRecarga, error) {
	if repo == nil || pasarela == nil || auditoria == nil {
		return nil, ErrDependenciaNula
	}
	return &ServicioRecarga{
		repo:      repo,
		pasarela:  pasarela,
		auditoria: auditoria,
	}, nil
}

func (s *ServicioRecarga) EjecutarRecarga(ctx context.Context, cmd ComandoRecarga) (domain.TransaccionRecarga, bool, error) {
	existente, encontrada, err := s.repo.BuscarPorIdempotencyKey(ctx, cmd.IdempotencyKey)
	if err != nil {
		return domain.TransaccionRecarga{}, false, fmt.Errorf("error consultando idempotencia: %w", err)
	}
	if encontrada {
		_ = s.auditoria.RegistrarEvento(ctx, "IDEMPOTENCY_HIT_REINTENTO_IGNORADO", existente)
		return existente, true, nil
	}

	tx, err := domain.NuevaTransaccionRecarga(cmd.IDTransaccion, cmd.IdempotencyKey, cmd.MSISDN, cmd.MontoCentavos)
	if err != nil {
		return domain.TransaccionRecarga{}, false, err
	}

	refExterna, errDebito := s.pasarela.Debitar(ctx, tx.MSISDN, tx.MontoCentavos, tx.IdempotencyKey)
	if errDebito != nil {
		_ = tx.Rechazar(errDebito.Error())
		if errSave := s.repo.Guardar(ctx, tx); errSave != nil {
			return domain.TransaccionRecarga{}, false, errSave
		}
		_ = s.auditoria.RegistrarEvento(ctx, "RECARGA_RECHAZADA", tx)
		return tx, false, errDebito
	}

	if errAprobar := tx.Aprobar(refExterna); errAprobar != nil {
		return domain.TransaccionRecarga{}, false, errAprobar
	}

	if errSave := s.repo.Guardar(ctx, tx); errSave != nil {
		return domain.TransaccionRecarga{}, false, errSave
	}

	_ = s.auditoria.RegistrarEvento(ctx, "RECARGA_APROBADA", tx)
	return tx, false, nil
}
