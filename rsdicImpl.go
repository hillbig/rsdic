package rsdic

import (
	"bytes"
	"github.com/ugorji/go/codec"
)

type rsdicImpl struct {
	bits            []uint64
	pointerBlocks   []uint64
	rankBlocks      []uint64
	selectOneInds   []uint64
	selectZeroInds  []uint64
	rankSmallBlocks []uint8
	num             uint64
	oneNum          uint64
	zeroNum         uint64
	lastBlock       uint64
	lastOneNum      uint64
	lastZeroNum     uint64
	codeLen         uint64
}

func (rs rsdicImpl) Num() uint64 {
	return rs.num
}

func (rs rsdicImpl) OneNum() uint64 {
	return rs.oneNum
}

func (rs rsdicImpl) ZeroNum() uint64 {
	return rs.zeroNum
}

func (rs *rsdicImpl) PushBack(bit bool) {
	if (rs.num % kSmallBlockSize) == 0 {
		rs.writeBlock()
	}
	if bit {
		rs.lastBlock |= (1 << (rs.num % kSmallBlockSize))
		if (rs.oneNum % kSelectBlockSize) == 0 {
			rs.selectOneInds = append(rs.selectOneInds, rs.num/kLargeBlockSize)
		}
		rs.oneNum++
		rs.lastOneNum++
	} else {
		if (rs.zeroNum % kSelectBlockSize) == 0 {
			rs.selectZeroInds = append(rs.selectZeroInds, rs.num/kLargeBlockSize)
		}
		rs.zeroNum++
		rs.lastZeroNum++
	}
	rs.num++
}

func (rs *rsdicImpl) writeBlock() {
	if rs.num > 0 {
		rankSB := uint8(rs.lastOneNum)
		rs.rankSmallBlocks = append(rs.rankSmallBlocks, rankSB)
		codeLen := kEnumCodeLength[rankSB]
		code := enumEncode(rs.lastBlock, rankSB)
		newSize := floor(rs.codeLen+uint64(codeLen), kSmallBlockSize)
		if newSize > uint64(len(rs.bits)) {
			rs.bits = append(rs.bits, 0)
		}
		setSlice(rs.bits, rs.codeLen, codeLen, code)
		rs.lastBlock = 0
		rs.lastZeroNum = 0
		rs.lastOneNum = 0
		rs.codeLen += uint64(codeLen)
	}
	if (rs.num % kLargeBlockSize) == 0 {
		rs.rankBlocks = append(rs.rankBlocks, rs.oneNum)
		rs.pointerBlocks = append(rs.pointerBlocks, rs.codeLen)
	}
}

func (rs rsdicImpl) lastBlockInd() uint64 {
	if rs.num == 0 {
		return 0
	}
	return ((rs.num - 1) / kSmallBlockSize) * kSmallBlockSize
}

func (rs rsdicImpl) isLastBlock(pos uint64) bool {
	return pos >= rs.lastBlockInd()
}

func (rs rsdicImpl) Bit(pos uint64) bool {
	if rs.isLastBlock(pos) {
		return getBit(rs.lastBlock, uint8(pos%kSmallBlockSize))
	}
	lblock := pos / kLargeBlockSize
	pointer := rs.pointerBlocks[lblock]
	sblock := pos / kSmallBlockSize
	for i := lblock * kSmallBlockPerLargeBlock; i < sblock; i++ {
		pointer += uint64(kEnumCodeLength[rs.rankSmallBlocks[i]])
	}
	rankSB := rs.rankSmallBlocks[sblock]
	code := getSlice(rs.bits, pointer, kEnumCodeLength[rankSB])
	return enumBit(code, rankSB, uint8(pos%kSmallBlockSize))
}

