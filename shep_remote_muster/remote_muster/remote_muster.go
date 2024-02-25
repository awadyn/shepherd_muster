package main

import (
	"fmt"
	"flag"
	"net"
	"context"
	"os"
	"strconv"
//	"time"

	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
	"google.golang.org/grpc"
	//"google.golang.org/grpc/credentials/insecure"

)

type node struct {
	ncores uint8
	ip string
}

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

type muster struct {
	node
	// CHANNELS
	logs map[string]*log
	controls map[string]*control
	hb_chan chan bool
	id string
}

type remote_muster struct {
	muster
	port *int
	pb.UnimplementedPulseServer
}

func (r_m *remote_muster) HeartBeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatReply, error) {
	fmt.Printf("-- Muster %s received: %v\n", r_m.id, in.GetShepRequest())
	return &pb.HeartbeatReply{MusterReply: r_m.id, ShepRequest: in.GetShepRequest()}, nil
}

func main() {
	n_ip := os.Args[1]
	n_port, err := strconv.Atoi(os.Args[2])
	if err != nil { fmt.Printf("** ** ** Could not parse port number argument: %v\n", err) }
	n := node{ip:n_ip, ncores:4}
	m := muster{node:n, logs: make(map[string]*log), controls: make(map[string]*control), hb_chan: make(chan bool), id: "remote-muster-"+n.ip}
	r_m := remote_muster{muster: m, port: flag.Int("port", n_port, "remote_muster_port") }

	fmt.Println("-- Starting remote muster ", r_m.muster, "with port ", *r_m.port)
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.port))
	if err != nil {
		fmt.Printf("** ** ** Failed to listen: %v\n", err)
	}
	s := grpc.NewServer()
	pb.RegisterPulseServer(s, &r_m)
	fmt.Printf("-- Remote muster listening at %v\n", lis.Addr())

	if err := s.Serve(lis); err != nil {
		fmt.Printf("** ** ** Failed to serve: %v\n", err)
	}
}


