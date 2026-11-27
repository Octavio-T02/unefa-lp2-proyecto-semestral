package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gocql/gocql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxPayload struct {
	MSISDN        string `json:"msisdn"`
	BucketFecha   string `json:"bucket_fecha"`
	Celda5G       string `json:"celda_5g"`
	BytesSubida   int64  `json:"bytes_subida"`
	BytesBajada   int64  `json:"bytes_bajada"`
	LatenciaMs    int    `json:"latencia_ms"`
	MontoCentavos int64  `json:"monto_centavos"`
}

func registrarConsumoACID(ctx context.Context, db *pgxpool.Pool, msisdn, idemKey string, montoCtv int64) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var abonadoID int64
	err = tx.QueryRow(ctx, `
		UPDATE abonados_bcnf
		SET saldo_centavos = saldo_centavos - $1, actualizado_en = NOW()
		WHERE msisdn = $2 AND estado_linea = 'ACTIVA' AND saldo_centavos >= $1
		RETURNING abonado_id
	`, montoCtv, msisdn).Scan(&abonadoID)
	if err != nil {
		return fmt.Errorf("no se pudo debitar abonado %s: %w", msisdn, err)
	}

	payload := OutboxPayload{
		MSISDN:        msisdn,
		BucketFecha:   time.Now().UTC().Format("2006-01-02"),
		Celda5G:       "CCS-NORTE-01",
		BytesSubida:   2048,
		BytesBajada:   16384,
		LatenciaMs:    14,
		MontoCentavos: montoCtv,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO transacciones_bss (abonado_id, idempotency_key, tipo_operacion, monto_centavos, metadata_jsonb)
		VALUES ($1, $2, 'CONSUMO_DATOS', $3, $4::jsonb)
	`, abonadoID, idemKey, montoCtv, string(payloadBytes))
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO outbox_events (aggregate_id, event_type, payload_jsonb, status)
		VALUES ($1, 'CONSUMO_DATOS_5G', $2::jsonb, 'PENDING')
	`, msisdn, string(payloadBytes))
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func procesarLoteOutbox(ctx context.Context, db *pgxpool.Pool, scylla *gocql.Session, workerID int) error {
	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT event_id, payload_jsonb
		FROM outbox_events
		WHERE status = 'PENDING'
		ORDER BY created_at
		LIMIT 50
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type eventoPendiente struct {
		id      int64
		payload []byte
	}
	var pendientes []eventoPendiente

	for rows.Next() {
		var ev eventoPendiente
		if err := rows.Scan(&ev.id, &ev.payload); err != nil {
			return err
		}
		pendientes = append(pendientes, ev)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, ev := range pendientes {
		var data OutboxPayload
		if err := json.Unmarshal(ev.payload, &data); err != nil {
			log.Printf("[Worker %d] Payload inválido en event_id=%d: %v", workerID, ev.id, err)
			continue
		}

		err := scylla.Query(`
			INSERT INTO cdr_por_abonado_dia
				(msisdn, bucket_fecha, timestamp_evento, id_cdr, celda_5g, bytes_subida, bytes_bajada, latencia_ms)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`,
			data.MSISDN,
			data.BucketFecha,
			time.Now().UTC(),
			gocql.TimeUUID(),
			data.Celda5G,
			data.BytesSubida,
			data.BytesBajada,
			data.LatenciaMs,
		).Consistency(gocql.LocalQuorum).Exec()

		if err != nil {
			log.Printf("[Worker %d] Fallo proyectando event_id=%d a ScyllaDB: %v", workerID, ev.id, err)
			continue
		}

		_, err = tx.Exec(ctx, `
			UPDATE outbox_events
			SET status = 'PROCESSED', processed_at = NOW()
			WHERE event_id = $1
		`, ev.id)
		if err != nil {
			return err
		}
		fmt.Printf("[Worker %d] Proyectado a ScyllaDB y marcado PROCESSED: event_id=%d (%s)\n", workerID, ev.id, data.MSISDN)
	}

	return tx.Commit(ctx)
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pgURL := "postgres://postgres:password123@127.0.0.1:5432/unefa_bss?sslmode=disable"
	db, err := pgxpool.New(ctx, pgURL)
	if err != nil {
		log.Fatalf("Error conectando a PostgreSQL: %v", err)
	}
	defer db.Close()

	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Port = 9042
	cluster.Keyspace = "unefa_telco_5g"
	cluster.Consistency = gocql.LocalQuorum
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	scylla, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("Error conectando a ScyllaDB: %v", err)
	}
	defer scylla.Close()

	// Simula 5 comandos ACID en PostgreSQL
	for i := 1; i <= 5; i++ {
		idemKey := fmt.Sprintf("outbox-demo-%d-%d", time.Now().UnixNano(), i)
		if err := registrarConsumoACID(ctx, db, "+584141112233", idemKey, 250); err != nil {
			log.Printf("Error registrando consumo %d: %v", i, err)
		}
	}

	// Pool de 4 workers despachando con FOR UPDATE SKIP LOCKED
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					if err := procesarLoteOutbox(ctx, db, scylla, id); err != nil {
						log.Printf("[Worker %d] Error: %v", id, err)
					}
					time.Sleep(500 * time.Millisecond)
				}
			}
		}(w)
	}

	wg.Wait()
}
