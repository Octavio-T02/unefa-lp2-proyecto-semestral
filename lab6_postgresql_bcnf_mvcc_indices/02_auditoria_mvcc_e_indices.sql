-- 1. Carga masiva de 50,000 transacciones con generate_series
INSERT INTO transacciones_bss (abonado_id, idempotency_key, tipo_operacion, monto_centavos, metadata_jsonb, created_at)
SELECT
    1,
    'seed-tx-' || g::text,
    'CONSUMO_DATOS',
    (g % 900) + 100,
    jsonb_build_object(
        'antena_5g', CASE WHEN g % 250 = 0 THEN 'CCS-NORTE-01' ELSE 'CCS-SUR-' || (g % 20)::text END,
        'roaming', (g % 250 = 0),
        'celda', 'NR-CELL-' || (g % 50)::text
    ),
    NOW() - ((50000 - g) || ' seconds')::interval
FROM generate_series(1, 50000) AS g
ON CONFLICT (idempotency_key) DO NOTHING;

-- 2. Inspección forense MVCC antes y después de un UPDATE
SELECT ctid, xmin, xmax, abonado_id, msisdn, saldo_centavos
FROM abonados_bcnf
WHERE msisdn = '+584141112233';

BEGIN;
UPDATE abonados_bcnf
SET saldo_centavos = saldo_centavos + 500, actualizado_en = NOW()
WHERE msisdn = '+584141112233';

SELECT ctid, xmin, xmax, abonado_id, msisdn, saldo_centavos
FROM abonados_bcnf
WHERE msisdn = '+584141112233';
COMMIT;

-- 3. Plan de ejecución ANTES de crear el índice GIN
EXPLAIN (ANALYZE, BUFFERS)
SELECT transaccion_id, abonado_id, monto_centavos, metadata_jsonb
FROM transacciones_bss
WHERE metadata_jsonb @> '{"antena_5g": "CCS-NORTE-01", "roaming": true}';

-- 4. Creación de índices GIN (jsonb_path_ops), BRIN y B-Tree
CREATE INDEX IF NOT EXISTS idx_tx_metadata_gin
    ON transacciones_bss USING GIN (metadata_jsonb jsonb_path_ops);

CREATE INDEX IF NOT EXISTS idx_tx_created_at_brin
    ON transacciones_bss USING BRIN (created_at);

CREATE INDEX IF NOT EXISTS idx_outbox_pending_created_at
    ON outbox_events (created_at)
    WHERE status = 'PENDING';

ANALYZE transacciones_bss;

-- 5. Plan de ejecución DESPUÉS de crear el índice GIN
EXPLAIN (ANALYZE, BUFFERS)
SELECT transaccion_id, abonado_id, monto_centavos, metadata_jsonb
FROM transacciones_bss
WHERE metadata_jsonb @> '{"antena_5g": "CCS-NORTE-01", "roaming": true}';