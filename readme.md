### Bloomfilters


This package allows you to use bloomfilters in golang. It is a simple and 
straightforward package used for simpler workfloads.


### How to use
```golang
    var bf = NewBloom(OptimalBitSizeValue(100000, 0.001), DefaultHashList...)
	assert.NoError(t, bf.Set([]byte("Hello")))
	assert.NoError(t, bf.Set([]byte("Bob")))
	assert.NoError(t, bf.Set([]byte("Sam")))

    assert.True(t, bf.Test([]byte("Bob")))
	assert.True(t, bf.Test([]byte("Sam")))
	assert.False(t, bf.Test([]byte("Joe")))
```
`DefaultHashList` contains two `fnv` and `murmur3` hash functions. You can add
to the existing list or create a list of your own.