package bloomfilters

import (
	"errors"
	"hash/fnv"
	"math"
	"sync"
	"sync/atomic"

	"github.com/spaolacci/murmur3"
)

type hashK = func(b []byte) uint64
type BitIndex = uint64 // this type refers to the bit index of the data value, handled using bit ops
type IndexMap = map[uint64][]BitIndex

func makeHash(b []byte) uint64 {
	var k = fnv.New64()
	k.Write(b)
	return k.Sum64()
}

type Bloom struct {
	totalEntriesCount atomic.Uint64
	size              uint64
	bitsize           uint64
	bitsmap           []uint64
	k                 []hashK

	lock *sync.RWMutex
}

// returns a 64 divisible, unsigned rounded-up integer value
// which is an optimal estimation based on the estimated number
// of items to be inserted and desired false
// positive percentage.
// n estimated number of items
// p percentage of the false positive desired
func OptimalValues(n uint64, p float64) (optimalBitArraySize uint64, optimalHashFuncCount uint64) {
	m := (-1 * float64(n)) * math.Log(p) / float64(math.Pow(2.0, float64(math.Log(2))))
	cl := uint64(math.Ceil(m))
	optimalBitArraySize = cl - (cl % 64)

	k := float64(m) / float64(n) * math.Log(2)
	optimalHashFuncCount = uint64(math.Ceil(k))

	return
}

// size automatically rounds down to the nearest number divisible to 64
// hashF a list of hash functions executed in the order they are added
//
// you can use NewBloomOptimal() which uses a community known formula
// to calculate size of bitarray
func NewBloom(size uint64, hashF ...hashK) *Bloom {
	if size < 64 {
		panic("size cannot be less than 64")
	}

	size = size - (size % 64)

	var b = &Bloom{}

	b.size = size / 64
	b.bitsize = size

	b.bitsmap = make([]uint64, size)

	b.k = hashF

	b.lock = &sync.RWMutex{}

	return b
}

// It returns, for each given integer (hash sum), the index array and the bit index
// within the uint64 data value for that specific index.
// the general forumla is simple: s / (n * b) where s is the given
// hash sum, and n is the size of bitarray and b is bitlength (uint64 in our case).
// So, for s = 100, n = 1 and b = 64, it would return
// map[1] = 36
// So, for example for a bitarray size of 1, and s = 1
// it returns map[0]1
func (b *Bloom) findIndexPair(nums []uint64) IndexMap {
	var result = make(IndexMap)
	for _, index := range nums {
		var bitIndex = index % 64
		var mainIndex = (index - bitIndex)
		if mainIndex > 0 {
			mainIndex = mainIndex / 64
		}
		if mainIndex > 0 && mainIndex-1 > b.size {
			mainIndex = mainIndex % b.size
		}
		if _, ok := result[mainIndex]; !ok {
			result[mainIndex] = make([]BitIndex, 0, 1)
		}
		result[mainIndex] = append(result[mainIndex], bitIndex)
	}
	return result
}

func (b *Bloom) setBits(sums []uint64) error {
	defer b.totalEntriesCount.Add(1)
	var indicesPair = b.findIndexPair(sums)
	for mainIndex, bitIndices := range indicesPair {
		for _, bitIndex := range bitIndices {
			// setting specific bit
			b.bitsmap[mainIndex] |= (1 << bitIndex)
		}
	}
	return nil
}

func (b *Bloom) applyHashes(d []byte) []uint64 {
	if len(d) > 0 {
		var result = make([]uint64, len(b.k))
		for n, v := range b.k {
			result[n] = v(d)
		}
		return result
	}

	return nil
}

func (b *Bloom) Set(d []byte) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	var numOfHashes = len(b.k)
	if numOfHashes > 0 {
		var err = b.setBits(b.applyHashes(d))
		return err
	}
	return errors.New("no hash function is defined")
}

func (b *Bloom) Test(d []byte) bool {
	b.lock.RLock()
	defer b.lock.RUnlock()
	var numOfHashes = len(b.k)
	if numOfHashes > 0 {
		var hashes = b.applyHashes(d)
		return b.testIfExists(hashes)
	}
	panic("no hash function is defined")
}

func (b *Bloom) testIfExists(sums []uint64) bool {
	var indices = b.findIndexPair(sums)
	return b.assertBitsArray(indices)
}

func (b *Bloom) assertBitsArray(indices IndexMap) bool {
	var val uint64
	if len(indices) == 0 {
		return false
	}
	for mainIndex, bitIndices := range indices {
		for _, bitIndex := range bitIndices {
			val = (b.bitsmap[mainIndex] >> bitIndex) & 1
			if val == 0 {
				return false
			}
		}
	}
	return true
}

// it is similar to assertBitsArray(), but doesn't return immediately on the first failure
// and as well returns the list of zero-bits; useful for testing or verbose error reporting
func (b *Bloom) checkBitsArray(indices IndexMap) (faultyIndices IndexMap, ok bool) {
	var val uint64
	faultyIndices = make(IndexMap)
	if len(indices) == 0 {
		return nil, false
	}
	for mainIndex, bitIndices := range indices {
		for _, bitIndex := range bitIndices {
			val = (b.bitsmap[mainIndex] >> bitIndex) & 1
			if val == 0 {
				if _, okk := faultyIndices[mainIndex]; !okk {
					faultyIndices[mainIndex] = make([]BitIndex, 0, 1)
				}
				faultyIndices[mainIndex] = append(faultyIndices[mainIndex], bitIndex)
			}
		}
	}
	if len(faultyIndices) > 0 {
		return faultyIndices, false
	}
	return nil, true
}

func (b *Bloom) GetTotalInsertsCount() uint64 {
	return b.totalEntriesCount.Load()
}

func assertBits(value uint64, index BitIndex, expected uint64) bool {
	var current = (value >> index) & 1
	return current == expected
}

func Fnv1(b []byte) uint64 {
	f := fnv.New64()
	f.Write(b)
	return f.Sum64()
}

func Murmur3(b []byte) uint64 {
	f := murmur3.New64()
	f.Write(b)
	return f.Sum64()
}

var DefaultHashList = make([]hashK, 0)

func init() {
	DefaultHashList = append(DefaultHashList, Fnv1)
	DefaultHashList = append(DefaultHashList, Murmur3)
}