func (rs rsdicImpl) Rank(pos uint64, bit bool) uint64 {
	if pos >= rs.num {
		return bitNum(rs.oneNum, rs.num, bit)
	}
	if rs.isLastBlock(pos) {
		afterRank := popCount(rs.lastBlock >> (pos % kSmallBlockSize))
		return bitNum(rs.oneNum-uint64(afterRank), pos, bit)
	}
	lblock := pos / kLargeBlockSize
	pointer := rs.pointerBlocks[lblock]
	sblock := pos / kSmallBlockSize
	rank := rs.rankBlocks[lblock]
	for i := lblock * kSmallBlockPerLargeBlock; i < sblock; i++ {
		rankSB := rs.rankSmallBlocks[i]
		pointer += uint64(kEnumCodeLength[rankSB])
		rank += uint64(rankSB)
	}
	if pos%kSmallBlockSize == 0 {
		return bitNum(rank, pos, bit)
	}
	rankSB := rs.rankSmallBlocks[sblock]
	code := getSlice(rs.bits, pointer, kEnumCodeLength[rankSB])
	rank += uint64(enumRank(code, rankSB, uint8(pos%kSmallBlockSize)))
	return bitNum(rank, pos, bit)
}

func (rs rsdicImpl) Select(rank uint64, bit bool) uint64 {
	if bit {
		return rs.Select1(rank)
	} else {
		return rs.Select0(rank)
	}
}

func (rs rsdicImpl) Select1(rank uint64) uint64 {
	if rank >= rs.oneNum-rs.lastOneNum {
		lastBlockRank := uint8(rank - (rs.oneNum - rs.lastOneNum))
		return rs.lastBlockInd() + uint64(selectRaw(rs.lastBlock, lastBlockRank+1))
	}
	selectInd := rank / kSelectBlockSize
	lblock := rs.selectOneInds[selectInd]
	for ; lblock < uint64(len(rs.rankBlocks)); lblock++ {
		if rank < rs.rankBlocks[lblock] {
			break
		}
	}
	lblock--
	sblock := lblock * kSmallBlockPerLargeBlock
	pointer := rs.pointerBlocks[lblock]
	remain := rank - rs.rankBlocks[lblock] + 1
	for ; sblock < uint64(len(rs.rankSmallBlocks)); sblock++ {
		rankSB := rs.rankSmallBlocks[sblock]
		if remain <= uint64(rankSB) {
			break
		}
		remain -= uint64(rankSB)
		pointer += uint64(kEnumCodeLength[rankSB])
	}
	rankSB := rs.rankSmallBlocks[sblock]
	code := getSlice(rs.bits, pointer, kEnumCodeLength[rankSB])
	return sblock*kSmallBlockSize + uint64(enumSelect1(code, rankSB, uint8(remain)))
}

func (rs rsdicImpl) Select0(rank uint64) uint64 {
	if rank >= rs.zeroNum-rs.lastZeroNum {
		lastBlockRank := uint8(rank - (rs.zeroNum - rs.lastZeroNum))
		return rs.lastBlockInd() + uint64(selectRaw(^rs.lastBlock, lastBlockRank+1))
	}
	selectInd := rank / kSelectBlockSize
	lblock := rs.selectZeroInds[selectInd]
	for ; lblock < uint64(len(rs.rankBlocks)); lblock++ {
		if rank < lblock*kLargeBlockSize-rs.rankBlocks[lblock] {
			break
		}
	}
	lblock--
	sblock := lblock * kSmallBlockPerLargeBlock
	pointer := rs.pointerBlocks[lblock]
	remain := rank - lblock*kLargeBlockSize + rs.rankBlocks[lblock] + 1
	for ; sblock < uint64(len(rs.rankSmallBlocks)); sblock++ {
		rankSB := kSmallBlockSize - rs.rankSmallBlocks[sblock]
		if remain <= uint64(rankSB) {
			break
		}
		remain -= uint64(rankSB)
		pointer += uint64(kEnumCodeLength[rankSB])
	}
	rankSB := rs.rankSmallBlocks[sblock]
	code := getSlice(rs.bits, pointer, kEnumCodeLength[rankSB])
	return sblock*kSmallBlockSize + uint64(enumSelect0(code, rankSB, uint8(remain)))
}

