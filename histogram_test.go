package histogram

import (
	"log"
	"reflect"
	"testing"
)

func TestHistogram(t *testing.T) {
	s1 := []int64{1, 2, 3, 3, 4}
	if _, err := New(s1); err == nil {
		t.Error("Expected error")
	} else {
		log.Println(err)
	}
	s2 := []int64{10, 24, 1, 4}
	if _, err := New(s2); err == nil {
		t.Error("Expected error")
	} else {
		log.Println(err)
	}
	s3 := []int64{1, 2, 3, 4}
	h, err := New(s3)
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	h.AtomicIncrement(1)
	h.AtomicIncrement(1)
	h.AtomicIncrement(2)
	h.AtomicIncrement(3)
	h.AtomicIncrement(0)
	h.AtomicIncrement(4)
	if h.BucketCount(0) != 1 ||
		h.BucketCount(1) != 2 ||
		h.BucketCount(2) != 1 ||
		h.BucketCount(3) != 1 ||
		h.BucketCount(4) != 1 {
		t.Error("Unexpected count")
	}
	for i := 0; i <= len(s3); i++ {
		log.Println(h.BucketRanges(i))
	}
	log.Println(h)
}

func TestRange(t *testing.T) {
	log.Println(Range(1, 10, 3))
	log.Println(Range(10, 1, 3))
	log.Println(Range(1, 10, -3))
	log.Println(Range(10, 1, -3))

	if !reflect.DeepEqual([]int64{1, 4, 7, 10}, Range(1, 10, 3)) {
		t.Error("Range(1, 10, 3) Expected", []int64{1, 4, 7, 10}, "Got", Range(1, 10, 3))
	}
	if !reflect.DeepEqual([]int64{10, 7, 4, 1}, Range(10, 1, -3)) {
		t.Error("Range(10, 1, -3) Expected", []int64{10, 7, 4, 1}, "Got", Range(10, 1, -3))
	}
}
