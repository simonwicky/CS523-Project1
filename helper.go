package main

import "github.com/ldsec/lattigo/ring"

func Pmod(x, mod int64) (res int64) {
	res = x % mod
	if res < 0 {
		res += mod
	}
	return
}

func secret_share(secret uint64, n int) []uint64 {
	secret_share := make([]uint64, n)

	for i := 1; i < n; i++ {
		share := ring.RandUniform(uint64(MODULUS), 0xffff)
		secret = uint64(Pmod(int64(secret)-int64(share), int64(MODULUS)))
		secret_share[i] = share
	}
	secret_share[0] = secret
	return secret_share
}

func newRandomVec(n, T uint64) []uint64 {
	output := make([]uint64, n)
	for i := range output {
		output[i] = ring.RandUniform(T, 0x1ffff)
	}
	return output
}

func addVec(a, b []uint64, T uint64) []uint64 {
	output := make([]uint64, len(a))
	for i := range output {
		output[i] = uint64(Pmod(int64(a[i])+int64(b[i]), int64(T)))
	}
	return output
}
func subVec(a, b []uint64, T uint64) []uint64 {
	output := make([]uint64, len(a))
	for i := range output {
		output[i] = uint64(Pmod(int64(a[i])-int64(b[i]), int64(T)))
	}
	return output
}
func mulVec(a, b []uint64, T uint64) []uint64 {
	output := make([]uint64, len(a))
	for i := range output {
		output[i] = uint64(Pmod(int64(a[i])*int64(b[i]), int64(T)))
	}
	return output
}
func negVec(a, b []uint64, T uint64) []uint64 {
	output := make([]uint64, len(a))
	for i := range output {
		output[i] = uint64(Pmod(-int64(a[i]), int64(T)))
	}
	return output
}
