BEGIN;

CREATE TABLE IF NOT EXISTS catalogo_planes_bcnf (
    plan_id             VARCHAR(32) PRIMARY KEY,
    nombre_plan         VARCHAR(100) NOT NULL UNIQUE,
    cuota_mensual_ctv   BIGINT NOT NULL CHECK (cuota_mensual_ctv >= 0),
    datos_incluidos_mb  BIGINT NOT NULL CHECK (datos_incluidos_mb >= 0),
    creado_en           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS abonados_bcnf (
    abonado_id          BIGSERIAL PRIMARY KEY,
    msisdn              VARCHAR(16) NOT NULL UNIQUE CHECK (msisdn ~ '^\+[1-9][0-9]{7,14}$'),
    plan_id             VARCHAR(32) NOT NULL REFERENCES catalogo_planes_bcnf(plan_id),
    saldo_centavos      BIGINT NOT NULL DEFAULT 0 CHECK (saldo_centavos >= 0),
    estado_linea        VARCHAR(20) NOT NULL DEFAULT 'ACTIVA' CHECK (estado_linea IN ('ACTIVA', 'SUSPENDIDA', 'CANCELADA')),
    actualizado_en      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transacciones_bss (
    transaccion_id      BIGSERIAL PRIMARY KEY,
    abonado_id          BIGINT NOT NULL REFERENCES abonados_bcnf(abonado_id),
    idempotency_key     VARCHAR(128) NOT NULL UNIQUE,
    tipo_operacion      VARCHAR(20) NOT NULL CHECK (tipo_operacion IN ('RECARGA', 'CONSUMO_DATOS', 'AJUSTE')),
    monto_centavos      BIGINT NOT NULL CHECK (monto_centavos > 0),
    metadata_jsonb      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS outbox_events (
    event_id            BIGSERIAL PRIMARY KEY,
    aggregate_id        VARCHAR(64) NOT NULL,
    event_type          VARCHAR(64) NOT NULL,
    payload_jsonb       JSONB NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSED', 'FAILED')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at        TIMESTAMPTZ
);

-- Semilla inicial de planes y abonados
INSERT INTO catalogo_planes_bcnf (plan_id, nombre_plan, cuota_mensual_ctv, datos_incluidos_mb)
VALUES
    ('PLAN-5G-PRO', 'Plan 5G Ilimitado Pro', 2500, 102400),
    ('PLAN-5G-BASE', 'Plan 5G Esencial', 1200, 30720)
ON CONFLICT (plan_id) DO NOTHING;

INSERT INTO abonados_bcnf (msisdn, plan_id, saldo_centavos, estado_linea)
VALUES
    ('+584141112233', 'PLAN-5G-PRO', 10000, 'ACTIVA'),
    ('+584129998877', 'PLAN-5G-BASE', 5000, 'ACTIVA')
ON CONFLICT (msisdn) DO NOTHING;

COMMIT;