package main

func Add_Gate(a, b uint64) uint64 {
	return (a + b) % MODULUS
}

func Sub_Gate(a, b uint64) uint64 {
	return uint64(Pmod(int64(a-b), int64(MODULUS)))
}

func AddCst_Gate(id PartyID, a, cst uint64) uint64 {
	if id == 0 {
		return (a + cst) % MODULUS
	}
	return a
}

func MultCst_Gate(a, cst uint64) uint64 {
	return uint64(Pmod(int64(a*cst), int64(MODULUS)))
}

func Mult_Gate(x, y uint64, gateID uint64, triplet [3]uint64, cep *DummyProtocol) uint64 {
	a := triplet[0]
	b := triplet[1]
	c := triplet[2]

	x_a := Sub_Gate(x, a)
	y_b := Sub_Gate(y, b)
	x_a = reveal_gate(cep, x_a, [2]uint64{gateID, 0})
	y_b = reveal_gate(cep, y_b, [2]uint64{gateID, 1})

	//[z] = [c] + [x] * (y − b) + [y] * (x − a) − (x − a)(y − b)

	// [x] * (y − b)
	term1 := MultCst_Gate(x, y_b)

	//[y] * (x − a)
	term2 := MultCst_Gate(y, x_a)

	//  − (x − a)(y − b)
	term3 := uint64(Pmod(int64(-x_a*y_b), int64(MODULUS)))

	// [c] + [x] * (y − b)
	half1 := Add_Gate(c, term1)

	// [y] * (x − a) − (x − a)(y − b)
	half2 := AddCst_Gate(cep.ID, term2, term3)

	//[z]
	return Add_Gate(half1, half2)
}

func reveal_gate(cep *DummyProtocol, value uint64, id [2]uint64) (output uint64) {
	output = value
	for _, peer := range cep.Peers {
		if peer.ID != cep.ID {
			peer.Chan <- DummyMessage{cep.ID, value, id[0], id[1]}
		}
	}

	received := 0
	for m := range cep.Chan {
		if m.Id0 == id[0] && m.Id1 == id[1] {
			output = (output + m.Value) % MODULUS
			received++
			if received == len(cep.Peers)-1 {
				break
			}
		} else {
			cep.Chan <- m
		}
	}
	return
}

func Reveal_Gate(cep *DummyProtocol, value uint64, gateID uint64) uint64 {
	return reveal_gate(cep, value, [2]uint64{gateID, 0})
}
