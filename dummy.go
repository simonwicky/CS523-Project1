package main

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"
)

const (
	MODULUS = 9973
)

type DummyMessage struct {
	Party PartyID
	Value uint64
}

type DummyProtocol struct {
	*LocalParty
	Chan  chan DummyMessage
	Peers map[PartyID]*DummyRemote

	Circuit []Operation

	Input       uint64
	Input_share uint64
	Secret      []uint64
	Output      uint64
}

type DummyRemote struct {
	*RemoteParty
	Chan chan DummyMessage
}

func (lp *LocalParty) NewDummyProtocol(input uint64) *DummyProtocol {
	cep := new(DummyProtocol)
	cep.LocalParty = lp
	cep.Chan = make(chan DummyMessage, 32)
	cep.Peers = make(map[PartyID]*DummyRemote, len(lp.Peers))
	for i, rp := range lp.Peers {
		cep.Peers[i] = &DummyRemote{
			RemoteParty: rp,
			Chan:        make(chan DummyMessage, 32),
		}
	}
	cep.Secret = make([]uint64, len(lp.Peers))
	cep.Input = input
	cep.Output = input
	return cep
}

func (cep *DummyProtocol) BindNetwork(nw *TCPNetworkStruct) {
	for partyID, conn := range nw.Conns {

		if partyID == cep.ID {
			continue
		}

		rp := cep.Peers[partyID]

		// Receiving loop from remote
		go func(conn net.Conn, rp *DummyRemote) {
			for {
				var id, val uint64
				var err error
				err = binary.Read(conn, binary.BigEndian, &id)
				check(err)
				err = binary.Read(conn, binary.BigEndian, &val)
				check(err)
				msg := DummyMessage{
					Party: PartyID(id),
					Value: val,
				}
				//fmt.Println(cep, "receiving", msg, "from", rp)
				cep.Chan <- msg
			}
		}(conn, rp)

		// Sending loop of remote
		go func(conn net.Conn, rp *DummyRemote) {
			var m DummyMessage
			var open = true
			for open {
				m, open = <-rp.Chan
				//fmt.Println(cep, "sending", m, "to", rp)
				check(binary.Write(conn, binary.BigEndian, m.Party))
				check(binary.Write(conn, binary.BigEndian, m.Value))
			}
		}(conn, rp)
	}
}

func (cep *DummyProtocol) Run() {
	fmt.Println(cep, "is running")
	rand.Seed(time.Now().UTC().UnixNano())

	//create secret shares and send them
	cep.Input_share = cep.Input
	for _, peer := range cep.Peers {
		if peer.ID != cep.ID {
			share := rand.Intn(MODULUS)
			cep.Input_share = uint64(Pmod(int(cep.Input_share)-share, MODULUS))
			peer.Chan <- DummyMessage{cep.ID, uint64(share)}
		}
	}
	cep.Secret[cep.ID] = cep.Input_share

	//collect shares from other peers
	received := 0
	for m := range cep.Chan {
		fmt.Println(cep, "received message from", m.Party, ":", m.Value)
		cep.Secret[m.Party] = m.Value
		received++
		if received == len(cep.Peers)-1 {
			close(cep.Chan)
		}
	}
	//shares are ready, let's compute the circuit

	cep.Output = cep.ComputeCircuit()

	if cep.WaitGroup != nil {
		cep.WaitGroup.Done()

	}
}
