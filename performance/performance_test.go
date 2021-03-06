package performance

import (
	"testing"
)

func TestConcurrency(t *testing.T) {
	IncreaseConcurrency()
	IncreaseConcurrency()
	if GetConcurrency() != 2 {
		t.Fatalf("increase concurrency excpect 2 but %v", GetConcurrency())
	}
	DecreaseConcurrency()
	if GetConcurrency() != 1 {
		t.Fatalf("decrease concurrency excpect 1 but %v", GetConcurrency())
	}
}

func TestIncreaseRequestCount(t *testing.T) {
	IncreaseRequestCount()
	IncreaseRequestCount()
	count := GetRequstCount()
	if count != 2 {
		t.Fatalf("the request count expect 2 but %v", count)
	}
}
