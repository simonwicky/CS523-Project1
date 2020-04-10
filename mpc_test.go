package main

import (
	"fmt"
	"testing"
)

//Part 1 is run by setting the trusted variable to true
//Part 2 is run by setting the trusted variable to false
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
			//set trusted here
			trusted := true
			circuit := TestCircuits[tc.index]
			mpcP, wg := SetUpMPC(circuit, trusted)

			//waitGroup and Run
			for _, cep := range mpcP {
				cep.Add(1)
				go cep.Run(trusted)
			}
			wg.Wait()

			for _, cep := range mpcP {
				if cep.Output != circuit.ExpOutput {
					t.Errorf("peer %v output %v did not match with expected value %v\n", cep.ID, cep.Output, circuit.ExpOutput)
				}
			}
			fmt.Printf("%v tested successfull with output %v\n", tc.name, circuit.ExpOutput)
		})
	}
}

//Part 1 is run by setting the trusted variable to true
//Part 2 is run by setting the trusted variable to false
func BenchmarkEval(b *testing.B) {
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
		b.Run(tc.name, func(b *testing.B) {
			//set trusted here
			trusted := true
			circuit := TestCircuits[tc.index]
			mpcP, wg := SetUpMPC(circuit, trusted)
			for i := 0; i < b.N; i++ {

				//waitGroup and Run
				for _, cep := range mpcP {
					cep.Add(1)
					go cep.Run(trusted)
				}
				wg.Wait()
			}
		})
	}
}
