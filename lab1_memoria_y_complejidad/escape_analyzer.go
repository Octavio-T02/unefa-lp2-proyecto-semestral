package main

import (
	"fmt"
	"time"
)

type Sesion5G struct {
	MSISDN      string
	CeldaID     uint32
	LatenciaMs  float64
	Activa      bool
}

// CrearEnStackPorValor retorna la estructura por valor; al no escaparse
// el puntero, el compilador puede mantenerla en el stack del llamador.
func CrearEnStackPorValor(msisdn string, celda uint32, latencia float64) Sesion5G {
	s := Sesion5G{
		MSISDN:     msisdn,
		CeldaID:    celda,
		LatenciaMs: latencia,
		Activa:     true,
	}
	return s
}

// CrearConEscapeAlHeap retorna &s de una variable local; el puntero sobrevive
// al retorno de la funcion, por lo que s escapa al Heap.
func CrearConEscapeAlHeap(msisdn string, celda uint32, latencia float64) *Sesion5G {
	s := Sesion5G{
		MSISDN:     msisdn,
		CeldaID:    celda,
		LatenciaMs: latencia,
		Activa:     true,
	}
	return &s
}

func busquedaLineal(slice []Sesion5G, objetivo string) (Sesion5G, bool) {
	for i := range slice {
		if slice[i].MSISDN == objetivo {
			return slice[i], true
		}
	}
	return Sesion5G{}, false
}

func main() {
	// 1. Demostracion de asignacion en Stack vs Escape al Heap
	sesionStack := CrearEnStackPorValor("+584121000001", 101, 12.4)
	sesionHeap := CrearConEscapeAlHeap("+584141000002", 202, 18.9)
	fmt.Printf("Stack valor: %+v | Heap puntero: %+v\n", sesionStack, *sesionHeap)

	// 2. Preparacion de lote de N = 100,000 abonados
	const totalAbonados = 100_000
	const consultas = 1_000

	abonadosSlice := make([]Sesion5G, totalAbonados)
	abonadosMapa := make(map[string]Sesion5G, totalAbonados)

	for i := 0; i < totalAbonados; i++ {
		msisdn := fmt.Sprintf("+58412%07d", i)
		reg := Sesion5G{
			MSISDN:     msisdn,
			CeldaID:    uint32(i%512) + 1,
			LatenciaMs: float64((i%40)+5) * 0.5,
			Activa:     true,
		}
		abonadosSlice[i] = reg
		abonadosMapa[msisdn] = reg
	}

	// Lista de 1,000 objetivos (incluye elementos distribuidos y algunos ausentes)
	objetivos := make([]string, consultas)
	for q := 0; q < consultas; q++ {
		idx := (q * 97) % totalAbonados
		objetivos[q] = fmt.Sprintf("+58412%07d", idx)
	}

	// Evita optimizacion de codigo muerto acumulando resultados
	var encontradosLineal int
	inicioLineal := time.Now()
	for _, target := range objetivos {
		if _, ok := busquedaLineal(abonadosSlice, target); ok {
			encontradosLineal++
		}
	}
	duracionLinealUs := float64(time.Since(inicioLineal).Microseconds())

	var encontradosMapa int
	inicioMapa := time.Now()
	for _, target := range objetivos {
		if _, ok := abonadosMapa[target]; ok {
			encontradosMapa++
		}
	}
	duracionMapaUs := float64(time.Since(inicioMapa).Microseconds())

	fmt.Printf("\nBenchmark (%d busquedas sobre N=%d):\n", consultas, totalAbonados)
	fmt.Printf("  Busqueda Lineal O(n) : %.2f us (encontrados=%d)\n", duracionLinealUs, encontradosLineal)
	fmt.Printf("  Busqueda Mapa   O(1) : %.2f us (encontrados=%d)\n", duracionMapaUs, encontradosMapa)
}