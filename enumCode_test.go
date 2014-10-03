package rsdic

import (
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"testing"
)

func runTestenumCode(x uint64, t *testing.T) {
	Convey("When encode value", t, func() {
		rankSB := popCount(x)
		code := enumEncode(x, rankSB)
		Convey("The decode value should be equal to x", func() {
			So(enumDecode(code, rankSB), ShouldEqual, x)
		})
		Convey("The bits should be equal to x", func() {
			for i := uint8(0); i < 64; i++ {
				So(enumBit(code, rankSB, i), ShouldEqual, getBit(x, i))
			}
		})
		Convey("The ranks should be equal to x", func() {
			rank := 0
			for i := uint8(0); i < 64; i++ {
				So(enumRank(code, rankSB, i), ShouldEqual, rank)
				if getBit(x, i) {
					rank++
				}
			}
		})
		Convey("The selects should be equal to x", func() {
			onenum := uint8(0)
			zeroNum := uint8(0)
			for i := uint8(0); i < 64; i++ {
				bit := getBit(x, i)
				if bit {
					onenum++
					So(enumSelect(code, rankSB, onenum, bit), ShouldEqual, i)
				} else {
					zeroNum++
					So(enumSelect(code, rankSB, zeroNum, bit), ShouldEqual, i)
				}
			}
		})
		Convey("The runzeros should be equal to x", func() {
			for i := uint8(0); i < 64; i++ {
				runZeros := uint8(0)
				for ; i+runZeros < 64; i++ {
					if getBit(x, i+runZeros) {
						break
					}
				}
				So(enumRunZeros(code, rankSB, i), ShouldEqual, runZeros)
			}
		})
	})
}

func TestenumCode(t *testing.T) {
	runTestenumCode(uint64(0), t)
	testN := 2
	for pc := 0; pc < 64; pc++ {
		for i := 0; i < testN; i++ {
			x := uint64(0)
			for j := 0; j < pc; j++ {
				pos := uint8(rand.Intn(64))
				x |= (1 << pos)
			}
			runTestenumCode(x, t)
		}
	}
}
