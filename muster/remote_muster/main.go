package main

import (
	"fmt"
	"flag"
	"net"
	"context"
	"time"

	//"github.com/awadyn/muster/config"
	pb "github.com/awadyn/muster/muster"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

)

var (
	local_addr = flag.String("addr", "localhost:50052", "the address to connect to")
	port = flag.Int("port", 50051, "remote_muster_port")
)

type remote_muster struct {
	pb.UnimplementedRegistrarServer
}

func (r_m *remote_muster) Register(ctx context.Context, in *pb.MusterRequest) (*pb.MusterReply, error) {
	fmt.Println("Remote muster Register()")
	fmt.Printf("Received: %v\n", in.GetId())
	return &pb.MusterReply{Ip: "10.0.0.1", Ncores: 4}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
	}
	s := grpc.NewServer()
	pb.RegisterRegistrarServer(s, &remote_muster{})
	fmt.Printf("remote muster listening at %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v\n", err)
	}

	conn, err := grpc.Dial(*local_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("did not connect: %v\n", err)
	}
	defer conn.Close()
	c := pb.NewRegistrarClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	time.Sleep(time.Second/5)
	r, err := c.SetupLogs(ctx, &pb.LogsRequest{LogsMapPtr:0x12345})
	if err != nil {
		fmt.Printf("could not setup logs with local muster: %v\n", err)
	}
	fmt.Printf("LogsReply: %s, %d\n", r.GetLogsMapPtr())
}



