#!/usr/bin/env bash
set -euo pipefail

echo "=== Verificando repositorio LP2 - UNEFA ==="

if [ -d "lab1_memoria_y_complejidad" ]; then
    echo "[Lab 1] Compilando C y Go..."
    gcc -Wall -Wextra -O0 lab1_memoria_y_complejidad/cdr_memory_profiler.c -o /tmp/cdr_memory_profiler
    /tmp/cdr_memory_profiler
    go build -gcflags="-m" -o /tmp/escape_analyzer ./lab1_memoria_y_complejidad/escape_analyzer.go
    rm -f /tmp/cdr_memory_profiler /tmp/escape_analyzer
fi

if [ -d "lab2_arquitectura_hexagonal" ]; then
    echo "[Lab 2] Verificando compilacion..."
    go build -o /tmp/lab2_api ./lab2_arquitectura_hexagonal/cmd/api/main.go
    rm -f /tmp/lab2_api
fi

if [ -d "lab5_ddd_grpc_y_resiliencia" ]; then
    echo "[Lab 5] Verificando compilacion..."
    go build -o /tmp/lab5_saga ./lab5_ddd_grpc_y_resiliencia/cmd/bss_saga_engine.go
    rm -f /tmp/lab5_saga
fi


if [ -d "lab8_transactional_outbox_cqrs" ]; then
    echo "[Lab 8] Verificando compilacion..."
    go build -o /tmp/lab8_worker ./lab8_transactional_outbox_cqrs/cmd/polyglot_outbox_worker.go
    rm -f /tmp/lab8_worker
fi

echo "=== Verificacion completada ==="
