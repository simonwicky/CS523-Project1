package main

import "github.com/ldsec/lattigo/ring"

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
