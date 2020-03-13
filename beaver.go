package main

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/ldsec/lattigo/bfv"
	"github.com/ldsec/lattigo/ring"
)

func generateBeaverTriplet(n int, dummyprotocol []*DummyProtocol) {
	for i := 0; i < n; i++ {
		a := ring.RandUniform(uint64(MODULUS), 0xffff)
		b := ring.RandUniform(uint64(MODULUS), 0xffff)
		c := (a * b) % MODULUS
		a_shares := secret_share(a, len(dummyprotocol))
		b_shares := secret_share(b, len(dummyprotocol))
		c_shares := secret_share(c, len(dummyprotocol))
		for i, dp := range dummyprotocol {
			dp.Triplets = append(dp.Triplets, [3]uint64{a_shares[i], b_shares[i], c_shares[i]})
		}

	}
}

type BeaverMessage struct {
	Party PartyID
	//what you need, don't forget to read the right thing in BindNetwork
}

type BeaverInputs struct {
	//store a and b and c
	A []uint64
	B []uint64
	C []uint64
}

type BeaverRemoteParty struct {
	Chan chan BeaverMessage
	*RemoteParty
}

type BeaverProtocol struct {
	Chan       chan BeaverMessage
	Parameters *bfv.Parameters
	Nb_triplet int
	Peers      map[PartyID]*BeaverRemoteParty
	*BeaverInputs
	*LocalParty
}

func (lp *LocalParty) NewBeaverProtocol(nb_triplet int) *BeaverProtocol {
	bp := new(BeaverProtocol)
	bp.Chan = make(chan BeaverMessage, 32)
	bp.LocalParty = lp
	bp.Parameters = bfv.DefaultParams[bfv.PN13QP218]
	bp.Peers = make(map[PartyID]*BeaverRemoteParty, len(lp.Peers))
	for i, rp := range lp.Peers {
		bp.Peers[i] = &BeaverRemoteParty{
			RemoteParty: rp,
			Chan:        make(chan BeaverMessage, 32),
		}
	}
	return bp
}

func (bp *BeaverProtocol) BindNetwork(nw *TCPNetworkStruct) {
	for partyID, conn := range nw.Conns {

		if partyID == bp.ID {
			continue
		}

		rp := bp.Peers[partyID]

		// Receiving loop from remote
		go func(conn net.Conn) {
			for {
				var id uint64
				var err error
				err = binary.Read(conn, binary.BigEndian, &id)
				check(err)
				// err = binary.Read(conn, binary.BigEndian, &val)
				// check(err)
				// err = binary.Read(conn, binary.BigEndian, &idmsg0)
				// check(err)
				// err = binary.Read(conn, binary.BigEndian, &idmsg1)
				// check(err)
				msg := BeaverMessage{
					Party: PartyID(id),
					// Value: val,
					// Id0:   idmsg0,
					// Id1:   idmsg1,
				}
				//fmt.Println(cep, "receiving", msg, "from", rp)
				bp.Chan <- msg
			}
		}(conn)

		// Sending loop of remote
		go func(conn net.Conn, rp *BeaverRemoteParty) {
			var m BeaverMessage
			var open = true
			for open {
				m, open = <-rp.Chan
				//fmt.Println(cep, "sending", m, "to", rp)
				check(binary.Write(conn, binary.BigEndian, m.Party))
				// check(binary.Write(conn, binary.BigEndian, m.Value))
				// check(binary.Write(conn, binary.BigEndian, m.Id0))
				// check(binary.Write(conn, binary.BigEndian, m.Id1))
			}
		}(conn, rp)
	}
}

func (bp *BeaverProtocol) Run() {
	//here the protocol is actually run
	fmt.Println("Protocol beaver is running")

	//beaver := cep.Bp
	//beaver.A = newRandomVec(...)
	//beaver.B = newRandomVec(...)
	//beaver.C = mulVec(beaver.A, beaver.B,...)
	for _, peer := range bp.Peers {
		close(peer.Chan)
	}
	return
}
