"use client";

import React, { useMemo, useState } from "react";

export interface CellTelemetryDTO {
  cellId: string;
  latencyMs: number;
  packetLossPct: number;
  throughputMbps: number;
}

const INITIAL_PACKETS: CellTelemetryDTO[] = [
  { cellId: "CELL-5G-01", latencyMs: 18.4, packetLossPct: 0.12, throughputMbps: 940 },
  { cellId: "CELL-5G-02", latencyMs: 64.2, packetLossPct: 1.45, throughputMbps: 410 },
  { cellId: "CELL-5G-03", latencyMs: 24.8, packetLossPct: 0.25, throughputMbps: 820 },
  { cellId: "CELL-5G-04", latencyMs: 53.0, packetLossPct: 0.90, throughputMbps: 530 },
];

const WINDOW_SIZE = 50;

export default function TelemetryPanelClient() {
  const [packets, setPackets] = useState<CellTelemetryDTO[]>(INITIAL_PACKETS);
  const [lastUpdateMs, setLastUpdateMs] = useState<number>(0);

  const metrics = useMemo(() => {
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
      criticalCells: critical,
    };
  }, [packets]);

  const handleInjectBurst = () => {
    const start = performance.now();
    setPackets((prev) => {
      const burst: CellTelemetryDTO[] = Array.from({ length: 20 }, (_, idx) => {
        const seq = prev.length + idx + 1;
        const latencyMs = Number((12 + ((seq * 13) % 75) + (idx % 3) * 4.5).toFixed(1));
        return {
          cellId: `CELL-5G-${String((seq % 16) + 1).padStart(2, "0")}`,
          latencyMs,
          packetLossPct: Number(((seq % 10) * 0.22).toFixed(2)),
          throughputMbps: 250 + ((seq * 47) % 900),
        };
      });
      return [...prev, ...burst].slice(-WINDOW_SIZE);
    });
    const elapsed = Number((performance.now() - start).toFixed(3));
    setLastUpdateMs(elapsed);
  };

  return (
    <section aria-label="Telemetría 5G React 19">
      <header>
        <h2>Panel de Telemetría de Celdas 5G (Next.js 15 / React 19 VDOM)</h2>
        <button type="button" onClick={handleInjectBurst}>
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
          {packets.map((pkt, i) => (
            <tr key={`${pkt.cellId}-${i}`}>
              <td>{pkt.cellId}</td>
              <td>{pkt.latencyMs}</td>
              <td>{pkt.packetLossPct}</td>
              <td>{pkt.throughputMbps}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </section>
  );
}
