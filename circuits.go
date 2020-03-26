package main

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

func SetUpMPC(circuit *TestCircuit, trusted bool) (dummyProtocol []*DummyProtocol, wg *sync.WaitGroup) {

	N := uint64(len(circuit.Peers))
	P := make([]*LocalParty, N, N)
	dummyProtocol = make([]*DummyProtocol, N, N)

	//nb of triplet to generate
	nb_mult := 0
	for _, op := range circuit.Circuit {
		switch op.(type) {
		case *Mult:
			nb_mult += 1
		}
	}

	var err error
	wg = new(sync.WaitGroup)
	for i := range circuit.Peers {
		P[i], err = NewLocalParty(i, circuit.Peers)
		P[i].WaitGroup = wg
		check(err)

		dummyProtocol[i] = P[i].NewDummyProtocol(circuit.Inputs[i][GateID(i)])
		dummyProtocol[i].Bp = P[i].NewBeaverProtocol(nb_mult)
		dummyProtocol[i].Circuit = circuit.Circuit
	}

	//trusted third party setting
	if trusted {
		generateBeaverTriplet(nb_mult, dummyProtocol)
	}

	network := GetTestingTCPNetwork(P)
	fmt.Println("parties connected")

	for i, Pi := range dummyProtocol {
		Pi.Bp.BindNetwork(network[i])
		Pi.BindNetwork(network[i])
	}
	return
}

func (cep *DummyProtocol) ComputeCircuit() (out uint64, err error) {
	secret := cep.Secret
	circuit := cep.Circuit

	n_secret := len(secret)
	n_circuit := len(circuit)

	triplets := cep.Triplets
	t := 0

	if n_circuit <= n_secret {
		return 0, errors.New("number of secrets does not match number of circuit inputs")
	}

	result := make([]uint64, n_circuit)

	for i, op := range circuit[:n_secret] {
		if reflect.TypeOf(op) != reflect.TypeOf(&Input{}) {
			return 0, errors.New("number of secrets does not match number of circuit inputs")
		}
		result[i] = secret[i]
	}

	revealed := false
	for i, op := range circuit[n_secret:] {
		i += n_secret
		if op.Output() != WireID(i) {
			return 0, errors.New("out WireIDs must be sorted in increasing order")
		}
		if revealed {
			return 0, errors.New("Reveal gate must be the last gate of the circuit.")
		}
		switch op.(type) {
		case *Input:
			return 0, errors.New("number of secrets does not match number of circuit inputs")
		case *Add:
			add := op.(*Add)
			in1, in2 := result[uint64(add.In1)], result[uint64(add.In2)]
			result[i] = Add_Gate(in1, in2)
		case *AddCst:
			addCst := op.(*AddCst)
			in, cst := result[uint64(addCst.In)], addCst.CstValue
			result[i] = AddCst_Gate(cep.ID, in, cst)
		case *Sub:
			sub := op.(*Sub)
			in1, in2 := result[uint64(sub.In1)], result[uint64(sub.In2)]
			result[i] = Sub_Gate(in1, in2)
		case *Mult:
			mult := op.(*Mult)
			in1, in2 := result[uint64(mult.In1)], result[uint64(mult.In2)]
			if t >= len(triplets) {
				return 0, errors.New("not enough triplets were provided")
			}
			result[i] = Mult_Gate(in1, in2, uint64(i), triplets[t], cep)
			t += 1
		case *MultCst:
			multCst := op.(*MultCst)
			in, cst := result[uint64(multCst.In)], multCst.CstValue
			result[i] = MultCst_Gate(in, cst)
		case *Reveal:
			result[i] = result[i-1]
			out = Reveal_Gate(cep, result[i], uint64(i))
			revealed = true
		default:
			return 0, errors.New("gate type is not recognized")
		}
	}
	return
}
