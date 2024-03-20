package main

import (
	"fmt"
	"flag"
	"net"
	"context"
	"os"
	"strconv"
	"time"
	"io"

	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*********************************************/
func (m *muster) init(n_ip string, n_cores int, n_pulse_port int, n_ctrl_port int, local_muster_port string) {
	n := node{ip:n_ip, ncores: uint8(n_cores)}
	m_id := "muster-" + n.ip
	m.id = m_id
	m.node = n 
	m.pasture = make(map[string]*sheep)
	m.hb_chan = make(chan bool)
	m.full_buff_chan = make(chan []string)

	var core uint8
	for core = 0; core < n.ncores; core ++ {
		c_str := strconv.Itoa(int(core))
		sheep_id := c_str + "-" + m.ip
		sheep_c := sheep{id: sheep_id, core: core,
				 logs: make(map[string]*log),
				 controls: make(map[string]*control)}
		m.pasture[sheep_id] = &sheep_c
	}
}

/* An ep-based remote muster watches over k sheep (i.e. cores) whereby 
   each sheep (i.e. core) can produce a list of logs and 
   each sheep (i.e. core) can be controlled by a list of controls
*/
func (r_m *remote_muster) init(n_ip string, n_cores int, n_pulse_port int, n_ctrl_port int, local_muster_port string, coordinate_port string) {
	r_m.pulse_port = flag.Int("pulse_port", n_pulse_port, "remote_muster_pulse_port")
	r_m.ctrl_port = flag.Int("ctrl_port", n_ctrl_port, "remote_muster_ctrl_port")
	r_m.local_muster_addr = flag.String("local_muster_addr_" + r_m.id, 
					    "localhost:" + local_muster_port, 
					    "address of mirror local_muster log sync server")
	r_m.coordinate_addr = flag.String("coordinate_addr", 
					    "localhost:" + coordinate_port, 
					    "address of shepherd coordination server")
	var core uint8
	for core = 0; core < r_m.ncores; core ++ {
		var max_size uint64 = 4096 * 8
		mem_buff := make([][]uint64, max_size)
		c_str := strconv.Itoa(int(core))
		sheep_id := c_str + "-" + r_m.ip
		log_id := "log-" + c_str + "-" + r_m.ip
		r_logger := remote_logger{stop_log_chan: make(chan bool, 1),
			     		  done_log_chan: make(chan bool, 1)}
		log_c := log{remote_logger: r_logger, id: log_id,
			     metrics: []string{"timestamp", "joules"},
			     max_size: max_size,
			     r_buff: &mem_buff,
			     ready_buff_chan: make(chan bool, 1)}
		ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + r_m.ip
		ctrl_itr_id := "ctrl-itr-" + c_str + "-" + r_m.ip
		ctrl_dvfs := control{id: ctrl_dvfs_id, n_ip: r_m.ip, //core: core, 
				     knob: "dvfs", value: 0x1100, dirty: false} 
		ctrl_itr := control{id: ctrl_itr_id, n_ip: r_m.ip, //core: core, 
				    knob: "itr-delay", value: 100, dirty: false}  
		r_m.pasture[sheep_id].logs[log_id] = &log_c
		r_m.pasture[sheep_id].controls[ctrl_dvfs_id] = &ctrl_dvfs
		r_m.pasture[sheep_id].controls[ctrl_itr_id] = &ctrl_itr
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
	for sheep_id, _ := range(r_m.pasture) {
		core := r_m.pasture[sheep_id].core
		for log_id, _ := range(r_m.pasture[sheep_id].logs) { 
			go r_m.simulate_remote_log(sheep_id, log_id, core) 
		}
	}
	for {
		select {
		case ids := <- r_m.full_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			fmt.Printf("-------- FULL_BUFF :  %v - %v \n", sheep_id, log_id)
			for {
				stream, err := c.SyncLogBuffers(ctx)
				if err != nil {
					fmt.Printf("** ** ** ERROR: %v could not call sync log server for %v:\n** ** ** %v\n", r_m.id, log_id, err)
					time.Sleep(time.Second/10)
					continue
				}
				for _, log_entry := range *(r_m.pasture[sheep_id].logs[log_id].r_buff) {
					for {
						err := stream.Send(&pb.SyncLogRequest{SheepId: sheep_id, LogId:log_id, LogEntry: &pb.LogEntry{Vals: log_entry}})
						if err != nil { 
							fmt.Printf("** ** ** ERROR: %v %v could not send log entry %v:\n** ** **%v\n", r_m.id, log_id, log_entry, err)
							time.Sleep(time.Second/20)
							continue
						}
						break
					}
				}
				r, err := stream.CloseAndRecv()
				if err != nil {
					fmt.Printf("** ** ** ERROR: %v problem receiving sync log reply %v:\n** ** ** %v\n", r_m.id, log_id, err)
					time.Sleep(time.Second/10)
					continue
				}
				fmt.Printf("-------- SYNC LOG REP : %v - %v\n", log_id, r.GetSyncComplete())
				if r.GetSyncComplete() {
					r_m.pasture[sheep_id].logs[log_id].ready_buff_chan <- true
					break
				}
			}
		}
	}
}

func (r_m *remote_muster) start_logger() {
	<- r_m.hb_chan
	fmt.Printf("-- STARTING REMOTE MUSTER LOGGER :  %v\n", r_m.id)

	conn, err := grpc.Dial(*r_m.local_muster_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v could not create connection to local muster %s:\n** ** ** %v\n", r_m.id, *r_m.local_muster_addr, err)
	}
	c := pb.NewLogClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	fmt.Printf("-- %v -- Initialized log client \n", r_m.id)
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

func (r_m *remote_muster) start_pulser() {
	fmt.Printf("-- STARTING REMOTE MUSTER PULSER :  %v\n", r_m.id)

	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.pulse_port))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to listen: %v\n", r_m.id, err)
	}
	s := grpc.NewServer()
	pb.RegisterPulseServer(s, r_m)
	fmt.Printf("-- %v -- Heartbeat server listening at %v ... ... ...\n", r_m.id, lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("** ** ** ERROR: failed to serve: %v\n", err)
	}
}

