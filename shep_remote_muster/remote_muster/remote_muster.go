package main

import (
	"fmt"
	"flag"
	"net"
	"context"
	"os"
	"strconv"
	"time"

	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*********************************************/
func (r_m *remote_muster) init(n_ip string, n_cores int, n_pulse_port int, n_ctrl_port int, local_muster_port string) {
	n := node{ip:n_ip, ncores: uint8(n_cores)}
	m := muster{node:n, 
		    id: "muster-" + n.ip,
		    logs: make(map[string]*log), 
		    controls: make(map[string]*control), 
		    hb_chan: make(chan bool),
		    full_buff_chan: make(chan string)}
	r_m.muster = m
	r_m.pulse_port = flag.Int("pulse_port", n_pulse_port, "remote_muster_pulse_port")
	r_m.ctrl_port = flag.Int("ctrl_port", n_ctrl_port, "remote_muster_ctrl_port")
	r_m.local_muster_addr = flag.String("local_muster_addr_" + m.id, 
					    "localhost:" + local_muster_port, 
					    "address of mirror local_muster log sync server")
	r_m.show()
	var core uint8
	for core = 0; core < n.ncores; core ++ {
		log_id := "log-" + strconv.Itoa(int(core)) + "-" + r_m.ip
		var max_size uint64 = 4096
		mem_buff := make([][]uint64, max_size)
		r_m.logs[log_id] = &log{id: log_id,
					metrics: []string{"timestamp", "joules"},
					max_size: max_size,
					r_buff: &mem_buff,
					core: core,
					ready_buff_chan: make(chan bool, 1)}
		r_m.logs[log_id].show()
	}
}

/*****************/
/* REMOTE LOGGER */
/*****************/

/* This function represents a single remote_muster logger thread
   which is a client of the shepherd server, which in turn performs
   all log processing and control computation relevant to this 
   remote_muster as well as all other remote_musters the shepherd
   is in charge of. 
*/
func (r_m *remote_muster) log(conn *grpc.ClientConn, c pb.LogClient, ctx context.Context, cancel context.CancelFunc) {
	<- r_m.hb_chan
	defer conn.Close()
	defer cancel()
	for log_id, log := range(r_m.logs) { go r_m.simulate_remote_log(log.core, log_id) }
	for {
		select {
		case log_id := <- r_m.full_buff_chan:
			for {
				fmt.Println("------------FULL_BUFF-- ", log_id)
				stream, err := c.SyncLogBuffers(ctx)
				if err != nil {
					fmt.Printf("** ** ** ERROR: %v could not send sync log request for %v:\n** ** ** %v\n", r_m.id, log_id, err)
					continue
				}
				for _, log_entry := range *(r_m.logs[log_id].r_buff) {
					for {
						err := stream.Send(&pb.SyncLogRequest{LogId:log_id, LogEntry: &pb.LogEntry{Vals: log_entry}})
						if err != nil { 
							fmt.Printf("** ** ** ERROR: %v %v could not send log entry %v:\n** ** **%v\n", r_m.id, log_id, log_entry, err)
							continue
						}
						break
					}
				}
				r, err := stream.CloseAndRecv()
				if err != nil {
					fmt.Printf("** ** ** ERROR: %v problem receiving sync log reply %v:\n** ** ** %v\n", r_m.id, log_id, err)
					continue
				}
				fmt.Printf("------------SYNC-REP-- %v %v\n", log_id, r.GetSyncComplete())
				if r.GetSyncComplete() {
					r_m.logs[log_id].ready_buff_chan <- true
					break
				}
			}
		}
	}
}

func (r_m *remote_muster) start_remote_logger() {
	<- r_m.hb_chan

	fmt.Println("-------------------------------------------------------------")
	fmt.Printf("STARTING REMOTE MUSTER LOGGER %v\n", r_m.id)
	fmt.Println("-------------------------------------------------------------")

	conn, err := grpc.Dial(*r_m.local_muster_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v could not create connection to local muster %s:\n** ** ** %v\n", r_m.id, *r_m.local_muster_addr, err)
	}
	c := pb.NewLogClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)

	go r_m.log(conn, c, ctx, cancel)
}

/*****************/
/* REMOTE PULSER */
/*****************/

func (r_m *remote_muster) HeartBeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatReply, error) {
	fmt.Printf("------HB-REQ-- %v\n", in.GetShepRequest())
	select {
	case r_m.hb_chan <- true:
	default:
	}
	return &pb.HeartbeatReply{MusterReply: r_m.id, ShepRequest: in.GetShepRequest()}, nil
}

func (r_m *remote_muster) start_remote_pulser() {
	fmt.Println("-------------------------------------------------------------")
	fmt.Printf("STARTING REMOTE MUSTER PULSER %v\n", r_m.id)
	fmt.Println("-------------------------------------------------------------")

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.pulse_port))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to listen: %v\n", r_m.id, err)
	}
	fmt.Printf("-- %v -- listening at %v\n", r_m.id, lis.Addr())
	s := grpc.NewServer()
	pb.RegisterPulseServer(s, r_m)
	fmt.Printf("-- %v -- starting heartbeat server\n", r_m.id)
	if err := s.Serve(lis); err != nil {
		fmt.Printf("** ** ** ERROR: failed to serve: %v\n", err)
	}
}

/*********************/
/* REMOTE CONTROLLER */
/*********************/

func (r_m *remote_muster) ApplyControl(stream pb.Control_ApplyControlServer) error {
	fmt.Printf("------CTRL-REQ-- \n")
	return stream.SendAndClose(&pb.ControlReply{CtrlComplete: true})
}

func (r_m *remote_muster) start_remote_controller() {
	<- r_m.hb_chan
	fmt.Println("-------------------------------------------------------------")
	fmt.Printf("STARTING REMOTE MUSTER CONTROLLER %v\n", r_m.id)
	fmt.Println("-------------------------------------------------------------")

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.ctrl_port))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to listen on control port: %v\n", r_m.id, err)
	}
	fmt.Printf("-- %v -- listening at %v\n", r_m.id, lis.Addr())
	s := grpc.NewServer()
	pb.RegisterControlServer(s, r_m)
	fmt.Printf("-- %v -- starting control server\n", r_m.id)
	if err := s.Serve(lis); err != nil {
		fmt.Printf("** ** ** ERROR: failed to serve: %v\n", err)
	}
}

/*****************/

func main() {
	n_ip := os.Args[1]
	n_cores, err := strconv.Atoi(os.Args[2])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_cores argument: %v\n", err)}
	n_pulse_port, err := strconv.Atoi(os.Args[3])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}
	local_muster_port := os.Args[4]
	n_ctrl_port, err := strconv.Atoi(os.Args[5])
	if err != nil {fmt.Printf("** ** ** ERROR: bad n_port argument: %v\n", err)}

	r_m := remote_muster{}
	r_m.init(n_ip, n_cores, n_pulse_port, n_ctrl_port, local_muster_port)

	go r_m.start_remote_pulser()
	r_m.start_remote_logger()
	go r_m.start_remote_controller()

	time.Sleep(time.Second*50)
}


