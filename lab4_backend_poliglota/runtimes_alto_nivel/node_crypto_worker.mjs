import crypto from "node:crypto";
import { Worker, isMainThread, parentPort, workerData } from "node:worker_threads";

function calcularHashPesado(payload, iteraciones) {
  const digest = crypto.pbkdf2Sync(payload, "unefa-5g-salt", iteraciones, 32, "sha256");
  return crypto.createHash("sha256").update(digest).digest("hex");
}

if (!isMainThread) {
  const { payload, iteraciones } = workerData;
  const hash = calcularHashPesado(payload, iteraciones);
  parentPort.postMessage(hash);
} else {
  const runWorkerJob = (payload, iteraciones) =>
    new Promise((resolve, reject) => {
      const worker = new Worker(new URL(import.meta.url), {
        workerData: { payload, iteraciones },
      });
      worker.on("message", resolve);
      worker.on("error", reject);
      worker.on("exit", (code) => {
        if (code !== 0) {
          reject(new Error(`Worker finalizo con codigo ${code}`));
        }
      });
    });

  const inicio = performance.now();
  let ticksEventLoop = 0;

  // Monitor liviano en el Event Loop principal para verificar que no quede bloqueado
  const timer = setInterval(() => {
    ticksEventLoop += 1;
  }, 5);

  const hashes = await Promise.all([
    runWorkerJob("CDR-LOTE-1:+584121111111", 120_000),
    runWorkerJob("CDR-LOTE-2:+584142222222", 120_000),
    runWorkerJob("CDR-LOTE-3:+584163333333", 120_000),
    runWorkerJob("CDR-LOTE-4:+584244444444", 120_000),
  ]);

  clearInterval(timer);
  const elapsedMs = (performance.now() - inicio).toFixed(2);

  console.log(`Completados ${hashes.length} hashes en ${elapsedMs} ms con ${ticksEventLoop} ticks en Event Loop libuv.`);
  console.log(`Muestra hash[0]: ${hashes[0].slice(0, 24)}...`);
}
