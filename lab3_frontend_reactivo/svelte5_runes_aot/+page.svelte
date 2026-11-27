<script lang="ts">
  export interface CellTelemetryDTO {
    cellId: string;
    latencyMs: number;
    packetLossPct: number;
    throughputMbps: number;
  }

  const WINDOW_SIZE = 50;

  let packets = $state<CellTelemetryDTO[]>([
    { cellId: "CELL-5G-01", latencyMs: 18.4, packetLossPct: 0.12, throughputMbps: 940 },
    { cellId: "CELL-5G-02", latencyMs: 64.2, packetLossPct: 1.45, throughputMbps: 410 },
    { cellId: "CELL-5G-03", latencyMs: 24.8, packetLossPct: 0.25, throughputMbps: 820 },
    { cellId: "CELL-5G-04", latencyMs: 53.0, packetLossPct: 0.90, throughputMbps: 530 }
  ]);

  let lastUpdateMs = $state<number>(0);

  let metrics = $derived.by(() => {
    if (packets.length === 0) {
      return { avgLatencyMs: 0, criticalCells: 0 };
    }
    let sumLatency = 0;
    let critical = 0;
    for (const pkt of packets) {
      sumLatency += pkt.latencyMs;
      if (pkt.latencyMs > 50) {
        critical += 1;
      }
    }
    return {
      avgLatencyMs: Number((sumLatency / packets.length).toFixed(2)),
      criticalCells: critical
    };
  });

  function injectBurst(): void {
    const start = performance.now();
    const currentLen = packets.length;
    const burst: CellTelemetryDTO[] = Array.from({ length: 20 }, (_, idx) => {
      const seq = currentLen + idx + 1;
      const latencyMs = Number((12 + ((seq * 13) % 75) + (idx % 3) * 4.5).toFixed(1));
      return {
        cellId: `CELL-5G-${String((seq % 16) + 1).padStart(2, "0")}`,
        latencyMs,
        packetLossPct: Number(((seq % 10) * 0.22).toFixed(2)),
        throughputMbps: 250 + ((seq * 47) % 900)
      };
    });
    packets = [...packets, ...burst].slice(-WINDOW_SIZE);
    lastUpdateMs = Number((performance.now() - start).toFixed(3));
  }
</script>

<section aria-label="Telemetría 5G Svelte 5 Runes">
  <header>
    <h2>Panel de Telemetría de Celdas 5G (Svelte 5 Runes AOT)</h2>
    <button type="button" onclick={injectBurst}>
      Inyectar Ráfaga de Telemetría
    </button>
  </header>

  <div>
    <p>Latencia promedio: <strong>{metrics.avgLatencyMs} ms</strong></p>
    <p>Celdas críticas (&gt;50 ms): <strong>{metrics.criticalCells}</strong></p>
    <p>Tiempo de actualización (performance.now): <strong>{lastUpdateMs} ms</strong></p>
    <p>Paquetes en ventana: <strong>{packets.length}</strong></p>
  </div>

  <table>
    <thead>
      <tr>
        <th>Celda</th>
        <th>Latencia (ms)</th>
        <th>Pérdida (%)</th>
        <th>Throughput (Mbps)</th>
      </tr>
    </thead>
    <tbody>
      {#each packets as pkt, i (`${pkt.cellId}-${i}`)}
        <tr>
          <td>{pkt.cellId}</td>
          <td>{pkt.latencyMs}</td>
          <td>{pkt.packetLossPct}</td>
          <td>{pkt.throughputMbps}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</section>
