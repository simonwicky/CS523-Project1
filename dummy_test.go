package main

import (
	"fmt"
	"testing"
)

func TestEval(t *testing.T) {
	testCases := []struct {
		name  string
		index int
	}{
		{"circuit1", 0},
		{"circuit2", 1},
		{"circuit3", 2},
		{"circuit4", 3},
		{"circuit5", 4},
		{"circuit6", 5},
		{"circuit7", 6},
		{"circuit8", 7},
		{"circuit9", 8},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			circuit := TestCircuits[tc.index]
			dp, wg := SetUpMPC(circuit)

			//waitGroup and Run
			for _, cep := range dp {
				cep.Add(1)
				go cep.Run()
			}
			wg.Wait()

			for _, cep := range dp {
				if cep.Output != circuit.ExpOutput {
					t.Errorf("peer %v output %v did not match with expected value %v\n", cep.ID, cep.Output, circuit.ExpOutput)
				}
			}
			fmt.Printf("%v tested successfull with output %v\n", tc.name, circuit.ExpOutput)
		})
	}
}
