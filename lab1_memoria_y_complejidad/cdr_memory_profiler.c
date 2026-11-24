#include <inttypes.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

/* Variable global inicializada -> segmento .data */
static uint64_t g_contador_global_cdr = 1000ULL;

/* Campos intercalados: 1B, 8B, 1B, 4B, 1B, 2B -> padding interno y al final */
struct CDR_Desalineado {
    char tipo_red;          /* 1 byte */
    uint64_t imsi_abonado;  /* 8 bytes */
    char bandera_roaming;   /* 1 byte */
    uint32_t celda_id;      /* 4 bytes */
    char codigo_estado;     /* 1 byte */
    uint16_t puerto_origen; /* 2 bytes */
};

/* Ordenado por alineacion descendente: 8B, 4B, 2B, 1B, 1B, 1B */
struct CDR_Optimizado {
    uint64_t imsi_abonado;  /* 8 bytes */
    uint32_t celda_id;      /* 4 bytes */
    uint16_t puerto_origen; /* 2 bytes */
    char tipo_red;          /* 1 byte */
    char bandera_roaming;   /* 1 byte */
    char codigo_estado;     /* 1 byte */
};

int main(void) {
    /* Variable local -> Stack */
    struct CDR_Optimizado cdr_stack = {
        .imsi_abonado = 734010012345678ULL,
        .celda_id = 4096U,
        .puerto_origen = 5060U,
        .tipo_red = '5',
        .bandera_roaming = 'N',
        .codigo_estado = 'A'
    };

    /* Bloque dinamico -> Heap */
    struct CDR_Optimizado *cdr_heap = malloc(sizeof(*cdr_heap));
    if (cdr_heap == NULL) {
        fprintf(stderr, "Error: fallo en malloc\n");
        return EXIT_FAILURE;
    }
    *cdr_heap = cdr_stack;

    printf("=== (a) DIRECCIONES VIRTUALES DE MEMORIA ===\n");
    printf("Global (.data) : %p\n", (void *)&g_contador_global_cdr);
    printf("Local (Stack)  : %p\n", (void *)&cdr_stack);
    printf("Dinamico (Heap): %p\n\n", (void *)cdr_heap);

    /* Liberar memoria dinamica sin dejar leaks */
    free(cdr_heap);
    cdr_heap = NULL;

    size_t sz_des = sizeof(struct CDR_Desalineado);
    size_t sz_opt = sizeof(struct CDR_Optimizado);

    printf("=== (b) TAMANO Y OFFSETS DE ESTRUCTURAS ===\n");
    printf("sizeof(CDR_Desalineado) = %zu bytes\n", sz_des);
    printf("  offset tipo_red        : %zu\n", offsetof(struct CDR_Desalineado, tipo_red));
    printf("  offset imsi_abonado    : %zu\n", offsetof(struct CDR_Desalineado, imsi_abonado));
    printf("  offset bandera_roaming : %zu\n", offsetof(struct CDR_Desalineado, bandera_roaming));
    printf("  offset celda_id        : %zu\n", offsetof(struct CDR_Desalineado, celda_id));
    printf("  offset codigo_estado   : %zu\n", offsetof(struct CDR_Desalineado, codigo_estado));
    printf("  offset puerto_origen   : %zu\n", offsetof(struct CDR_Desalineado, puerto_origen));

    printf("\nsizeof(CDR_Optimizado)  = %zu bytes\n", sz_opt);
    printf("  offset imsi_abonado    : %zu\n", offsetof(struct CDR_Optimizado, imsi_abonado));
    printf("  offset celda_id        : %zu\n", offsetof(struct CDR_Optimizado, celda_id));
    printf("  offset puerto_origen   : %zu\n", offsetof(struct CDR_Optimizado, puerto_origen));
    printf("  offset tipo_red        : %zu\n", offsetof(struct CDR_Optimizado, tipo_red));
    printf("  offset bandera_roaming : %zu\n", offsetof(struct CDR_Optimizado, bandera_roaming));
    printf("  offset codigo_estado   : %zu\n\n", offsetof(struct CDR_Optimizado, codigo_estado));

    /* (c) Impacto en 10,000,000 de registros y Cache Line L1 (64 bytes) */
    const double lote = 10000000.0;
    const size_t cache_line_l1 = 64U;

    double ram_des_mb = (lote * (double)sz_des) / 1000000.0;
    double ram_opt_mb = (lote * (double)sz_opt) / 1000000.0;
    double ahorro_bytes = lote * (double)(sz_des > sz_opt ? (sz_des - sz_opt) : 0U);
    double ahorro_mib = ahorro_bytes / (1024.0 * 1024.0);

    size_t caben_des = cache_line_l1 / sz_des;
    size_t caben_opt = cache_line_l1 / sz_opt;

    printf("=== (c) LOTE DE 10,000,000 REGISTROS Y CACHE L1 (64B) ===\n");
    printf("RAM Desalineado        : %.0f bytes (%.2f MB decimal)\n", lote * (double)sz_des, ram_des_mb);
    printf("RAM Optimizado         : %.0f bytes (%.2f MB decimal)\n", lote * (double)sz_opt, ram_opt_mb);
    printf("Ahorro exacto          : %.0f bytes (%.2f MiB binarios)\n", ahorro_bytes, ahorro_mib);
    printf("Registros por L1 (64B) - Desalineado: %zu\n", caben_des);
    printf("Registros por L1 (64B) - Optimizado : %zu\n", caben_opt);

    return EXIT_SUCCESS;
}