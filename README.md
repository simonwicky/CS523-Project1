# CS523-Project1
This is the repository for the project 1 of the CS-523 course

## Overview
This project is splitted in two parts :  
Part 1 : Secure MPC using a trusted third party for Beaver triplets generation  
Part 2 : Secure MPC using homomorphic encryption for Beaver triplets generation

## Main
**Used only for Part 2**   
Syntax is as follows:  
`./mpc [PartyID] [Input] [CircuitID]`   
* PartyID : The ID of the Party  
* Input : The Input of the party  
* CircuitID : The ID of the circuit to compute. Circuits are stored in `test_circuits.go`  

Note that the number of party should be equal to the number of inputs of the circuits, and that all parties should be run simultaneously.

## Tests
Tests works for both parts. To change parts, set the boolean trusted in `mpc_test.go:28`to true for Part 1, false for Part 2.
The tests can be run as follows:
* To run all tests : `go test -v -run=.`
* To run a test on a specific circuit with id e.g 1 : `go test -v -run=^TestEval$/^circuit1$`

## Benchmark
Benchmark works for both parts. To change parts, set the boolean trusted in `mpc_test.go:69`to true for Part 1, false for Part 2.
The benchmarks can be run as follows:
* To run all benchmarks : `go test -v -bench=. -run=XXX`
* To run a benchmark on a specific circuit with id e.g 1 : `go test -v -bench=^BenchmarkEval$/^circuit1$ -run=XXX`
It is possible that the benchmarks fails because of too many open file. It that case, run it again with the tag `-benchtime=XXXms`to avoid this issue. 