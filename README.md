# Proyecto Semestral — Lenguajes de Programacion II (UNEFA)

Sistema distribuido de facturacion y telemetria de telecomunicaciones (BSS/5G) desarrollado de forma incremental a lo largo de la asignatura.

## Estructura del Repositorio

- `lab1_memoria_y_complejidad/` — Perfilado de memoria en C y Escape Analysis en Go.
- `lab2_arquitectura_hexagonal/` — Pasarela de recarga con arquitectura hexagonal y contrato OpenAPI 3.1.
- `lab3_frontend_reactivo/` — Paneles reactivos (React 19 / Svelte 5) y Web Vitals.
- `lab4_backend_poliglota/` — Modulos concurrentes poliglotas (Rust, Go, Node.js, Python y PHP).
- `lab5_ddd_grpc_y_resiliencia/` — Motor Saga en Go, Circuit Breaker, Jitter y contrato gRPC (`billing_bss.proto`).
- `lab6_postgresql_bcnf_mvcc_indices/` — Esquema BCNF, Outbox, auditoria MVCC e indices.
- `lab7_scylladb_cql_benchmark/` — Esquema CQL 5G y benchmark en Go para ScyllaDB.
- `lab8_transactional_outbox_cqrs/` — Worker de Transactional Outbox y CQRS.

## Comandos Principales

- `make infra-up` / `make infra-down`: Levantar o detener Docker Compose.
- `make lab1`, `make lab2`, `make lab5`, `make lab6`, `make lab7`, `make lab8`: Ejecutar cada laboratorio.
- `./run_all_checks.sh`: Verificar compilacion y estado del repositorio.
