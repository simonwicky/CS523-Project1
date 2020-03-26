package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	prog := os.Args[0]
	args := os.Args[1:]

	if len(args) < 2 {
		fmt.Println("Usage:", prog, "[Party ID] [Input]")
		os.Exit(1)
	}

	partyID, errPartyID := strconv.ParseUint(args[0], 10, 64)
	if errPartyID != nil {
		fmt.Println("Party ID should be an unsigned integer")
		os.Exit(1)
	}

	partyInput, errPartyInput := strconv.ParseUint(args[1], 10, 64)
	if errPartyInput != nil {
		fmt.Println("Party input should be an unsigned integer")
		os.Exit(1)
	}

	circuitID, errCircuitID := strconv.ParseUint(args[2], 10, 64)
	if errCircuitID != nil {
		fmt.Println("Circuit ID should be an unsigned integer")
		os.Exit(1)
	}

	Client(PartyID(partyID), partyInput, circuitID)
}

func Client(partyID PartyID, partyInput, circuitID uint64) {

	circuit := TestCircuits[circuitID]

	//nb of triplets to generate
	nb_mult := 0
	for _, op := range circuit.Circuit {
		switch op.(type) {
		case *Mult:
			nb_mult += 1
		}
	}

	// Create a local party
	lp, err := NewLocalParty(partyID, circuit.Peers)
	check(err)

	// Create the network for the circuit
	network, err := NewTCPNetwork(lp)
	check(err)

	// Connect the circuit network
	err = network.Connect(lp)
	check(err)
	fmt.Println(lp, "connected")
	<-time.After(time.Second) // Leave time for others to connect

	// Create a new circuit evaluation protocol
	dummyProtocol := lp.NewDummyProtocol(partyInput)
	dummyProtocol.Bp = lp.NewBeaverProtocol(nb_mult)
	dummyProtocol.Circuit = circuit.Circuit

	// Bind evaluation protocol to the network
	dummyProtocol.BindNetwork(network)
	dummyProtocol.Bp.BindNetwork(network)

	// Evaluate the circuit
	dummyProtocol.Run(false)

	fmt.Println(lp, "completed with output", dummyProtocol.Output)
}
