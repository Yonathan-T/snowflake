package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"snowflake/pkg/snowflake"
	snowflakepb "snowflake/proto"

	"google.golang.org/grpc"
)

type server struct {
	snowflakepb.UnimplementedSnowflakeServiceServer
	node *snowflake.Snowflake
}

func (s *server) GenerateID(ctx context.Context, req *snowflakepb.GenerateIDRequest) (*snowflakepb.GenerateIDResponse, error) {
	id, err := s.node.NextID()
	if err != nil {
		return nil, err
	}
	return &snowflakepb.GenerateIDResponse{
		Id:    id,
		IdStr: strconv.FormatInt(id, 10),
	}, nil
}
func (s *server) ParseID(ctx context.Context, req *snowflakepb.ParseIDRequest) (*snowflakepb.ParseIDResponse, error) {
	t, workerID, seq := snowflake.Deconstruct(req.Id)
	return &snowflakepb.ParseIDResponse{
		Id:           req.Id,
		TimestampMs:  t.UnixMilli(),
		TimestampIso: t.UTC().Format(time.RFC3339),
		WorkerId:     workerID,
		Sequence:     seq,
	}, nil
}

func (s *server) GetServerStatus(ctx context.Context, req *snowflakepb.GetServerStatusRequest) (*snowflakepb.GetServerStatusResponse, error) {
	return &snowflakepb.GetServerStatusResponse{
		WorkerId:         s.node.WorkerID(),
		CurrentTimestamp: time.Now().UnixMilli(),
		Epoch:            snowflake.DefaultEpoch,
	}, nil
}

func main() {
	port := flag.Int("port", 50051, "gRPC server port")
	workerID := flag.Int64("worker", 1, "worker ID")
	flag.Parse()

	node, err := snowflake.NewSnowflake(*workerID)
	if err != nil {
		log.Fatalf("failed to create a snowflake node: %v", err)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", *port, err)
	}

	grpcServer := grpc.NewServer()
	snowflakepb.RegisterSnowflakeServiceServer(grpcServer, &server{node: node})
	log.Printf("Snowflake gRPC server listening on port %d (Worker ID: %d)", *port, *workerID)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