func (rs rsdicImpl) BitAndRank(pos uint64) (bool, uint64) {
	if rs.isLastBlock(pos) {
		offset := uint8(pos % kSmallBlockSize)
		bit := getBit(rs.lastBlock, offset)
		afterRank := uint64(popCount(rs.lastBlock >> offset))
		return bit, bitNum(rs.oneNum-afterRank, pos, bit)
	}
	lblock := pos / kLargeBlockSize
	pointer := rs.pointerBlocks[lblock]
	sblock := pos / kSmallBlockSize
	rank := rs.rankBlocks[lblock]
	for i := lblock * kSmallBlockPerLargeBlock; i < sblock; i++ {
		rankSB := rs.rankSmallBlocks[i]
		pointer += uint64(kEnumCodeLength[rankSB])
		rank += uint64(rankSB)
	}
	rankSB := rs.rankSmallBlocks[sblock]
	code := getSlice(rs.bits, pointer, kEnumCodeLength[rankSB])
	rank += uint64(enumRank(code, rankSB, uint8(pos%kSmallBlockSize)))
	bit := enumBit(code, rankSB, uint8(pos%kSmallBlockSize))
	return bit, bitNum(rank, pos, bit)
}

func (rs rsdicImpl) RunZeros(pos uint64) uint64 {
	if rs.isLastBlock(pos) {
		offset := uint8(pos - rs.lastBlockInd())
		lastRunZero := uint64(runZerosRaw(rs.lastBlock, offset))
		if lastRunZero <= rs.num-rs.lastBlockInd()-uint64(offset) {
			return lastRunZero
		} else {
			return rs.num - rs.lastBlockInd() - uint64(offset)
		}
	}
	lblock := pos / kLargeBlockSize
	pointer := rs.pointerBlocks[lblock]
	sblock := pos / kSmallBlockSize
	for i := lblock * kSmallBlockPerLargeBlock; i < sblock; i++ {
		pointer += uint64(kEnumCodeLength[rs.rankSmallBlocks[i]])
	}
	ret := uint64(0)
	offset := uint8(pos % kSmallBlockSize)
	rankSB := rs.rankSmallBlocks[sblock]
	code := getSlice(rs.bits, pointer, kEnumCodeLength[rankSB])
	runZeros := enumRunZeros(code, rankSB, offset)
	ret += uint64(runZeros)
	if uint64(offset)+uint64(runZeros) < kSmallBlockSize {
		return ret
	}
	// Since zero continues beyond a small block,
	// use rank/select to make sure O(1) time.
	rank := rs.Rank(pos, true)
	if rank+1 < rs.OneNum() {
		return rs.Select1(rank) - pos
	} else {
		return rs.Num() - pos
	}
}

func (rsd rsdicImpl) AllocSize() int {
	return len(rsd.bits)*8 +
		len(rsd.pointerBlocks)*8 +
		len(rsd.rankBlocks)*8 +
		len(rsd.selectOneInds)*8 +
		len(rsd.selectZeroInds)*8 +
		len(rsd.rankSmallBlocks)*1
}

func (rsd rsdicImpl) MarshalBinary() (out []byte, err error) {
	w := new(bytes.Buffer)
	var bh codec.BincHandle
	enc := codec.NewEncoder(w, &bh)
	err = enc.Encode(rsd.bits)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.pointerBlocks)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.rankBlocks)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.selectOneInds)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.selectZeroInds)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.rankSmallBlocks)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.num)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.oneNum)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.zeroNum)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.lastBlock)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.lastOneNum)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.lastZeroNum)
	if err != nil {
		return
	}
	err = enc.Encode(rsd.codeLen)
	if err != nil {
		return
	}
	out = w.Bytes()
	return
}

func (rsd *rsdicImpl) UnmarshalBinary(in []byte) (err error) {
	r := bytes.NewBuffer(in)
	var bh codec.BincHandle
	dec := codec.NewDecoder(r, &bh)
	err = dec.Decode(&rsd.bits)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.pointerBlocks)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.rankBlocks)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.selectOneInds)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.selectZeroInds)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.rankSmallBlocks)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.num)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.oneNum)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.zeroNum)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.lastBlock)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.lastOneNum)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.lastZeroNum)
	if err != nil {
		return
	}
	err = dec.Decode(&rsd.codeLen)
	if err != nil {
		return
	}
	return nil
}
