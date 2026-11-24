package ports

import (
	"context"

	"unefa-lp2-proyecto-semestral/lab2_arquitectura_hexagonal/internal/domain"
)

type RepositorioTransacciones interface {
	BuscarPorIdempotencyKey(ctx context.Context, key string) (domain.TransaccionRecarga, bool, error)
	Guardar(ctx context.Context, tx domain.TransaccionRecarga) error
}

type PasarelaPagoBancaria interface {
	Debitar(ctx context.Context, msisdn string, montoCentavos int64, idempotencyKey string) (string, error)
}

type NotificadorAuditoria interface {
	RegistrarEvento(ctx context.Context, evento string, tx domain.TransaccionRecarga) error
}
