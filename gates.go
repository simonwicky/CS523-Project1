package main

func Add_Gate(a, b uint64) uint64 {
	return (a + b) % MODULUS
}

func Sub_Gate(a, b uint64) uint64 {
	return uint64(Pmod(int64(a-b), MODULUS))
}

func AddCst_Gate(id PartyID, a, cst uint64) uint64 {
	if id == 0 {
		return (a + cst) % MODULUS
	}
	return a
}

func MultCst_Gate(a, cst uint64) uint64 {
	return uint64(Pmod(int64(a*cst), MODULUS))
}

func Mult_Gate(x, y uint64, id PartyID, triplet [3]uint64, cep *DummyProtocol) uint64 {
	a := triplet[0]
	b := triplet[1]
	c := triplet[2]

	x_a := Sub_Gate(x, a)
	y_b := Sub_Gate(y, b)
	x_a = Reveal_Gate(cep, x_a)
	y_b = Reveal_Gate(cep, y_b)

	//[z] = [c] + [x] * (y − b) + [y] * (x − a) − (x − a)(y − b)

	// [x] * (y − b)
	term1 := MultCst_Gate(x, y_b)

	//[y] * (x − a)
	term2 := MultCst_Gate(y, x_a)

	//  − (x − a)(y − b)
	term3 := uint64(Pmod(int64(-x_a*y_b), MODULUS))

	// [c] + [x] * (y − b)
	half1 := Add_Gate(c, term1)

	// [y] * (x − a) − (x − a)(y − b)
	half2 := AddCst_Gate(id, term2, term3)

	//[z]
	return Add_Gate(half1, half2)
}

func Reveal_Gate(cep *DummyProtocol, value uint64) (output uint64) {
	//might need to reopen the receiving channel
	output = value
	for _, peer := range cep.Peers {
		if peer.ID != cep.ID {
			peer.Chan <- DummyMessage{cep.ID, value}
		}
	}

	received := 0
	for m := range cep.Chan {
		output = (output + m.Value) % MODULUS
		received++
		if received == len(cep.Peers)-1 {
			//close(cep.Chan)
			break
		}
	}
	return
}
