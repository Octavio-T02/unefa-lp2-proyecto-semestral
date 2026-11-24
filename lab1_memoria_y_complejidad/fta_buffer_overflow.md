# Análisis de Árbol de Fallos (FTA) e Ishikawa: Corrupción de `RIP` por Buffer Overflow en Stack

## 1. Diagrama de Ishikawa (Espina de Pescado)

```mermaid
flowchart LR
    subgraph Cod["Código y Memoria"]
        C1["Funciones inseguras (strcpy, gets, sprintf)"]
        C2["Sin verificación de límites en índices/tamaños"]
    end
    subgraph Comp["Compilación y Build"]
        B1["Stack Canary deshabilitado (-fno-stack-protector)"]
        B2["Sin ASLR / PIE activo"]
    end
    subgraph Proc["Datos y Entorno"]
        D1["Carga útil de entrada mayor al buffer en Stack"]
        D2["Falta de sanitización/validación de longitud"]
    end

    C1 --> E(("Corrupción de RIP\npor Stack Overflow"))
    C2 --> E
    B1 --> E
    B2 --> E
    D1 --> E
    D2 --> E
```

## 2. Árbol de Fallos (FTA — Compuertas `AND` / `OR`)

En una arquitectura x86_64 con pila que crece hacia direcciones bajas, sobrescribir la dirección de retorno (`RIP` guardado en el marco de pila) requiere que se escriba más allá del buffer local (`OR` entre fallos de escritura/longitud) **Y** que no se detenga antes de alcanzar `RIP` (`AND` con ausencia de mitigaciones/controles efectivos).

```mermaid
flowchart TD
    TOP["Evento Tope: Corrupción del registro RIP en Stack"] --> AND1{"AND"}

    AND1 --> OR1{"OR: Escritura fuera de límites en Stack"}
    AND1 --> OR2{"OR: Ausencia de mitigación efectiva"}

    OR1 --> F1["Copia sin límite de tamaño\n(ej. memcpy/strcpy con len > cap)"]
    OR1 --> F2["Desbordamiento de índice en arreglo local\n(off-by-many)"]

    OR2 --> F3["Stack Canary ausente o burlado\n(sin -fstack-protector)"]
    OR2 --> F4["Control de flujo / Guard Page no impide\nsobreescritura de frame"]
```