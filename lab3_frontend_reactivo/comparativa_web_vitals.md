# Auditoría Comparativa Web Vitals: Next.js 15 / React 19 (VDOM Fiber) vs. Svelte 5 (Runes AOT)

## 1. Cálculo del Presupuesto de Cuadro (Frame Budget)

El tiempo disponible por cuadro en función de la tasa de refresco $f$ (en Hz) es:

$$T_{\text{frame}} = \frac{1000\text{ ms}}{f}$$

* **A $60\text{ Hz}$:**
  $$T_{60\text{Hz}} = \frac{1000}{60} \approx 16.67\text{ ms}$$
* **A $120\text{ Hz}$:**
  $$T_{120\text{Hz}} = \frac{1000}{120} \approx 8.33\text{ ms}$$

Para evitar *frame drops* (caídas de cuadros) en telemetría 5G de alta frecuencia, todo el trabajo en el hilo principal (cálculo de métricas, reconciliación/actualización reactiva, *layout* y *paint*) debe caber holgadamente dentro de $16.67\text{ ms}$ (o $8.33\text{ ms}$ en pantallas de alta tasa).

## 2. Tabla Comparativa de Arquitecturas Reactivas

| Criterio de Auditoría | Next.js 15 / React 19 (`useState` + `useMemo` + VDOM Fiber) | Svelte 5 (`$state` + `$derived.by` + Compilación AOT) |
| :--- | :--- | :--- |
| **Complejidad de actualización** | Re-renderizado de componente y reconciliación/diffing de árbol Virtual DOM / Fiber ($O(N)$ sobre nodos del árbol afectado). | Actualización reactiva granular por grafo de dependencias/señales compiladas AOT sin Virtual DOM ($O(1)$ en el enlace directo de datos a nodos DOM). |
| **Tamaño del bundle JS** | Mayor: incluye runtime de React/Fiber y reconciliador en el cliente. | Menor: el compilador AOT emite código imperativo/reactivo directo sin runtime de VDOM. |
| **Hidratación** | Requiere hidratar el árbol de componentes en cliente tras el SSR/HTML inicial. | Menor carga de hidratación/arranque por compilación previa y reactividad fina. |
| **INP (Interaction to Next Paint)** | Sensible a bloqueos del hilo principal si el árbol reconciliado crece o se acumulan ráfagas en el mismo turno. | Menor latencia de respuesta en hilo principal al evitar recorrido de VDOM en actualizaciones puntuales. |
