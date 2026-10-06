package main

import (
	"context"
	"log"
	"time"

	snowflakepb "snowflake/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := snowflakepb.NewSnowflakeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	status, err := client.GetServerStatus(ctx, &snowflakepb.GetServerStatusRequest{})
	if err != nil {
    log.Fatalf("failed to get status: %v", err)
	}
	log.Printf("Server Status -> WorkerID: %d, ServerTime: %d", status.WorkerId, status.CurrentTimestamp)

	genResp, err := client.GenerateID(ctx, &snowflakepb.GenerateIDRequest{})
	if err != nil {
    log.Fatalf("failed to generate ID: %v", err)
	}
	log.Printf("Generated ID: %d (String: %s)", genResp.Id, genResp.IdStr)

	parseResp, err := client.ParseID(ctx, &snowflakepb.ParseIDRequest{Id: genResp.Id})
	if err != nil {
    log.Fatalf("failed to parse ID: %v", err)
	}
	log.Printf("Parsed ID -> Time: %s, WorkerID: %d, Sequence: %d",
    parseResp.TimestampIso, parseResp.WorkerId, parseResp.Sequence)
}
