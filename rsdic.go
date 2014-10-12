// Package rsdic provides a rank/select dictionary
// supporting many basic operations in constant time
// using very small working space (smaller than original).
package rsdic

// RSDic provides rank/select operations.
//
// Conceptually RSDic represents a bit vector B[0...num), B[i] = 0 or 1,
// and these bits are set by PushBack (Thus RSDic can handle growing bits).
// All operations (Bit, Rank, Select) are supported in O(1) time.
// (also called as fully indexable dictionary in CS literatures (FID)).
//
// In RSDic, a bit vector is stored in compressed (Note, we don't need to decode all at operations)
// A bit vector is divided into small blocks of length 64, and each small block
// is compressed using enum coding. For example, if a small block contains 10 ones
// and 54 zeros, the block is compressed in 38 bits (See enumCode.go for detail)
// This achieves not only its information theoretic bound, but also achieves more compression
// if same bits appeared togather (e.g. 000...000111...111000...000)
//
// See performance in readme.md
//
// C++ version https://code.google.com/p/rsdic/
// [1] "Fast, Small, Simple Rank/Select on Bitmaps", Gonzalo Navarro and Eliana Providel, SEA 2012

type RSDic interface {
	// PushBack appends the bit to the end of B
	PushBack(bit bool)

	// Num returns the number of bits
	Num() uint64

	// OneNum returns the number of ones in bits
	OneNum() uint64

	// ZeroNum returns the number of zeros in bits
	ZeroNum() uint64

	// Bit returns the (pos+1)-th bit in bits, i.e. bits[pos]
	Bit(pos uint64) bool

	// Rank returns the number of bit's in B[0...pos)
	Rank(pos uint64, bit bool) uint64

	// Select returns the position of (rank+1)-th occurence of bit in B
	// Select returns num if rank+1 is larger than the possible range.
	// (i.e. Select(oneNum, true) = num, Select(zeroNum, false) = num)
	Select(rank uint64, bit bool) uint64

	// BitAndRank returns the (pos+1)-th bit (=b) and Rank(pos, b)
	// Although this is equivalent to b := Bit(pos), r := Rank(pos, b),
	// BitAndRank is faster.
	BitAndRank(pos uint64) (bool, uint64)

	// AllocSize returns the allocated size in bytes.
	AllocSize() int

	// MarshalBinary encodes the RSDic into a binary form and returns the result.
	MarshalBinary() ([]byte, error)

	// UnmarshalBinary decodes the RSDic from a binary from generated MarshalBinary
	UnmarshalBinary([]byte) error
}

// New returns RSDic with a bit array of length 0.
func New() RSDic {
	return &rsdicImpl{
		bits:            make([]uint64, 0),
		pointerBlocks:   make([]uint64, 0),
		rankBlocks:      make([]uint64, 0),
		selectOneInds:   make([]uint64, 0),
		selectZeroInds:  make([]uint64, 0),
		rankSmallBlocks: make([]uint8, 0),
		num:             0,
		oneNum:          0,
		zeroNum:         0,
		lastBlock:       0,
		lastOneNum:      0,
		lastZeroNum:     0,
		codeLen:         0,
	}
}
