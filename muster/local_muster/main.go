package main

import (
	"fmt"
	"flag"
	"time"
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/awadyn/muster/muster"
)

type log struct {
	core uint8
	l_buff *[][]uint64
	r_buff *[][]uint64
	max_size uint64
	metrics []string
	n_ip string
	id string
	/* e.g. { "log-ep-i", "10.0.0.1", ["joules", "timestamp"], 64KB, 0xdeadbeef, 0x12345678:PORT(i):10.0.0.1, i}
		0xdeadbeef: l_buff  ->  [  [x, 0]
					   [y, 1]
		   			   [z, 2], ...]  */
}

type control struct {
	core uint8
	value uint64
	knob string
	n_ip string
	id string
	/* e.g. { "ctrl-dvfs-i", "10.0.0.1", "dvfs", 0x1234, i }*/
}

type node struct {
	registration_chan chan string
	ncores uint32
	ip string
}


var (
	remote_addr = flag.String("addr", "localhost:50051", "the address to connect to")
	port = flag.Int("port", 50052, "local_muster_port")
)

type local_muster struct {
	node
	logs map[string]*log
	controls map[string]*control
	id string
	/* e.g. {"muster_n", {"ctrl-dvfs-i": {..}, "ctrl-itr-i": {..} ...}, {"log-ep-i": {..}, "log-ep-j": {..}, ...}, node{"10.0.0.1", 24}} */
	pb.UnimplementedRegistrarServer
}

func (l_m *local_muster) SetupLogs(ctx context.Context, in *pb.LogsRequest) (*pb.LogsReply, error) {
	fmt.Printf("Received: %v\n", in.GetLogsMapPtr())
	return &pb.LogsReply{LogsMapPtr: 0x6789a}, nil
}

func main() {
	flag.Parse()
	conn, err := grpc.Dial(*remote_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("did not connect: %v\n", err)
	}
	defer conn.Close()
	c := pb.NewRegistrarClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Register(ctx, &pb.MusterRequest{Id: "muster-10.0.0.1"})
	if err != nil {
		fmt.Printf("could not register with remote muster: %v\n", err)
	}
	fmt.Printf("MusterReply: %s, %d\n", r.GetIp(), r.GetNcores())

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
	}
	s := grpc.NewServer()
	n := node{ip: r.GetIp(), ncores: r.GetNcores()}
	pb.RegisterRegistrarServer(s, &local_muster{node: n, id: "local-muster-10.0.0.1",
					logs: make(map[string]*log), controls: make(map[string]*control)})
	fmt.Printf("local muster listening at %v\n", lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v\n", err)
	}

}



