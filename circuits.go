package main

import (
	"fmt"
	"sync"
)

func SetUpMPC(circuit *TestCircuit) (dummyProtocol []*DummyProtocol, wg *sync.WaitGroup) {

	N := uint64(len(circuit.Peers))
	P := make([]*LocalParty, N, N)
	dummyProtocol = make([]*DummyProtocol, N, N)

	var err error
	wg = new(sync.WaitGroup)
	for i := range circuit.Peers {
		P[i], err = NewLocalParty(i, circuit.Peers)
		P[i].WaitGroup = wg
		check(err)

		dummyProtocol[i] = P[i].NewDummyProtocol(circuit.Inputs[i][GateID(i)])
		dummyProtocol[i].Circuit = circuit.Circuit
	}

	network := GetTestingTCPNetwork(P)
	fmt.Println("parties connected")

	for i, Pi := range dummyProtocol {
		Pi.BindNetwork(network[i])
	}
	return
}

func (cep *DummyProtocol) CheckResult(circuit *TestCircuit) bool {
	return cep.Output == circuit.ExpOutput
}

func ComputeCircuit(circuit *TestCircuit, inputs []uint64) uint64 {
	return 0
}
