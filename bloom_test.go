package bloomfilters

import (
	"fmt"
	"testing"

	"github.com/tjarratt/babble"

	"github.com/stretchr/testify/assert"
)

func TestBitIndexSimple_MustAssertTrue(t *testing.T) {
	var bf = NewBloom(64, func(b []byte) uint64 {
		return 1
	})
	bf.setBits([]uint64{0, 3, 5})

	fmt.Printf("%064b", bf.bitsmap[0])
	assert.True(t, assertBits(bf.bitsmap[0], 0, 1))
	assert.True(t, assertBits(bf.bitsmap[0], 3, 1))
	assert.True(t, assertBits(bf.bitsmap[0], 5, 1))
	assert.True(t, assertBits(bf.bitsmap[0], 6, 0))
}

func TestBitIndex_MultipleIndices_MustAssertTrue(t *testing.T) {
	var bf = NewBloom(64, func(b []byte) uint64 {
		return 1
	})
	bf.setBits([]uint64{0, 3, 5, 800})

	fmt.Printf("%064b", bf.bitsmap[0])
	assert.True(t, bf.testIfExists([]uint64{0, 3, 5}))
	failedIndices, ok := bf.checkBitsArray(bf.findIndexPair([]uint64{0, 3, 5, 800}))
	assert.Empty(t, failedIndices)
	assert.True(t, ok)
}

func TestBitIndex_MultipleIndices_MustFail(t *testing.T) {
	var bf = NewBloom(64, func(b []byte) uint64 {
		return 1
	})
	bf.setBits([]uint64{0, 3, 5, 800})

	fmt.Printf("%064b", bf.bitsmap[0])
	assert.True(t, bf.testIfExists([]uint64{0, 3, 5}))
	failedIndices, ok := bf.checkBitsArray(bf.findIndexPair([]uint64{55, 3, 5, 801}))
	assert.NotEmpty(t, failedIndices)
	assert.False(t, ok)
	if len(failedIndices) > 0 {
		assert.Contains(t, failedIndices, uint64(0))
		assert.Contains(t, failedIndices[0], uint64(55)) // for 55
		assert.Contains(t, failedIndices[0], uint64(33)) // for 801
	}
}

func TestBitIndex_BigArray_MustAssertTrue(t *testing.T) {
	var bf = NewBloom(64*1000, func(b []byte) uint64 {
		return 1
	})
	bf.setBits([]uint64{64*1000 + 32})
	failedIndices, ok := bf.checkBitsArray(bf.findIndexPair([]uint64{64*1000 + 32}))
	assert.Empty(t, failedIndices)
	assert.True(t, ok)
	fmt.Printf("%064b", bf.bitsmap[1000])
	var n = uint64(0)
	n = bf.bitsmap[1000] >> 32 & 1
	assert.Equal(t, uint64(1), n)
}

func TestBitIndex_BigArray_MustFail(t *testing.T) {
	var bf = NewBloom(64*1000, func(b []byte) uint64 {
		return 1
	})
	bf.setBits([]uint64{64*1000 + 32})
	failedIndices, ok := bf.checkBitsArray(bf.findIndexPair([]uint64{64*1000 + 33}))
	assert.NotEmpty(t, failedIndices)
	assert.False(t, ok)
	fmt.Printf("%064b", bf.bitsmap[1000])
	var n = uint64(0)
	n = bf.bitsmap[1000] >> 32 & 1
	assert.Equal(t, uint64(1), n)
	n = uint64(0)
	n = bf.bitsmap[1000] >> 33 & 1
	assert.Equal(t, uint64(0), n)
}

func Test_RealWorld_Usage(t *testing.T) {
	m, k := OptimalValues(100000, 0.001)
	assert.NotZero(t, m)
	assert.NotZero(t, k)
	var bf = NewBloom(m, DefaultHashList...)
	assert.NoError(t, bf.Set([]byte("Hello")))
	assert.NoError(t, bf.Set([]byte("Bob")))
	assert.NoError(t, bf.Set([]byte("Sam")))
	assert.True(t, bf.Test([]byte("Hello")))
	assert.True(t, bf.Test([]byte("Bob")))
	assert.True(t, bf.Test([]byte("Sam")))
	assert.False(t, bf.Test([]byte("Joe")))

	assert.Equal(t, uint64(3), bf.GetTotalInsertsCount())
}

func Benchmark_Bloom_BigInsertion(b *testing.B) {
	m, _ := OptimalValues(10_000_000, 0.001)
	var bf = NewBloom(m, DefaultHashList...)
	babbler := babble.NewBabbler()

	for b.Loop() {
		w := babbler.Babble()
		bf.Set([]byte(w))
		bf.Test([]byte(w))
	}
}
