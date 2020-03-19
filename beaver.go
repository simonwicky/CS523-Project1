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
	Party  PartyID
	D      *bfv.Ciphertext
	loopID uint64
	//what you need, don't forget to read the right thing in BindNetwork
}

type BeaverInputs struct {
	//store a and b and c
	A  []uint64
	B  []uint64
	C  []uint64
	SK *bfv.SecretKey
}

type BeaverRemoteParty struct {
	Chan chan BeaverMessage
	*RemoteParty
}

type BeaverProtocol struct {
	Chan       chan BeaverMessage
	Param      *bfv.Parameters
	Nb_triplet int
	Peers      map[PartyID]*BeaverRemoteParty
	Inputs     []*BeaverInputs
	*LocalParty
}

func (lp *LocalParty) NewBeaverProtocol(nb_triplet int) *BeaverProtocol {
	bp := new(BeaverProtocol)
	bp.Chan = make(chan BeaverMessage, 32)
	bp.LocalParty = lp
	bp.Param = bfv.DefaultParams[bfv.PN13QP218]
	bp.Peers = make(map[PartyID]*BeaverRemoteParty, len(lp.Peers))
	bp.Inputs = make([]*BeaverInputs, len(lp.Peers))
	for i, rp := range lp.Peers {
		bp.Peers[i] = &BeaverRemoteParty{
			RemoteParty: rp,
			Chan:        make(chan BeaverMessage, 32),
		}
		bp.Inputs[i] = &BeaverInputs{}
	}
	return bp
}

func (bp *BeaverProtocol) BindNetwork(nw *TCPNetworkStruct) {
	for partyID, conn := range nw.Conns {

		if partyID == bp.ID {
			continue
		}

		rp := bp.Peers[partyID]

		// Sending loop of remote
		go func(conn net.Conn, rp *BeaverRemoteParty) {
			var m BeaverMessage
			var open = true
			for open {
				m, open = <-rp.Chan
				beaverID := uint64(0)
				check(binary.Write(conn, binary.BigEndian, beaverID))
				check(binary.Write(conn, binary.BigEndian, m.Party))
				c, _ := m.D.MarshalBinary()
				//fmt.Println(len(c))
				check(binary.Write(conn, binary.BigEndian, uint64(len(c))))
				check(binary.Write(conn, binary.BigEndian, c))
				check(binary.Write(conn, binary.BigEndian, m.loopID))
			}
		}(conn, rp)
	}
}

func (bp *BeaverProtocol) Run() {
	//here the protocol is actually run
	fmt.Println("Protocol beaver is running for", bp.Nb_triplet, "triplets")
	encoder := bfv.NewEncoder(bp.Param)
	kgen := bfv.NewKeyGenerator(bp.Param)
	//evaluator := bfv.NewEvaluator(bp.Param)

	//first loop
	for i, Pi := range bp.Peers {
		bp.Inputs[i].A = newRandomVec(1<<bp.Param.LogN, bp.Param.T)
		bp.Inputs[i].B = newRandomVec(1<<bp.Param.LogN, bp.Param.T)
		bp.Inputs[i].C = mulVec(bp.Inputs[i].A, bp.Inputs[i].B, bp.Param.T)

		plaintextA := bfv.NewPlaintext(bp.Param)
		encoder.EncodeUint(bp.Inputs[i].A, plaintextA)

		sk, _ := kgen.GenKeyPair()
		bp.Inputs[i].SK = sk
		encryptorSk := bfv.NewEncryptorFromSk(bp.Param, sk)

		cipherextA := encryptorSk.EncryptNew(plaintextA)

		m := BeaverMessage{
			Party:  Pi.ID,
			D:      cipherextA,
			loopID: 0,
		}
		for _, Pj := range bp.Peers {
			if Pj.ID != Pi.ID {
				if Pj.ID == bp.ID {
					//store own d
				} else {
					Pj.Chan <- m
				}
			}
		}
	}

	for i, _ := range bp.Peers {
		for j, _ := range bp.Peers {
			if i != j {
				m := <-bp.Chan
				fmt.Println(bp.ID, m)
			}
		}
	}

	return
}
