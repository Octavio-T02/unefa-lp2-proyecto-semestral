use std::sync::{Arc, Mutex};
use std::thread;

#[derive(Debug, Clone)]
struct TransaccionCruda {
    id_tx: String,
    msisdn: String,
    monto_centavos: u64,
    velocidad_tx_min: u32,
}

#[derive(Debug)]
struct AlertaFraude<'a> {
    id_tx: &'a str,
    msisdn: &'a str,
    motivo: &'a str,
    monto_centavos: u64,
}

#[derive(Debug, Default)]
struct EstadisticasRiesgo {
    total_evaluadas: u64,
    total_alertas: u64,
    monto_riesgo_centavos: u64,
}

fn evaluar_lote<'a>(
    lote: &'a [TransaccionCruda],
    stats: &Arc<Mutex<EstadisticasRiesgo>>,
) -> Vec<AlertaFraude<'a>> {
    let mut alertas_locales = Vec::new();
    let mut evaluadas_loc = 0_u64;
    let mut monto_riesgo_loc = 0_u64;

    for tx in lote {
        evaluadas_loc += 1;
        let motivo = if tx.monto_centavos >= 50_000 {
            Some("MONTO_ALTO_ATIPICO")
        } else if tx.velocidad_tx_min >= 15 {
            Some("RAFAGA_ALTA_FRECUENCIA")
        } else {
            None
        };

        if let Some(razon) = motivo {
            monto_riesgo_loc += tx.monto_centavos;
            alertas_locales.push(AlertaFraude {
                id_tx: tx.id_tx.as_str(),
                msisdn: tx.msisdn.as_str(),
                motivo: razon,
                monto_centavos: tx.monto_centavos,
            });
        }
    }

    let mut guard = stats.lock().expect("mutex envenenado");
    guard.total_evaluadas += evaluadas_loc;
    guard.total_alertas += alertas_locales.len() as u64;
    guard.monto_riesgo_centavos += monto_riesgo_loc;

    alertas_locales
}

fn main() {
    let transacciones: Vec<TransaccionCruda> = (1..=40)
        .map(|i| TransaccionCruda {
            id_tx: format!("TX-{:04}", i),
            msisdn: format!("+58412{:07}", 1_000_000 + i),
            monto_centavos: if i % 5 == 0 { 75_000 } else { 1_500 },
            velocidad_tx_min: if i % 7 == 0 { 22 } else { 3 },
        })
        .collect();

    let stats = Arc::new(Mutex::new(EstadisticasRiesgo::default()));
    let tam_lote = transacciones.len().div_ceil(4);

    let mut total_alertas_impresas = 0_usize;

    thread::scope(|scope| {
        let mut handles = Vec::with_capacity(4);
        for chunk in transacciones.chunks(tam_lote) {
            let stats_ref = Arc::clone(&stats);
            handles.push(scope.spawn(move || evaluar_lote(chunk, &stats_ref)));
        }

        for h in handles {
            let alertas_hilo = h.join().expect("hilo de auditoria entro en panico");
            for alerta in &alertas_hilo {
                println!(
                    "Alerta Zero-Copy -> id={} msisdn={} motivo={} monto={}",
                    alerta.id_tx, alerta.msisdn, alerta.motivo, alerta.monto_centavos
                );
            }
            total_alertas_impresas += alertas_hilo.len();
        }
    });

    let final_stats = stats.lock().expect("mutex envenenado");
    println!(
        "\nEstadisticas Globales -> evaluadas={} alertas={} monto_riesgo={} alertas_impresas={}",
        final_stats.total_evaluadas,
        final_stats.total_alertas,
        final_stats.monto_riesgo_centavos,
        total_alertas_impresas
    );
}
