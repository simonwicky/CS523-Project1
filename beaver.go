package main

import (
	"encoding/binary"
	"math"
	"net"

	"github.com/ldsec/lattigo/bfv"
	"github.com/ldsec/lattigo/ring"
)

//Beaver Triplet Generation for part 1
func generateBeaverTriplet(n int, mpcProtocol []*MPCProtocol) {
	for i := 0; i < n; i++ {
		a := ring.RandUniform(uint64(MODULUS), 0xffff)
		b := ring.RandUniform(uint64(MODULUS), 0xffff)
		c := (a * b) % MODULUS
		a_shares := secret_share(a, len(mpcProtocol))
		b_shares := secret_share(b, len(mpcProtocol))
		c_shares := secret_share(c, len(mpcProtocol))
		for i, mpcP := range mpcProtocol {
			mpcP.Triplets = append(mpcP.Triplets, [3]uint64{a_shares[i], b_shares[i], c_shares[i]})
		}

	}
}

type BeaverMessage struct {
	Party  PartyID
	D      *bfv.Ciphertext
	loopID uint64
}

type BeaverInputs struct {
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
	Param      *bfv.Parameters
	Nb_triplet uint64
	Peers      map[PartyID]*BeaverRemoteParty
	Inputs     *BeaverInputs
	*LocalParty
}

func (lp *LocalParty) NewBeaverProtocol(nb_triplet int) *BeaverProtocol {
	bp := new(BeaverProtocol)
	bp.Chan = make(chan BeaverMessage, 32)
	bp.LocalParty = lp
	bp.Param = bfv.DefaultParams[bfv.PN13QP218]
	bp.Peers = make(map[PartyID]*BeaverRemoteParty, len(lp.Peers))
	bp.Inputs = &BeaverInputs{}
	bp.Nb_triplet = uint64(nb_triplet)
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

				check(binary.Write(conn, binary.BigEndian, uint64(len(c))))
				check(binary.Write(conn, binary.BigEndian, c))
				check(binary.Write(conn, binary.BigEndian, m.loopID))
			}
		}(conn, rp)
	}
}

func (bp *BeaverProtocol) Run() [][3]uint64 {
	//here the protocol is actually run
	//fmt.Println("Protocol beaver is running for", bp.Nb_triplet, "triplets")
	if bp.Nb_triplet == 0 {
		return nil
	}
	encoder := bfv.NewEncoder(bp.Param)
	kgen := bfv.NewKeyGenerator(bp.Param)
	evaluator := bfv.NewEvaluator(bp.Param)

	//part 1
	bp.Inputs.A = newRandomVec(1<<bp.Param.LogN, bp.Param.T)
	bp.Inputs.B = newRandomVec(1<<bp.Param.LogN, bp.Param.T)
	bp.Inputs.C = mulVec(bp.Inputs.A, bp.Inputs.B, bp.Param.T)

	plaintextA := bfv.NewPlaintext(bp.Param)
	encoder.EncodeUint(bp.Inputs.A, plaintextA)

	sk, _ := kgen.GenKeyPair()
	encryptorSk := bfv.NewEncryptorFromSk(bp.Param, sk)
	decryptorSk := bfv.NewDecryptor(bp.Param, sk)

	ciphertextA := encryptorSk.EncryptNew(plaintextA)

	m := BeaverMessage{
		Party:  bp.ID,
		D:      ciphertextA,
		loopID: 0,
	}
	for _, peer := range bp.Peers {
		if peer.ID != bp.ID {
			peer.Chan <- m
		}
	}

	//part 2
	received := 0
	for m := range bp.Chan {
		if m.loopID != 0 {
			bp.Chan <- m
			continue
		}

		r := newRandomVec(1<<bp.Param.LogN, bp.Param.T)

		bp.Inputs.C = subVec(bp.Inputs.C, r, bp.Param.T)

		plaintextR := bfv.NewPlaintext(bp.Param)
		encoder.EncodeUint(r, plaintextR)

		plaintextB := bfv.NewPlaintext(bp.Param)
		encoder.EncodeUint(bp.Inputs.B, plaintextB)

		//error
		ciphertextE := bfv.NewCiphertext(bp.Param, 1<<bp.Param.LogN)
		context, _ := ring.NewContextWithParams(1<<bp.Param.LogN, bp.Param.Qi)
		error0 := context.NewPoly()
		error1 := context.NewPoly()
		context.SampleGaussian(error0, bp.Param.Sigma, uint64(math.Floor(6*bp.Param.Sigma)))
		context.SampleGaussian(error1, bp.Param.Sigma, uint64(math.Floor(6*bp.Param.Sigma)))
		ciphertextE.SetValue([]*ring.Poly{error0, error1})

		ciphertextD := evaluator.MulNew(m.D, plaintextB)
		dij := evaluator.AddNew(ciphertextD, plaintextR)
		evaluator.Add(dij, ciphertextE, dij)

		msg := BeaverMessage{
			Party:  bp.ID,
			D:      dij,
			loopID: 1,
		}

		for _, peer := range bp.Peers {
			if peer.ID == m.Party {
				peer.Chan <- msg
			}
		}

		received++
		if received == len(bp.Peers)-1 {
			break
		}

	}
	//part 3
	received = 0
	ciphertextC := bfv.NewCiphertext(bp.Param, 1<<bp.Param.LogN)
	for m := range bp.Chan {
		if m.loopID != 1 {
			bp.Chan <- m
			continue
		}
		evaluator.Add(ciphertextC, m.D, ciphertextC)
		received++
		if received == len(bp.Peers)-1 {
			break
		}
	}
	plaintextC := decryptorSk.DecryptNew(ciphertextC)
	c := encoder.DecodeUint(plaintextC)
	bp.Inputs.C = addVec(bp.Inputs.C, c, bp.Param.T)

	triplets := make([][3]uint64, bp.Nb_triplet)
	for i := uint64(0); i < bp.Nb_triplet; i++ {
		triplets[i][0] = bp.Inputs.A[i]
		triplets[i][1] = bp.Inputs.B[i]
		triplets[i][2] = bp.Inputs.C[i]
	}
	return triplets
}
