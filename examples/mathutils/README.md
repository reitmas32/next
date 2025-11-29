# mathutils

Biblioteca de utilidades matemáticas simples para Go.

## Instalación

```bash
go get github.com/reitmas32/mathutils@v1.0.0
```

## Uso

```go
package main

import (
    "fmt"
    "github.com/reitmas32/mathutils"
)

func main() {
    // Suma y promedio
    sum := mathutils.Sum(1, 2, 3, 4, 5)
    avg := mathutils.Average(1, 2, 3, 4, 5)
    fmt.Printf("Suma: %.0f, Promedio: %.1f\n", sum, avg)

    // División segura
    result, err := mathutils.Divide(10, 2)
    if err != nil {
        panic(err)
    }
    fmt.Printf("10 / 2 = %.0f\n", result)

    // Factorial
    fact, _ := mathutils.Factorial(5)
    fmt.Printf("5! = %d\n", fact)

    // Números primos
    fmt.Printf("¿7 es primo? %v\n", mathutils.IsPrime(7))

    // Fibonacci
    fmt.Printf("Fibonacci(10) = %d\n", mathutils.Fibonacci(10))

    // Min/Max
    numbers := []float64{3, 1, 4, 1, 5, 9}
    fmt.Printf("Max: %.0f, Min: %.0f\n", 
        mathutils.Max(numbers...), 
        mathutils.Min(numbers...))

    // Clamp
    clamped := mathutils.Clamp(15, 0, 10)
    fmt.Printf("Clamp(15, 0, 10) = %.0f\n", clamped)
}
```

## Funciones disponibles

| Función | Descripción |
|---------|-------------|
| `Sum(numbers...)` | Suma todos los números |
| `Average(numbers...)` | Calcula el promedio |
| `Divide(a, b)` | División segura (retorna error si b=0) |
| `Factorial(n)` | Calcula n! |
| `IsPrime(n)` | Verifica si es primo |
| `Fibonacci(n)` | N-ésimo número de Fibonacci |
| `Sqrt(n)` | Raíz cuadrada (con validación) |
| `Max(numbers...)` | Valor máximo |
| `Min(numbers...)` | Valor mínimo |
| `Clamp(value, min, max)` | Restringe valor a rango |

## Versionado

Esta biblioteca usa versionado semántico. Para crear una nueva versión:

```bash
cd examples/mathutils
next create-version v1.0.0
```

## Licencia

MIT

