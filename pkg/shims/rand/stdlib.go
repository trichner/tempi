// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rand implements pseudo-random number generators suitable for tasks
// such as simulation, but it should not be used for security-sensitive work.
//
// Random numbers are generated by a [Source], usually wrapped in a [Rand].
// Both types should be used by a single goroutine at a time: sharing among
// multiple goroutines requires some kind of synchronization.
//
// Top-level functions, such as [Float64] and [Int],
// are safe for concurrent use by multiple goroutines.
//
// This package's outputs might be easily predictable regardless of how it's
// seeded. For random numbers suitable for security-sensitive work, see the
// crypto/rand package.
package rand

// A Source represents a source of uniformly-distributed
// pseudo-random int64 values in the range [0, 1<<63).
//
// A Source is not safe for concurrent use by multiple goroutines.
type Source interface {
	Int63() int64
	Seed(seed int64)
}

// A Source64 is a Source that can also generate
// uniformly-distributed pseudo-random uint64 values in
// the range [0, 1<<64) directly.
// If a Rand r's underlying Source s implements Source64,
// then r.Uint64 returns the result of one call to s.Uint64
// instead of making two calls to s.Int63.
type Source64 interface {
	Source
	Uint64() uint64
}

// A Rand is a source of random numbers.
type Rand struct {
	src Source
	s64 Source64 // non-nil if src is source64

	// readVal contains remainder of 63-bit integer used for bytes
	// generation during most recent Read call.
	// It is saved so next Read call can start where the previous
	// one finished.
	readVal int64
	// readPos indicates the number of low-order bytes of readVal
	// that are still valid.
	readPos int8
}

// New returns a new Rand that uses random values from src
// to generate other random values.
func New(src Source) *Rand {
	s64, _ := src.(Source64)
	return &Rand{src: src, s64: s64}
}

