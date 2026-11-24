package main

import (
	"context"
	"fmt"
	"log"

	"unefa-lp2-proyecto-semestral/lab2_arquitectura_hexagonal/internal/adapters"
	"unefa-lp2-proyecto-semestral/lab2_arquitectura_hexagonal/internal/application"
)

func main() {
	ctx := context.Background()

	repo := adapters.NuevoRepositorioMemoria()
	pasarela := adapters.NuevaPasarelaBancariaSimulada()
	auditoria := adapters.NuevaAuditoriaMemoria()

	servicio, err := application.NuevoServicioRecarga(repo, pasarela, auditoria)
	if err != nil {
		log.Fatalf("error inicializando servicio: %v", err)
	}

	// (a) Recarga exitosa
	txExitosa, repetidaA, errA := servicio.EjecutarRecarga(ctx, application.ComandoRecarga{
		IDTransaccion:  "TX-001",
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
		MSISDN:         "+584121234567",
		MontoCentavos:  500,
	})
	fmt.Printf("(a) Recarga exitosa -> err=%v repetida=%v estado=%s ref=%s\n",
		errA, repetidaA, txExitosa.Estado, txExitosa.ReferenciaExterna)

	// (b) Rechazo de numero que viola ITU-T E.164 (falta '+' inicial)
	_, repetidaB, errB := servicio.EjecutarRecarga(ctx, application.ComandoRecarga{
		IDTransaccion:  "TX-002",
		IdempotencyKey: "550e8400-e29b-41d4-a716-446655440002",
		MSISDN:         "04121234567",
		MontoCentavos:  500,
	})
	fmt.Printf("(b) Numero invalido E.164 -> err=%v repetida=%v\n", errB, repetidaB)

	// (c) Tres reintentos de red con la misma Idempotency-Key -> cobra una sola vez
	claveReintento := "550e8400-e29b-41d4-a716-446655440003"
	for intento := 1; intento <= 3; intento++ {
		txIntento, repetidaC, errC := servicio.EjecutarRecarga(ctx, application.ComandoRecarga{
			IDTransaccion:  fmt.Sprintf("TX-003-INT-%d", intento),
			IdempotencyKey: claveReintento,
			MSISDN:         "+584149876543",
			MontoCentavos:  1200,
		})
		fmt.Printf("(c) Intento #%d -> id_guardado=%s repetida=%v estado=%s err=%v\n",
			intento, txIntento.ID, repetidaC, txIntento.Estado, errC)
	}

	totalCobros, montoTotal := pasarela.ResumenCobros()
	fmt.Printf("\nResumen real en pasarela bancaria -> cobros_ejecutados=%d, monto_total_centavos=%d\n",
		totalCobros, montoTotal)
}
