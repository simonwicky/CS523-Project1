package main

func Add_Gate(a, b uint64) uint64 {
	return (a + b) % MODULUS
}

func Sub_Gate(a, b uint64) uint64 {
	a_int := int(a)
	b_int := int(b)

	return uint64(Pmod(a_int-b_int, MODULUS))
}

func AddCst_Gate(id PartyID, a, cst uint64) uint64 {
	if id == 0 {
		return (a + cst) % MODULUS
	}
	return a
}

func MultCst_Gate(a, cst uint64) uint64 {
	a_int := int(a)
	cst_int := int(cst)

	return uint64(Pmod(a_int*cst_int, MODULUS))
}

func Mult_Gate(a, b uint64) uint64 {
	return 0
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
		output += m.Value
		received++
		if received == len(cep.Peers)-1 {
			close(cep.Chan)
		}
	}

	return
}
