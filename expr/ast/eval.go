// eval.go
package ast

import (
	"fmt"
	"math"
)

// tiny wrapper so sym.go can call pow()
func float64FromMathPow(a, b float64) float64 {
	return math.Pow(a, b)
}

// Provide built-in functions and helper to evaluate parse result
func defaultEnv() Env {
	env := Env{
		Vars: make(map[string]float64),
		Funcs: map[string]func([]float64) (float64, error){
			"sqrt": func(args []float64) (float64, error) {
				if len(args) != 1 {
					return 0, fmt.Errorf("sqrt expects 1 arg")
				}
				return math.Sqrt(args[0]), nil
			},
			"max": func(args []float64) (float64, error) {
				if len(args) == 0 {
					return 0, fmt.Errorf("max expects >=1 arg")
				}
				m := args[0]
				for _, v := range args[1:] {
					if v > m {
						m = v
					}
				}
				return m, nil
			},
			"min": func(args []float64) (float64, error) {
				if len(args) == 0 {
					return 0, fmt.Errorf("min expects >=1 arg")
				}
				m := args[0]
				for _, v := range args[1:] {
					if v < m {
						m = v
					}
				}
				return m, nil
			},
		},
	}
	// example variables
	env.Vars["pi"] = math.Pi
	env.Vars["e"] = math.E
	return env
}
