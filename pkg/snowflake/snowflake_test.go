package snowflake

import (
	"sync"
	"testing"
)

func TestBasicIDGeneration(t *testing.T) {
	node, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("failed to create snowflake node: %v", err)
	}

	id, err := node.NextID()
	if err != nil {
		t.Fatalf("failed to generate ID: %v", err)
	}

	t.Logf("Generated ID (Decimal): %d", id)
	t.Logf("Generated ID (Binary):  %064b", id)

	if id <= 0 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestMonotonicity(t *testing.T) {
	node, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("failed to create snowflake node: %v", err)
	}

	prevID, err := node.NextID()
	if err != nil {
		t.Fatalf("failed to generate initial ID: %v", err)
	}

	for i := 0; i < 5000; i++ {
		id, err := node.NextID()
		if err != nil {
			t.Fatalf("failed to generate ID at iteration %d: %v", i, err)
		}
		if id <= prevID {
			t.Fatalf("ID monotonicity violated: got %d, previous was %d", id, prevID)
		}
		prevID = id
	}
}

func TestConcurrentUniqueness(t *testing.T) {
	node, err := NewSnowflake(42)
	if err != nil {
		t.Fatalf("failed to create snowflake node: %v", err)
	}

	const (
		numGoroutines = 50
		idsPerRoutine = 2000
		totalIDs      = numGoroutines * idsPerRoutine
	)

	var wg sync.WaitGroup
	idChan := make(chan int64, totalIDs)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerRoutine; j++ {
				id, err := node.NextID()
				if err != nil {
					t.Errorf("generation error: %v", err)
					return
				}
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	seen := make(map[int64]struct{}, totalIDs)
	for id := range idChan {
		if _, exists := seen[id]; exists {
			t.Fatalf("Duplicate ID detected: %d (binary: %064b)", id, id)
		}
		seen[id] = struct{}{}
	}

	if len(seen) != totalIDs {
		t.Fatalf("expected %d unique IDs, got %d", totalIDs, len(seen))
	}
	t.Logf("Successfully generated and verified %d unique IDs under high concurrency!", totalIDs)
}

func BenchmarkNextID(b *testing.B) {
	node, err := NewSnowflake(1)
	if err != nil {
		b.Fatalf("failed to create snowflake node: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = node.NextID()
	}
}

// TestDeconstruct generates an ID on a specific worker and verifies the unpacked values.
func TestDeconstruct(t *testing.T) {
	const expectedWorkerID = int64(77)
	node, err := NewSnowflake(expectedWorkerID)
	if err != nil {
		t.Fatalf("failed to create snowflake node: %v", err)
	}

	id, err := node.NextID()
	if err != nil {
		t.Fatalf("failed to generate ID: %v", err)
	}

	parsedTime, parsedWorker, parsedSeq := Deconstruct(id)

	t.Logf("Original ID: %d", id)
	t.Logf("Unpacked Time:     %s", parsedTime.Format("2006-01-02 15:04:05.000 MST"))
	t.Logf("Unpacked WorkerID: %d", parsedWorker)
	t.Logf("Unpacked Sequence: %d", parsedSeq)

	if parsedWorker != expectedWorkerID {
		t.Errorf("expected workerID %d, got %d", expectedWorkerID, parsedWorker)
	}
	if parsedSeq != 0 {
		t.Errorf("expected sequence 0 for first ID, got %d", parsedSeq)
	}
}
