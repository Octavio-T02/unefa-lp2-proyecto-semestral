from __future__ import annotations

import asyncio
import re
from dataclasses import dataclass
from typing import Final

E164_PATTERN: Final[re.Pattern[str]] = re.compile(r"^\+[1-9]\d{6,14}$")


@dataclass(slots=True)
class SolicitudValidacion:
    tx_id: str
    msisdn: str
    monto_centavos: int


@dataclass(slots=True)
class ResultadoValidacion:
    tx_id: str
    valida: bool
    detalle: str


async def validar_solicitud_async(solicitud: SolicitudValidacion) -> ResultadoValidacion:
    # Simula verificacion I/O no bloqueante (p. ej. consulta HLR/consulta externa)
    await asyncio.sleep(0.01)

    if not E164_PATTERN.match(solicitud.msisdn):
        return ResultadoValidacion(
            tx_id=solicitud.tx_id,
            valida=False,
            detalle="MSISDN fuera de formato ITU-T E.164",
        )
    if solicitud.monto_centavos <= 0:
        return ResultadoValidacion(
            tx_id=solicitud.tx_id,
            valida=False,
            detalle="Monto en centavos debe ser positivo",
        )
    return ResultadoValidacion(
        tx_id=solicitud.tx_id,
        valida=True,
        detalle="OK",
    )


async def validar_lote(lote: list[SolicitudValidacion]) -> list[ResultadoValidacion]:
    tareas = [validar_solicitud_async(item) for item in lote]
    resultados = await asyncio.gather(*tareas)
    return list(resultados)


def main() -> None:
    lote: list[SolicitudValidacion] = [
        SolicitudValidacion(tx_id="PY-001", msisdn="+584121234567", monto_centavos=500),
        SolicitudValidacion(tx_id="PY-002", msisdn="04121234567", monto_centavos=500),
        SolicitudValidacion(tx_id="PY-003", msisdn="+584149876543", monto_centavos=0),
        SolicitudValidacion(tx_id="PY-004", msisdn="+584245550199", monto_centavos=2500),
    ]
    resultados = asyncio.run(validar_lote(lote))
    for res in resultados:
        print(f"{res.tx_id}: valida={res.valida} -> {res.detalle}")


if __name__ == "__main__":
    main()