/*********************/
/* REMOTE CONTROLLER */
/*********************/

func (r_m *remote_muster) ApplyControl(stream pb.Control_ApplyControlServer) error {
	var sheep_id string
	for {
		// TODO 1) stop logging, 2) change control settings, 3) restart logging
		ctrl_req, err := stream.Recv()
		switch {
		case err == io.EOF:
			fmt.Printf("------------ COMPLETED CTRL-REQ -- %v\n", sheep_id)
//			for log_id, _ := range(r_m.pasture[sheep_id].logs) {
//				r_m.pasture[sheep_id].logs[log_id].stop_log_chan <- true
//			}
			return stream.SendAndClose(&pb.ControlReply{CtrlComplete: true})
		case err != nil:
			fmt.Printf("** ** ** ERROR: could not receive control request: %v\n", err)
			return err
		default:
			fmt.Printf("------------ CTRL-REQ -- ")
			// change remote muster's control settings of the calling sheep (i.e. core)
			sheep_id = ctrl_req.GetSheepId()
			ctrl_id := ctrl_req.GetCtrlEntry().GetCtrlId()
			ctrl_val := ctrl_req.GetCtrlEntry().GetVal()
			fmt.Printf("%v -- %v -- %v \n", sheep_id, ctrl_id, ctrl_val)
			r_m.pasture[sheep_id].controls[ctrl_id].value = ctrl_val
			r_m.pasture[sheep_id].controls[ctrl_id].dirty = true
		}
	}
	return stream.SendAndClose(&pb.ControlReply{CtrlComplete: true})
}

func (r_m *remote_muster) start_controller() {
	<- r_m.hb_chan
	fmt.Printf("-- STARTING REMOTE MUSTER CONTROLLER :  %v\n", r_m.id)
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.ctrl_port))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to listen on control port: %v\n", r_m.id, err)
	}
	s := grpc.NewServer()
	pb.RegisterControlServer(s, r_m)
	fmt.Printf("-- %v -- Control server listening at %v ... ... ...\n", r_m.id, lis.Addr())
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
	coordinate_port := os.Args[6]

	m := muster{}
	m.init(n_ip, n_cores, n_pulse_port, n_ctrl_port, local_muster_port)
	r_m := remote_muster{muster: m}
	r_m.init(n_ip, n_cores, n_pulse_port, n_ctrl_port, local_muster_port, coordinate_port)
	r_m.show()

	go r_m.start_pulser()
	r_m.start_logger()
	go r_m.start_controller()

	// confirm that all muster loggers are done
	for _, sheep := range(r_m.pasture) {
		for _, log := range(sheep.logs) {
			<- log.done_log_chan
		}
		conn, err := grpc.Dial(*r_m.coordinate_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			fmt.Printf("** ** ** ERROR: %v could not create connection to shepherd %s:\n** ** ** %v\n", r_m.id, *r_m.coordinate_addr, err)
		}
		c := pb.NewCoordinateClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		fmt.Printf("-- Initialized coordinate client for %v\n", sheep.id)
		r, err := c.CompleteRun(ctx, &pb.CompleteRunRequest{MusterId: r_m.id, SheepId: sheep.id})
		fmt.Printf("-- Coordination complete:  %v\n", r)
		conn.Close()
		cancel()
	}

	// confirm that all muster controllers are done
	// TODO

	// TODO remove sleep: add rpc to shepherd to confirm no buffers are being copied, then kill remote muster
//	time.Sleep(time.Second)
}


