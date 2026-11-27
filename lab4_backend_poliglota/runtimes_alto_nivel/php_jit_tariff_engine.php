<?php

declare(strict_types=1);

enum TipoRed: string
{
    case LTE_4G = '4G';
    case NR_5G = '5G';
    case ROAMING = 'ROAMING';
}

readonly class CalculadorTarifa5G
{
    public function __construct(
        public TipoRed $tipoRed,
        public int $megabytesConsumidos,
    ) {
        if ($this->megabytesConsumidos < 0) {
            throw new InvalidArgumentException('Los megabytes consumidos no pueden ser negativos.');
        }
    }

    public function calcularCentavos(): int
    {
        $tarifaPorMbCentavos = match ($this->tipoRed) {
            TipoRed::LTE_4G => 2,
            TipoRed::NR_5G => 3,
            TipoRed::ROAMING => 15,
        };

        return $this->megabytesConsumidos * $tarifaPorMbCentavos;
    }
}

$calculos = [
    new CalculadorTarifa5G(TipoRed::LTE_4G, 500),
    new CalculadorTarifa5G(TipoRed::NR_5G, 1000),
    new CalculadorTarifa5G(TipoRed::ROAMING, 120),
];

foreach ($calculos as $calc) {
    echo sprintf(
        "Red=%s | MB=%d | Total=%d centavos\n",
        $calc->tipoRed->value,
        $calc->megabytesConsumidos,
        $calc->calcularCentavos()
    );
}
