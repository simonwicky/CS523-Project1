package main

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/ldsec/lattigo/bfv"
)

var MODULUS uint64 = bfv.DefaultParams[bfv.PN13QP218].T

type MPCMessage struct {
	Party PartyID
	Value uint64
	Id0   uint64
	Id1   uint64
}

type MPCProtocol struct {
	*LocalParty
	Chan  chan MPCMessage
	Peers map[PartyID]*MPCRemote

	Bp *BeaverProtocol

	Circuit []Operation

	Triplets    [][3]uint64
	Input       uint64
	Input_share uint64
	Secret      []uint64
	Output      uint64
}

type MPCRemote struct {
	*RemoteParty
	Chan chan MPCMessage
}

func (lp *LocalParty) NewMPCProtocol(input uint64) *MPCProtocol {
	cep := new(MPCProtocol)
	cep.LocalParty = lp
	cep.Chan = make(chan MPCMessage, 32)
	cep.Peers = make(map[PartyID]*MPCRemote, len(lp.Peers))
	for i, rp := range lp.Peers {
		cep.Peers[i] = &MPCRemote{
			RemoteParty: rp,
			Chan:        make(chan MPCMessage, 32),
		}
	}
	cep.Secret = make([]uint64, len(lp.Peers))
	cep.Input = input
	cep.Output = input
	return cep
}

func (cep *MPCProtocol) BindNetwork(nw *TCPNetworkStruct) {
	for partyID, conn := range nw.Conns {

		if partyID == cep.ID {
			continue
		}

		rp := cep.Peers[partyID]

		// Receiving loop from remote
		go func(conn net.Conn, rp *MPCRemote) {
			for {
				var msgID uint64
				var err error
				err = binary.Read(conn, binary.BigEndian, &msgID)
				check(err)
				if msgID == 0 {
					//beaverMessage
					var id, loopID uint64
					check(binary.Read(conn, binary.BigEndian, &id))
					var len uint64
					check(binary.Read(conn, binary.BigEndian, &len))
					c := make([]byte, len)
					check(binary.Read(conn, binary.BigEndian, &c))
					check(binary.Read(conn, binary.BigEndian, &loopID))
					msg := BeaverMessage{
						Party:  PartyID(id),
						D:      bfv.NewCiphertext(cep.Bp.Param, 1<<cep.Bp.Param.LogN),
						loopID: loopID,
					}
					msg.D.UnmarshalBinary(c)
					cep.Bp.Chan <- msg
				} else if msgID == 1 {
					//MPCMessage
					var id, val, idmsg0, idmsg1 uint64
					err = binary.Read(conn, binary.BigEndian, &id)
					check(err)
					err = binary.Read(conn, binary.BigEndian, &val)
					check(err)
					err = binary.Read(conn, binary.BigEndian, &idmsg0)
					check(err)
					err = binary.Read(conn, binary.BigEndian, &idmsg1)
					check(err)
					msg := MPCMessage{
						Party: PartyID(id),
						Value: val,
						Id0:   idmsg0,
						Id1:   idmsg1,
					}
					cep.Chan <- msg
				} else {
					fmt.Println(cep, "Unknown message")
				}

			}
		}(conn, rp)

		// Sending loop of remote
		go func(conn net.Conn, rp *MPCRemote) {
			var m MPCMessage
			var open = true
			for open {
				m, open = <-rp.Chan
				mpcID := uint64(1)
				check(binary.Write(conn, binary.BigEndian, mpcID))
				check(binary.Write(conn, binary.BigEndian, m.Party))
				check(binary.Write(conn, binary.BigEndian, m.Value))
				check(binary.Write(conn, binary.BigEndian, m.Id0))
				check(binary.Write(conn, binary.BigEndian, m.Id1))
			}
		}(conn, rp)
	}
}

func (cep *MPCProtocol) Run(trusted bool) {
	//fmt.Println(cep, "is running")

	//beaverPart
	if !trusted {
		cep.Triplets = cep.Bp.Run()
		//fmt.Println(cep, "Protocol beaver has terminated")
	}

	//create secret shares and send them
	cep.Input_share = cep.Input
	shares := secret_share(cep.Input, len(cep.Peers))
	for _, peer := range cep.Peers {
		if peer.ID != cep.ID {
			peer.Chan <- MPCMessage{cep.ID, uint64(shares[peer.ID]), 0, 0}
		}
	}
	cep.Secret[cep.ID] = shares[cep.ID]
	//collect shares from other peers
	received := 0
	for m := range cep.Chan {
		//fmt.Println(cep, "received message from", m.Party, ":", m.Value)
		if m.Id0 == 0 {
			cep.Secret[m.Party] = m.Value
			received++
			if received == len(cep.Peers)-1 {
				break
			}
		} else {
			cep.Chan <- m
		}
	}

	//shares are ready, let's compute the circuit
	err := error(nil)
	cep.Output, err = cep.ComputeCircuit()
	if err != nil {
		fmt.Println(err)
	}

	if cep.WaitGroup != nil {
		cep.WaitGroup.Done()
	}
}
