package main

func Pmod(x, mod int) int {
	res := x % mod
	if res >= 0 {
		return res
	}
	return res + mod
}
