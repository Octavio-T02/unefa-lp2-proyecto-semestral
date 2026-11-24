package domain

import (
	"errors"
	"regexp"
	"strings"
)

type EstadoTransaccion string

const (
	EstadoPendiente EstadoTransaccion = "PENDIENTE"
	EstadoAprobada  EstadoTransaccion = "APROBADA"
	EstadoRechazada EstadoTransaccion = "RECHAZADA"
)

var (
	ErrIDVacio               = errors.New("el id de transaccion es obligatorio")
	ErrIdempotencyKeyVacia   = errors.New("la idempotency key es obligatoria")
	ErrMSISDNInvalido        = errors.New("el msisdn no cumple con el estandar ITU-T E.164")
	ErrMontoInvalido         = errors.New("el monto en centavos debe ser mayor a cero")
	ErrTransicionEstadoInval = errors.New("transicion de estado no permitida")

	regexE164 = regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
)

type TransaccionRecarga struct {
	ID                string
	IdempotencyKey    string
	MSISDN            string
	MontoCentavos     int64
	Estado            EstadoTransaccion
	ReferenciaExterna string
}

func NuevaTransaccionRecarga(id, idempKey, msisdn string, montoCentavos int64) (TransaccionRecarga, error) {
	idLimpio := strings.TrimSpace(id)
	keyLimpia := strings.TrimSpace(idempKey)
	msisdnLimpio := strings.TrimSpace(msisdn)

	if idLimpio == "" {
		return TransaccionRecarga{}, ErrIDVacio
	}
	if keyLimpia == "" {
		return TransaccionRecarga{}, ErrIdempotencyKeyVacia
	}
	if !regexE164.MatchString(msisdnLimpio) {
		return TransaccionRecarga{}, ErrMSISDNInvalido
	}
	if montoCentavos <= 0 {
		return TransaccionRecarga{}, ErrMontoInvalido
	}

	return TransaccionRecarga{
		ID:             idLimpio,
		IdempotencyKey: keyLimpia,
		MSISDN:         msisdnLimpio,
		MontoCentavos:  montoCentavos,
		Estado:         EstadoPendiente,
	}, nil
}

func (t *TransaccionRecarga) Aprobar(referenciaExterna string) error {
	if t.Estado != EstadoPendiente {
		return ErrTransicionEstadoInval
	}
	t.Estado = EstadoAprobada
	t.ReferenciaExterna = strings.TrimSpace(referenciaExterna)
	return nil
}

func (t *TransaccionRecarga) Rechazar(motivo string) error {
	if t.Estado != EstadoPendiente {
		return ErrTransicionEstadoInval
	}
	t.Estado = EstadoRechazada
	t.ReferenciaExterna = strings.TrimSpace(motivo)
	return nil
}
