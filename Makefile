.PHONY: all infra-up infra-down lab1 lab2 lab3 lab4 lab5 lab6 lab7 lab8 clean

PG_URL := postgres://postgres:password123@localhost:5432/unefa_bss?sslmode=disable

all: lab1 lab2 lab5

infra-up:
	docker compose up -d

infra-down:
	docker compose down

lab1:
	gcc -Wall -Wextra -O0 lab1_memoria_y_complejidad/cdr_memory_profiler.c -o /tmp/cdr_memory_profiler && /tmp/cdr_memory_profiler
	go build -gcflags="-m" -o /tmp/escape_analyzer ./lab1_memoria_y_complejidad/escape_analyzer.go
	go run ./lab1_memoria_y_complejidad/escape_analyzer.go

lab2:
	go run ./lab2_arquitectura_hexagonal/cmd/api/main.go

lab3:
	@echo "=== Lab 3: Frontend Reactivo (React 19 / Svelte 5) ==="
	@ls -la lab3_frontend_reactivo/

lab4:
	@echo "=== Lab 4: Backend Políglota (Rust / Go / Node / Python / PHP) ==="
	@ls -la lab4_backend_poliglota/

lab5:
	go run -race ./lab5_ddd_grpc_y_resiliencia/cmd/bss_saga_engine.go

lab6:
	psql "$(PG_URL)" -f lab6_postgresql_bcnf_mvcc_indices/01_schema_bcnf_y_outbox.sql
	psql "$(PG_URL)" -f lab6_postgresql_bcnf_mvcc_indices/02_auditoria_mvcc_e_indices.sql

lab7:
	docker exec -i scylla-node1 cqlsh < lab7_scylladb_cql_benchmark/esquema_telco_5g.cql
	go run -race ./lab7_scylladb_cql_benchmark/cmd/scylla_bench.go

lab8:
	go run -race ./lab8_transactional_outbox_cqrs/cmd/polyglot_outbox_worker.go

clean:
	rm -f /tmp/cdr_memory_profiler /tmp/escape_analyzer escape_analyzer
