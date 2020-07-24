package main

import (
	"fmt"

	"github.com/hillbig/rsdic"
)

func main() {

	rsd := rsdic.New()
	rsd.PushBack(true)
	rsd.PushBack(false)
	rsd.PushBack(true)
	rsd.PushBack(true)
	// rsd = 1011
	fmt.Printf("%d %d %d\n", rsd.Num(), rsd.OneNum(), rsd.ZeroNum()) // 4 3 1

	// Bit(pos uint64) returns B[pos]
	fmt.Printf("%v\n", rsd.Bit(2)) // true

	// Rank(pos uint64, bit bool) returns the number of bit's in B[0...pos)
	fmt.Printf("%d %d\n", rsd.Rank(2, false), rsd.Rank(4, true)) // 1 3

	// Select(rank uint64, bit bool) returns the position of (rank+1)-th occurence of bit in B.
	oneNum := rsd.OneNum()
	for i := uint64(0); i < oneNum; i++ {
		fmt.Printf("%d:%d\n", i, rsd.Select(i, true))
	}
	// 0:0
	// 1:2
	// 2:3

	rsd.PushBack(false) // You can add anytime

	// Use MarshalBinary() and UnmarshalBinary() for serialize/deserialize RSDic.
	bytes, _ := rsd.MarshalBinary()
	newrsd := rsdic.New()
	_ = newrsd.UnmarshalBinary(bytes)
	// Enjoy !
}
