package histogram

import (
	"errors"
	"math"
	"sort"
	"sync/atomic"
)

// Histogram is a histogram implementation.
// Histogram must have bucket boundaries which defines the buckets.
// It will have one more bucket than length of bucket boundaries.
// Values less than first bucket boundary are stored in first bucket.
// Values greater than last bucket boundary are store in last bucket.
// Bucket boundaries must be sorted and all values must be different.
// Negative boundaries are okay.
// All operations are not thread-safe except AtomicIncrement.
// Note: User must make sure that index is valid for all methods which uses an index
// index must belong to [0, len(bucketBoundaries)]
type Histogram struct {
	// bucketBoundaries stores the boundaries between buckets.
	// Values in half-open range [bucketBoundaries[i-1], bucketBoundaries[i])
	// will be stored in bucket[i]
	bucketBoundaries []int64
	// Each Increment() adds one to to the appropriate bucketCounts, add adds the
	// value to bucketsTotal. Size of these slices is one more that size of bucketBoundaries
	bucketCounts []int64
	bucketTotals []int64
	// Track the total number of samples and total value of the samples to allow quick
	// computation of an average
	numSamples int64
	total      int64
}

var (
	emptyError             = errors.New("Slice is empty")
	invalidBoundariesError = errors.New("Invalid bucket boundaries")
)

func New(bucketBoundaries []int64) (*Histogram, error) {
	if bucketBoundaries == nil {
		// length of bucketBoundaries must be atleast one
		return nil, emptyError
	}
	for i := 0; i < len(bucketBoundaries)-1; i++ {
		// Check if the bucketBoundaries are in sorted order
		// and are strictly increasing
		if bucketBoundaries[i] >= bucketBoundaries[i+1] {
			return nil, invalidBoundariesError
		}
	}
	return &Histogram{
		bucketBoundaries: bucketBoundaries,
		bucketCounts:     make([]int64, len(bucketBoundaries)+1),
		bucketTotals:     make([]int64, len(bucketBoundaries)+1),
	}, nil
}

// Increment method inserts a sample into the histogram
func (h *Histogram) Increment(val int64) {
	// A value falls into a bucket i if it is in [bucketBoundaries[i-1], bucketBoundaries[i])
	// Search does a binary search to find the smallest index that matches the search condition
	index := sort.Search(len(h.bucketBoundaries), func(i int) bool {
		return h.bucketBoundaries[i] > val
	})
	h.bucketCounts[index]++
	h.bucketTotals[index] += val
	h.numSamples++
	h.total += val
}

// AtomicIncrement method inserts a sample into the histogram in thread safe manner
func (h *Histogram) AtomicIncrement(val int64) {
	index := sort.Search(len(h.bucketBoundaries), func(i int) bool {
		return h.bucketBoundaries[i] > val
	})
	atomic.AddInt64(&h.bucketCounts[index], 1)
	atomic.AddInt64(&h.bucketTotals[index], val)
	atomic.AddInt64(&h.numSamples, 1)
	atomic.AddInt64(&h.total, val)
}

// BucketRanges method returns the low and high boundaries of this bucket.
func (h *Histogram) BucketRanges(index int) (int64, int64) {
	if index < 0 || index > len(h.bucketBoundaries) {
		panic("index out of bound")
	}
	if index == 0 {
		return math.MinInt64, h.bucketBoundaries[index]
	} else if index == len(h.bucketBoundaries) {
		return h.bucketBoundaries[index-1], math.MaxInt64
	} else {
		return h.bucketBoundaries[index-1], h.bucketBoundaries[index]
	}
}

// BucketCount method returns the number of increments that went into this bucket
func (h *Histogram) BucketCount(index int) int64 {
	return h.bucketCounts[index]
}

// BucketTotal method returns the total of all values inserted to a particular bucket
func (h *Histogram) BucketTotal(index int) int64 {
	return h.bucketTotals[index]
}

// BucketAverage method returns the average of all values inserted to a particular bucket.
func (h *Histogram) BucketAverage(index int) float64 {
	if h.bucketCounts[index] == 0 {
		return 0
	}
	return float64(h.bucketTotals[index]) / float64(h.bucketCounts[index])
}

// Size method returns the number of buckets
func (h *Histogram) Size() int {
	if len(h.bucketCounts) != len(h.bucketTotals) {
		panic("Mismatch in lengths of bucketCounts and bucketTotals")
	}
	return len(h.bucketCounts)
}

// Count method returns the total number of samples in all buckets
func (h *Histogram) Count() int64 {
	return h.numSamples
}

// Total method returns the sum of all samples inserted into the histogram
func (h *Histogram) Total() int64 {
	return h.total
}

// Average method returns the average of all values inserted
func (h *Histogram) Average() float64 {
	if h.numSamples == 0 {
		return float64(0)
	}
	return float64(h.total) / float64(h.numSamples)
}

// Clear method zeros out the buckets
func (h *Histogram) Clear() {
	if len(h.bucketCounts) != len(h.bucketTotals) {
		panic("Mismatch in lengths of bucketCounts and bucketTotals")
	}
	for i := range h.bucketCounts {
		h.bucketCounts[i] = 0
		h.bucketTotals[i] = 0
		h.numSamples = 0
		h.total = 0
	}
}

// IncrementFromHistogram method includes all the samples of other histogram into this.
// This bucketBoundaries used to construct other histogram must be identical to this.
func (h *Histogram) IncrementFromHistogram(other *Histogram) {
	if len(other.bucketBoundaries) != len(h.bucketBoundaries) {
		panic("Mismatch in sizes of  bucketBoundaries")
	}
	for i := 0; i < len(h.bucketCounts); i++ {
		h.bucketCounts[i] += other.bucketCounts[i]
		h.bucketTotals[i] += other.bucketTotals[i]
	}
	h.numSamples += other.numSamples
	h.total += other.total
}

// DecrementFromHistogram method reduces the this bucket by the values in another histogram
func (h *Histogram) DecrementFromHistogram(other *Histogram) {
	if len(other.bucketBoundaries) != len(h.bucketBoundaries) {
		panic("Mismatch in sizes of  bucketBoundaries")
	}
	for i := 0; i < len(h.bucketCounts); i++ {
		h.bucketCounts[i] -= other.bucketCounts[i]
		h.bucketTotals[i] -= other.bucketTotals[i]
	}
	h.numSamples -= other.numSamples
	h.total -= other.total
}

// Copy method makes a deep copy of the histogram
func (h *Histogram) Copy() *Histogram {
	bucketBoundaries := make([]int64, len(h.bucketBoundaries))
	copy(bucketBoundaries, h.bucketBoundaries)
	bucketCounts := make([]int64, len(h.bucketCounts))
	copy(bucketCounts, h.bucketCounts)
	bucketTotals := make([]int64, len(h.bucketTotals))
	copy(bucketTotals, h.bucketTotals)
	return &Histogram{
		bucketBoundaries: bucketBoundaries,
		bucketCounts:     bucketCounts,
		bucketTotals:     bucketTotals,
		numSamples:       h.numSamples,
		total:            h.total,
	}
}
func (h *Histogram) BucketBoundaries() []int64 {
	return h.bucketBoundaries
}
func (h *Histogram) BucketCounts() []int64 {
	return h.bucketCounts
}
