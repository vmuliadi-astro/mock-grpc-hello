package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	health "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type greeterServer struct {
	UnimplementedGreeterServer
	hostname string
}

func (s *greeterServer) SayHello(
	ctx context.Context,
	req *HelloRequest,
) (*HelloReply, error) {

	sleep := time.Duration(req.SleepSec) * time.Second
	if sleep > 0 {
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			return nil, status.Error(codes.DeadlineExceeded, "timeout")
		}
	}

	return &HelloReply{
		Message: fmt.Sprintf("Hello %s from %s", req.Name, s.hostname),
	}, nil
}

func main() {
	addr := flag.String("listen", ":50051", "listen addr")
	flag.Parse()

	hostname, _ := os.Hostname()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	RegisterGreeterServer(s, &greeterServer{hostname: hostname})

	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, hs)

	reflection.Register(s)

	log.Println("gRPC listening on", *addr)
	log.Fatal(s.Serve(lis))
}
