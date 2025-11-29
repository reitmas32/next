// Package mathutils proporciona utilidades matemáticas simples.
// Esta es una biblioteca de ejemplo para demostrar el uso del CLI next.
package mathutils

import (
	"errors"
	"math"
)

// ErrDivisionByZero se retorna cuando se intenta dividir por cero.
var ErrDivisionByZero = errors.New("división por cero")

// ErrNegativeNumber se retorna cuando se requiere un número positivo.
var ErrNegativeNumber = errors.New("número negativo no permitido")

// Sum retorna la suma de todos los números.
func Sum(numbers ...float64) float64 {
	var total float64
	for _, n := range numbers {
		total += n
	}
	return total
}

// Average calcula el promedio de los números.
func Average(numbers ...float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	return Sum(numbers...) / float64(len(numbers))
}

// Divide divide a entre b, retorna error si b es cero.
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, ErrDivisionByZero
	}
	return a / b, nil
}

// Factorial calcula el factorial de n.
func Factorial(n int) (int, error) {
	if n < 0 {
		return 0, ErrNegativeNumber
	}
	if n == 0 || n == 1 {
		return 1, nil
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result, nil
}

// IsPrime verifica si un número es primo.
func IsPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// Fibonacci retorna el n-ésimo número de Fibonacci.
func Fibonacci(n int) int {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// Sqrt calcula la raíz cuadrada de un número.
func Sqrt(n float64) (float64, error) {
	if n < 0 {
		return 0, ErrNegativeNumber
	}
	return math.Sqrt(n), nil
}

// Max retorna el número máximo de la lista.
func Max(numbers ...float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	max := numbers[0]
	for _, n := range numbers[1:] {
		if n > max {
			max = n
		}
	}
	return max
}

// Min retorna el número mínimo de la lista.
func Min(numbers ...float64) float64 {
	if len(numbers) == 0 {
		return 0
	}
	min := numbers[0]
	for _, n := range numbers[1:] {
		if n < min {
			min = n
		}
	}
	return min
}

// Clamp restringe un valor entre un mínimo y máximo.
func Clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Subtract subtracts b from a.
func Subtract(a, b float64) float64 {
	return a - b
}

func Logarithm(base, x float64) float64 {
	return math.Log(x) / math.Log(base)
}
