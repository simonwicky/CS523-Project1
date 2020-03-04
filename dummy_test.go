package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestDummyProtocol(t *testing.T) {
	peers := map[PartyID]string{
		0: "localhost:6660",
		1: "localhost:6661",
		2: "localhost:6662",
	}

	N := uint64(len(peers))
	P := make([]*LocalParty, N, N)
	dummyProtocol := make([]*DummyProtocol, N, N)

	var err error
	wg := new(sync.WaitGroup)
	for i := range peers {
		P[i], err = NewLocalParty(i, peers)
		P[i].WaitGroup = wg
		check(err)

		dummyProtocol[i] = P[i].NewDummyProtocol(uint64(i + 10))
	}

	network := GetTestingTCPNetwork(P)
	fmt.Println("parties connected")

	for i, Pi := range dummyProtocol {
		Pi.BindNetwork(network[i])
	}

	for _, p := range dummyProtocol {
		p.Add(1)
		go p.Run()
	}
	wg.Wait()

	for _, p := range dummyProtocol {
		fmt.Println(p, "completed with output", p.Secret)
	}

	fmt.Println("test completed")
}

func TestDummyProtocol2(t *testing.T) {
	peers := map[PartyID]string{
		0: "localhost:6660",
		1: "localhost:6661",
		2: "localhost:6662",
	}

	N := uint64(len(peers))
	dummyProtocol := make([]*DummyProtocol, N, N)

	P := make([]*LocalParty, N, N)

	var err error
	for i := range peers {
		P[i], err = NewLocalParty(i, peers)
		check(err)

		dummyProtocol[i] = P[i].NewDummyProtocol(uint64(i + 10))
	}

	//dummyProtocol[0].Run()
}

func TestEval(t *testing.T) {
	t.Run("circuit1", func(t *testing.T) {

		dummyProtocol, wg := SetUpMPC(TestCircuits[0])

		//waitGroup and Run
		for _, p := range dummyProtocol {
			p.Add(1)
			go p.Run()
		}
		wg.Wait()

		for _, p := range dummyProtocol {
			fmt.Println(p, "completed with output", p.Secret)
		}

	})
	t.Run("circuit2", func(t *testing.T) {
		//test circuit2
	})
	t.Run("circuit3", func(t *testing.T) {
		//test circuit3
	})
	t.Run("circuit4", func(t *testing.T) {
		//test circuit4
	})
}
