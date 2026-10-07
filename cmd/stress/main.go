package main

import (
	"context"
	"flag"
	"log"
	"sync"
	"sync/atomic"
	"time"

	snowflakepb "snowflake/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	numWorkers := flag.Int("workers", 100, "number of concurrent worker goroutines")
	requestsPerWorker := flag.Int("requests", 1000, "requests per worker")
	target := flag.String("target", "localhost:50051", "gRPC server address")
	flag.Parse()
	totalRequests := *numWorkers * *requestsPerWorker

	conn, err := grpc.NewClient(*target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", *target, err)
	}
	defer conn.Close()

	client := snowflakepb.NewSnowflakeServiceClient(conn)

	log.Printf("Starting stress test against %s: %d concurrent goroutines, %d requests each (%d total IDs)...",
		*target, *numWorkers, *requestsPerWorker, totalRequests)

	var wg sync.WaitGroup
	var completedRequests int64
	var seenIDs sync.Map

	start := time.Now()

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < *requestsPerWorker; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				resp, err := client.GenerateID(ctx, &snowflakepb.GenerateIDRequest{})
				cancel()

				if err != nil {
					log.Printf("Worker %d request %d failed: %v", workerID, j, err)
					return
				}

				if _, exists := seenIDs.LoadOrStore(resp.Id, struct{}{}); exists {
					log.Fatalf("COLLISION DETECTED! ID %d was generated twice!", resp.Id)
				}

				atomic.AddInt64(&completedRequests, 1)
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	log.Printf("Stress test completed in %v", duration)
	log.Printf("Total IDs generated: %d / %d", completedRequests, totalRequests)
	log.Printf("Throughput: %.0f IDs/sec over gRPC network", float64(completedRequests)/duration.Seconds())
}
