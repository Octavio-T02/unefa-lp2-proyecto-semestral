# Reporte de Auditoría MVCC y Análisis `EXPLAIN (ANALYZE, BUFFERS)`

## 1. Comportamiento MVCC (`ctid`, `xmin`, `xmax`)
Al ejecutar un `UPDATE` sobre `abonados_bcnf`, PostgreSQL no sobreescribe el registro físico existente:
- La versión previa mantiene su `ctid` original y recibe en `xmax` el `xid` de la transacción que lo actualizó (marcándolo como muerto/obsoleto para transacciones posteriores).
- Se inserta una nueva versión de tupla con un nuevo `ctid` (por ejemplo, pasando de `(0,1)` a `(0,2)`) y con `xmin` igual al identificador de la transacción actual.

## 2. Comparativa antes y después del índice `GIN (metadata_jsonb jsonb_path_ops)`
| Métrica | Antes del índice GIN | Después del índice GIN |
| :--- | :--- | :--- |
| **Tipo de escaneo** | `Seq Scan` sobre `transacciones_bss` | `Bitmap Index Scan` + `Bitmap Heap Scan` |
| **Lectura de buffers (`shared hit` / `read`)** | Lee todas las páginas de la tabla completa | Lee solo las páginas del índice y las páginas exactas del heap |
| **Tiempo de ejecución** | Mayor (proporcional al total de filas y tamaño JSONB) | Drásticamente menor al filtrar por rutas/valores específicos |

## 3. Justificación de índices `BRIN` y `B-Tree`
- **`BRIN` en `created_at`:** Muy compacto y eficiente para columnas con correlación temporal creciente como timestamps de inserción.
- **`B-Tree` parcial en `outbox_events` (`WHERE status = 'PENDING'`):** Mantiene un índice pequeño y rápido exclusivamente para los eventos pendientes de despachar.