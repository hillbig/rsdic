package rsdic

import (
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"testing"
)

func TestEmptyRSDic(t *testing.T) {
	Convey("When a bit vector is empty", t, func() {
		rsd := New()
		Convey("The num should be 0", func() {
			So(rsd.Num(), ShouldEqual, 0)
			So(rsd.ZeroNum(), ShouldEqual, 0)
			So(rsd.OneNum(), ShouldEqual, 0)
			So(rsd.Rank(0, true), ShouldEqual, 0)
			So(rsd.AllocSize(), ShouldEqual, 0)
		})
	})
}

type rawBitVector struct {
	orig   []uint8
	ranks  []uint64
	num    uint64
	oneNum uint64
}

func initBitVector(num uint64, ratio float32) (*rawBitVector, *RSDic) {
	orig := make([]uint8, num)
	ranks := make([]uint64, num)
	oneNum := uint64(0)
	rsd := New()
	for i := uint64(0); i < num; i++ {
		ranks[i] = oneNum
		if rand.Float32() > ratio {
			orig[i] = 0
			rsd.PushBack(false)
		} else {
			orig[i] = 1
			rsd.PushBack(true)
			oneNum++
		}
	}
	return &rawBitVector{
		orig,
		ranks,
		num,
		oneNum,
	}, rsd
}

const (
	testNum = 10
)

func runTestRSDic(name string, t *testing.T, rsd *RSDic, raw *rawBitVector) {
	orig := raw.orig
	ranks := raw.ranks
	num := raw.num
	oneNum := raw.oneNum
	Convey(name, t, func() {
		rsd.Select(0, true)
		So(rsd.Num(), ShouldEqual, num)
		So(rsd.OneNum(), ShouldEqual, oneNum)
		So(rsd.Rank(num, true), ShouldEqual, oneNum)
		for i := 0; i < testNum; i++ {
			ind := uint64(rand.Int31n(int32(num)))
			if i == 0 {
				ind = 0 // 0 is special case, and need test
			}
			So(rsd.Bit(ind), ShouldEqual, orig[ind] == 1)
			So(rsd.Rank(ind, false), ShouldEqual, ind-ranks[ind])
			So(rsd.Rank(ind, true), ShouldEqual, ranks[ind])
			bit, rank := rsd.BitAndRank(ind)
			So(bit, ShouldEqual, orig[ind] == 1)
			So(rank, ShouldEqual, bitNum(ranks[ind], ind, bit))
			So(rsd.Select(rank, bit), ShouldEqual, ind)
		}
		out, err := rsd.MarshalBinary()
		So(err, ShouldBeNil)
		newrsd := New()
		err = newrsd.UnmarshalBinary(out)
		So(err, ShouldBeNil)
		for i := 0; i < testNum; i++ {
			ind := uint64(rand.Int31n(int32(num)))
			So(newrsd.Bit(ind), ShouldEqual, orig[ind] == 1)
			So(newrsd.Rank(ind, false), ShouldEqual, ind-ranks[ind])
			So(newrsd.Rank(ind, true), ShouldEqual, ranks[ind])
			bit, rank := rsd.BitAndRank(ind)
			So(bit, ShouldEqual, orig[ind] == 1)
			So(rank, ShouldEqual, bitNum(ranks[ind], ind, bit))
			So(newrsd.Select(rank, bit), ShouldEqual, ind)
		}
	})
}

func TestRandomSmallRSDic(t *testing.T) {
	raw, rsd := initBitVector(100, 0.5)
	runTestRSDic("When a small bit vector is assigned", t, rsd, raw)
}

func TestRandomLargeRSDic(t *testing.T) {
	raw, rsd := initBitVector(100000, 0.5)
	runTestRSDic("When a large bit vector is assigned", t, rsd, raw)
}

func TestRandomVeryLargeRSDic(t *testing.T) {
	raw, rsd := initBitVector(4000000, 0.8)
	runTestRSDic("When a large bit vector is assigned", t, rsd, raw)
}

func TestRandomLargeSparseRSDic(t *testing.T) {
	raw, rsd := initBitVector(100000, 0.01)
	runTestRSDic("When a large sparse bit vector is assigned", t, rsd, raw)
}

func TestRandomAllZeroRSDic(t *testing.T) {
	raw, rsd := initBitVector(100, 0)
	runTestRSDic("When a large zero bit vector is assigned", t, rsd, raw)
}

func setupRSDic(num uint64, ratio float32) *RSDic {
	rsd := New()
	for i := uint64(0); i < num; i++ {
		if rand.Float32() < ratio {
			rsd.PushBack(true)
		} else {
			rsd.PushBack(false)
		}
	}
	return rsd
}

const (
	N = 100000000 // 100Mbit 10^8
)

func BenchmarkDenseRawBit(b *testing.B) {
	raw := make([]uint8, N)
	for i := uint64(0); i < N; i++ {
		if rand.Float32() > 0.5 {
			raw[i] = 1
		} else {
			raw[i] = 0
		}
	}
	dummy := uint64(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dummy += uint64(raw[i%N])
	}
}

func BenchmarkDenseRawRank(b *testing.B) {
	raw := make([]uint8, N)
	for i := uint64(0); i < N; i++ {
		if rand.Float32() > 0.5 {
			raw[i] = 1
		} else {
			raw[i] = 0
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Rank(N, true)
		rank := uint64(0)
		for j := uint64(0); j < N; j++ {
			rank += uint64(raw[j])
		}
	}
}

func BenchmarkDenseRawSelect(b *testing.B) {
	raw := make([]uint8, N)
	oneNum := uint64(0)
	for i := uint64(0); i < N; i++ {
		if rand.Float32() > 0.5 {
			raw[i] = 1
			oneNum++
		} else {
			raw[i] = 0
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rank := uint64(0)
		// Select(oneNum, true)
		for j := uint64(0); j < N; j++ {
			if raw[j] == 1 {
				rank++
				if rank == oneNum {
					break
				}
			}
		}
	}
}

func BenchmarkBit(b *testing.B) {
	rsd := setupRSDic(N, 0.5)
	//	fmt.Printf("%d bytes (%.2f bpc)\n", rsd.AllocSize(), float32(rsd.AllocSize()*8)/N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsd.Bit(uint64(rand.Int31n(int32(N))))
	}
}

func BenchmarkDenseRSDicRank(b *testing.B) {
	rsd := setupRSDic(N, 0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsd.Rank(uint64(rand.Int31n(int32(N))), true)
	}
}

func BenchmarkDenseRSDicSelect(b *testing.B) {
	rsd := setupRSDic(N, 0.5)
	oneNum := rsd.OneNum()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsd.Select(uint64(rand.Int31n(int32(oneNum))), true)
	}
}

func BenchmarkSparseRSDicBit(b *testing.B) {
	rsd := setupRSDic(N, 0.01)
	//fmt.Printf("%d bytes (%.2f)\n", rsd.AllocSize(), float32(rsd.AllocSize()*8)/N)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsd.Bit(uint64(rand.Int31n(int32(N))))
	}
}

func BenchmarkSparseRSDicRank(b *testing.B) {
	rsd := setupRSDic(N, 0.01)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsd.Rank(uint64(rand.Int31n(int32(N))), true)
	}
}

func BenchmarkSparseRSDicSelect(b *testing.B) {
	rsd := setupRSDic(N, 0.01)
	oneNum := rsd.OneNum()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rsd.Select(uint64(rand.Int31n(int32(oneNum))), true)
	}
}