// Seed uses the provided seed value to initialize the generator to a deterministic state.
// Seed should not be called concurrently with any other Rand method.
func (r *Rand) Seed(seed int64) {
	r.src.Seed(seed)
	r.readPos = 0
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func (r *Rand) Int63() int64 { return r.src.Int63() }

// Uint32 returns a pseudo-random 32-bit value as a uint32.
func (r *Rand) Uint32() uint32 { return uint32(r.Int63() >> 31) }

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func (r *Rand) Uint64() uint64 {
	if r.s64 != nil {
		return r.s64.Uint64()
	}
	return uint64(r.Int63())>>31 | uint64(r.Int63())<<32
}

// Int31 returns a non-negative pseudo-random 31-bit integer as an int32.
func (r *Rand) Int31() int32 { return int32(r.Int63() >> 32) }

// Int returns a non-negative pseudo-random int.
func (r *Rand) Int() int {
	u := uint(r.Int63())
	return int(u << 1 >> 1) // clear sign bit if int == int32
}

// Int63n returns, as an int64, a non-negative pseudo-random number in the half-open interval [0,n).
// It panics if n <= 0.
func (r *Rand) Int63n(n int64) int64 {
	if n <= 0 {
		panic("invalid argument to Int63n")
	}
	if n&(n-1) == 0 { // n is power of two, can mask
		return r.Int63() & (n - 1)
	}
	max := int64((1 << 63) - 1 - (1<<63)%uint64(n))
	v := r.Int63()
	for v > max {
		v = r.Int63()
	}
	return v % n
}

// Int31n returns, as an int32, a non-negative pseudo-random number in the half-open interval [0,n).
// It panics if n <= 0.
func (r *Rand) Int31n(n int32) int32 {
	if n <= 0 {
		panic("invalid argument to Int31n")
	}
	if n&(n-1) == 0 { // n is power of two, can mask
		return r.Int31() & (n - 1)
	}
	max := int32((1 << 31) - 1 - (1<<31)%uint32(n))
	v := r.Int31()
	for v > max {
		v = r.Int31()
	}
	return v % n
}

// int31n returns, as an int32, a non-negative pseudo-random number in the half-open interval [0,n).
// n must be > 0, but int31n does not check this; the caller must ensure it.
// int31n exists because Int31n is inefficient, but Go 1 compatibility
// requires that the stream of values produced by math/rand remain unchanged.
// int31n can thus only be used internally, by newly introduced APIs.
//
// For implementation details, see:
// https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction
// https://lemire.me/blog/2016/06/30/fast-random-shuffling
func (r *Rand) int31n(n int32) int32 {
	v := r.Uint32()
	prod := uint64(v) * uint64(n)
	low := uint32(prod)
	if low < uint32(n) {
		thresh := uint32(-n) % uint32(n)
		for low < thresh {
			v = r.Uint32()
			prod = uint64(v) * uint64(n)
			low = uint32(prod)
		}
	}
	return int32(prod >> 32)
}

// Intn returns, as an int, a non-negative pseudo-random number in the half-open interval [0,n).
// It panics if n <= 0.
func (r *Rand) Intn(n int) int {
	if n <= 0 {
		panic("invalid argument to Intn")
	}
	if n <= 1<<31-1 {
		return int(r.Int31n(int32(n)))
	}
	return int(r.Int63n(int64(n)))
}

// Float64 returns, as a float64, a pseudo-random number in the half-open interval [0.0,1.0).
func (r *Rand) Float64() float64 {
	// A clearer, simpler implementation would be:
	//	return float64(r.Int63n(1<<53)) / (1<<53)
	// However, Go 1 shipped with
	//	return float64(r.Int63()) / (1 << 63)
	// and we want to preserve that value stream.
	//
	// There is one bug in the value stream: r.Int63() may be so close
	// to 1<<63 that the division rounds up to 1.0, and we've guaranteed
	// that the result is always less than 1.0.
	//
	// We tried to fix this by mapping 1.0 back to 0.0, but since float64
	// values near 0 are much denser than near 1, mapping 1 to 0 caused
	// a theoretically significant overshoot in the probability of returning 0.
	// Instead of that, if we round up to 1, just try again.
	// Getting 1 only happens 1/2⁵³ of the time, so most clients
	// will not observe it anyway.
again:
	f := float64(r.Int63()) / (1 << 63)
	if f == 1 {
		goto again // resample; this branch is taken O(never)
	}
	return f
}

// Float32 returns, as a float32, a pseudo-random number in the half-open interval [0.0,1.0).
func (r *Rand) Float32() float32 {
	// Same rationale as in Float64: we want to preserve the Go 1 value
	// stream except we want to fix it not to return 1.0
	// This only happens 1/2²⁴ of the time (plus the 1/2⁵³ of the time in Float64).
again:
	f := float32(r.Float64())
	if f == 1 {
		goto again // resample; this branch is taken O(very rarely)
	}
	return f
}

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers
// in the half-open interval [0,n).
func (r *Rand) Perm(n int) []int {
	m := make([]int, n)
	// In the following loop, the iteration when i=0 always swaps m[0] with m[0].
	// A change to remove this useless iteration is to assign 1 to i in the init
	// statement. But Perm also effects r. Making this change will affect
	// the final state of r. So this change can't be made for compatibility
	// reasons for Go 1.
	for i := 0; i < n; i++ {
		j := r.Intn(i + 1)
		m[i] = m[j]
		m[j] = i
	}
	return m
}

// Shuffle pseudo-randomizes the order of elements.
// n is the number of elements. Shuffle panics if n < 0.
// swap swaps the elements with indexes i and j.
func (r *Rand) Shuffle(n int, swap func(i, j int)) {
	if n < 0 {
		panic("invalid argument to Shuffle")
	}

	// Fisher-Yates shuffle: https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
	// Shuffle really ought not be called with n that doesn't fit in 32 bits.
	// Not only will it take a very long time, but with 2³¹! possible permutations,
	// there's no way that any PRNG can have a big enough internal state to
	// generate even a minuscule percentage of the possible permutations.
	// Nevertheless, the right API signature accepts an int n, so handle it as best we can.
	i := n - 1
	for ; i > 1<<31-1-1; i-- {
		j := int(r.Int63n(int64(i + 1)))
		swap(i, j)
	}
	for ; i > 0; i-- {
		j := int(r.int31n(int32(i + 1)))
		swap(i, j)
	}
}