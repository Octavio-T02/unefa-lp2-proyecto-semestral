.PHONY: all lab1 lab2 lab3 lab4 clean

all: lab1 lab2 lab4

lab1:
	gcc -Wall -Wextra -Werror -O0 lab1_memoria_y_complejidad/cdr_memory_profiler.c -o lab1_memoria_y_complejidad/cdr_profiler && ./lab1_memoria_y_complejidad/cdr_profiler
	go build -gcflags="-m" lab1_memoria_y_complejidad/escape_analyzer.go
	go run lab1_memoria_y_complejidad/escape_analyzer.go

lab2:
	cd lab2_arquitectura_hexagonal && go vet ./... && go run ./cmd/api/main.go

lab3:
	@echo "Archivos de frontend listos en lab3_frontend_reactivo/nextjs_react19_vdom y lab3_frontend_reactivo/svelte5_runes_aot"

lab4:
	go run -race lab4_backend_poliglota/go_worker_pool/worker_pool.go
	@if command -v rustc >/dev/null 2>&1; then \
		rustc -O lab4_backend_poliglota/rust_antifraude/motor_antifraude.rs -o lab4_backend_poliglota/rust_antifraude/motor_antifraude && ./lab4_backend_poliglota/rust_antifraude/motor_antifraude; \
	fi
	@if command -v node >/dev/null 2>&1; then \
		node lab4_backend_poliglota/runtimes_alto_nivel/node_crypto_worker.mjs; \
	fi
	@if command -v python3 >/dev/null 2>&1; then \
		python3 lab4_backend_poliglota/runtimes_alto_nivel/python_async_validator.py; \
	fi
	@if command -v php >/dev/null 2>&1; then \
		php lab4_backend_poliglota/runtimes_alto_nivel/php_jit_tariff_engine.php; \
	fi

clean:
	rm -f lab1_memoria_y_complejidad/cdr_profiler lab1_memoria_y_complejidad/escape_analyzer lab4_backend_poliglota/rust_antifraude/motor_antifraude
