package cache

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

// TestCacheFieldAlignment verifies that the itemCount field is properly aligned
// for 64-bit atomic operations on ARM architecture.
func TestCacheFieldAlignment(t *testing.T) {
	cache := NewDefault()
	defer cache.Close()

	// Get the address of the itemCount field
	itemCountAddr := uintptr(unsafe.Pointer(&cache.itemCount))

	// On ARM, 64-bit atomic operations require 8-byte alignment
	if itemCountAddr%8 != 0 {
		t.Errorf("itemCount field is not 8-byte aligned. Address: 0x%x, offset: %d", 
			itemCountAddr, itemCountAddr%8)
	}

	// Verify the field is actually at the beginning of the struct
	cacheAddr := uintptr(unsafe.Pointer(cache))
	if itemCountAddr != cacheAddr {
		t.Errorf("itemCount should be at the beginning of the struct. Cache: 0x%x, itemCount: 0x%x", 
			cacheAddr, itemCountAddr)
	}
}

// TestAtomicOperationsOnARM simulates concurrent operations that would fail 
// on ARM with unaligned atomic operations.
func TestAtomicOperationsOnARM(t *testing.T) {
	if runtime.GOARCH != "arm" && runtime.GOARCH != "arm64" {
		t.Skipf("Skipping ARM-specific test on %s", runtime.GOARCH)
	}

	ctx := context.Background()
	cache := NewDefault()
	defer cache.Close()

	const goroutines = 50
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	// This test would panic with "unaligned 64-bit atomic operation" 
	// if the itemCount field is not properly aligned on ARM
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			
			// Perform many operations that trigger atomic operations on itemCount
			for j := 0; j < operationsPerGoroutine; j++ {
				key := "test-key"
				cache.Set(ctx, key, "test-value")
				cache.Get(ctx, key)
				cache.Delete(ctx, key)
			}
		}(i)
	}

	wg.Wait()

	// If we reach here without panic, the alignment is correct
	t.Logf("Successfully completed %d concurrent operations on ARM architecture", 
		goroutines*operationsPerGoroutine*3)
}