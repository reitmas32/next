package mathutils

import (
	"math"
	"testing"
)

func TestSum(t *testing.T) {
	tests := []struct {
		name     string
		numbers  []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"multiple", []float64{1, 2, 3, 4, 5}, 15},
		{"negative", []float64{-1, -2, -3}, -6},
		{"mixed", []float64{-5, 10, -5}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Sum(tt.numbers...)
			if result != tt.expected {
				t.Errorf("Sum(%v) = %v, want %v", tt.numbers, result, tt.expected)
			}
		})
	}
}

func TestAverage(t *testing.T) {
	tests := []struct {
		name     string
		numbers  []float64
		expected float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{10}, 10},
		{"multiple", []float64{2, 4, 6}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Average(tt.numbers...)
			if result != tt.expected {
				t.Errorf("Average(%v) = %v, want %v", tt.numbers, result, tt.expected)
			}
		})
	}
}

func TestDivide(t *testing.T) {
	// Test normal division
	result, err := Divide(10, 2)
	if err != nil {
		t.Errorf("Divide(10, 2) returned unexpected error: %v", err)
	}
	if result != 5 {
		t.Errorf("Divide(10, 2) = %v, want 5", result)
	}

	// Test division by zero
	_, err = Divide(10, 0)
	if err != ErrDivisionByZero {
		t.Errorf("Divide(10, 0) should return ErrDivisionByZero")
	}
}

func TestFactorial(t *testing.T) {
	tests := []struct {
		n        int
		expected int
		wantErr  bool
	}{
		{0, 1, false},
		{1, 1, false},
		{5, 120, false},
		{10, 3628800, false},
		{-1, 0, true},
	}

	for _, tt := range tests {
		result, err := Factorial(tt.n)
		if tt.wantErr && err == nil {
			t.Errorf("Factorial(%d) should return error", tt.n)
		}
		if !tt.wantErr && result != tt.expected {
			t.Errorf("Factorial(%d) = %d, want %d", tt.n, result, tt.expected)
		}
	}
}

func TestIsPrime(t *testing.T) {
	primes := []int{2, 3, 5, 7, 11, 13, 17, 19, 23, 29}
	nonPrimes := []int{0, 1, 4, 6, 8, 9, 10, 12, 14, 15}

	for _, p := range primes {
		if !IsPrime(p) {
			t.Errorf("IsPrime(%d) = false, want true", p)
		}
	}

	for _, np := range nonPrimes {
		if IsPrime(np) {
			t.Errorf("IsPrime(%d) = true, want false", np)
		}
	}
}

func TestFibonacci(t *testing.T) {
	expected := []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34}

	for i, exp := range expected {
		result := Fibonacci(i)
		if result != exp {
			t.Errorf("Fibonacci(%d) = %d, want %d", i, result, exp)
		}
	}
}

func TestSqrt(t *testing.T) {
	result, err := Sqrt(16)
	if err != nil {
		t.Errorf("Sqrt(16) returned unexpected error: %v", err)
	}
	if result != 4 {
		t.Errorf("Sqrt(16) = %v, want 4", result)
	}

	_, err = Sqrt(-1)
	if err != ErrNegativeNumber {
		t.Errorf("Sqrt(-1) should return ErrNegativeNumber")
	}
}

func TestMinMax(t *testing.T) {
	numbers := []float64{3, 1, 4, 1, 5, 9, 2, 6}

	if Max(numbers...) != 9 {
		t.Errorf("Max should return 9")
	}

	if Min(numbers...) != 1 {
		t.Errorf("Min should return 1")
	}

	if Max() != 0 {
		t.Errorf("Max() should return 0 for empty input")
	}

	if Min() != 0 {
		t.Errorf("Min() should return 0 for empty input")
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected float64
	}{
		{5, 0, 10, 5},
		{-5, 0, 10, 0},
		{15, 0, 10, 10},
		{0, 0, 10, 0},
		{10, 0, 10, 10},
	}

	for _, tt := range tests {
		result := Clamp(tt.value, tt.min, tt.max)
		if math.Abs(result-tt.expected) > 0.0001 {
			t.Errorf("Clamp(%v, %v, %v) = %v, want %v",
				tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}

