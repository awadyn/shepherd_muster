package main

import (
	"fmt"
	"flag"
	"context"
	"strconv"
	"time"
	"net"
	"io"

	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*********************************************/

/*
  A remote muster watches over k sheep (i.e. cores) - 
  each sheep (i.e. core) can produce a list of logs and 
  each sheep (i.e. core) can be controlled by a list of controls
*/
func (r_m *remote_muster) init() { 
	r_m.hb_chan = make(chan bool)
	r_m.log_server_addr = flag.String("log_server_addr_" + r_m.id, 
					  mirror_ip + ":" + strconv.Itoa(r_m.log_port), 
					  "address of mirror local_muster log sync server")
	r_m.pulse_server_port = flag.Int("pulse_port_" + r_m.id, r_m.pulse_port, 
						"remote_muster pulse server port")
	r_m.ctrl_server_port = flag.Int("ctrl_port_" + r_m.id, r_m.ctrl_port, 
						"remote_muster ctrl server port")
	r_m.coordinate_server_port = flag.Int("coordinate_port_" + r_m.id, r_m.coordinate_port, 
						"remote muster coordinate server port")
}

/*****************/
/* REMOTE PULSER */
/*****************/
var hb_counter int = 0

func (r_m *remote_muster) HeartBeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatReply, error) {
	if hb_counter % 3 == 0 { fmt.Printf("-- HB REQ %v\n", in.GetShepRequest()) }
	select {
	case r_m.hb_chan <- true:
	default:
	}
	hb_counter ++
	return &pb.HeartbeatReply{MusterReply: r_m.id, ShepRequest: in.GetShepRequest()}, nil
}

func (r_m *remote_muster) start_pulser() {
	fmt.Printf("\033[35;1m-- STARTING PULSER :  %v\n\033[0m", r_m.id)
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.pulse_server_port))
	if err != nil {
		fmt.Printf("\033[31;1m****** ERROR: %v failed to listen: %v\n\033[0m", r_m.id, err)
	}
	s := grpc.NewServer()
	pb.RegisterPulseServer(s, r_m)
	fmt.Printf("\033[35m---- Heartbeat server listening at %v - %v\n\033[0m", r_m.ip, lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("\033[31;1m****** ERROR: failed to serve: %v\n\033[0m", err)
	}
}

/*****************/
/* REMOTE LOGGER */
/*****************/

func (r_m *remote_muster) start_logger() {
	fmt.Printf("\033[35;1m-- STARTING LOGGER :  %v\n\033[0m", r_m.id)
	conn, err := grpc.Dial(*r_m.log_server_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("\033[31;1m****** ERROR: %v could not create connection to local muster %s:\n****** %v\n\033[0m", r_m.id, *r_m.log_server_addr, err)
	}
	c := pb.NewLogClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), exp_timeout)
	fmt.Printf("\033[35m---- Initialized log client -- %v\n\033[0m", r_m.id)
	go r_m.log(conn, c, ctx, cancel)
}

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
	for {
		select {
		case ids := <- r_m.full_buff_chan:
			sheep_id := ids[0]
			log_id := ids[1]
			for {
				stream, err := c.SyncLogBuffers(ctx)
				if err != nil {
					fmt.Printf("\033[31;1m****** ERROR: %v could not initialize log sync stream %v:\n****** %v\n\033[0m", r_m.id, log_id, err)
					time.Sleep(time.Second/10)
					continue
				}
				fmt.Printf("\033[36m<----- SYNC REQ -- %v - %v\n\033[0m", sheep_id, log_id)
				for _, log_entry := range *(r_m.pasture[sheep_id].logs[log_id].mem_buff) {
					for {
						err := stream.Send(&pb.SyncLogRequest{SheepId: sheep_id, LogId:log_id, LogEntry: &pb.LogEntry{Vals: log_entry}})
						if err != nil { 
							fmt.Printf("\033[31;1m****** ERROR: %v %v could not send log entry %v:\n******%v\n\033[0m", r_m.id, log_id, log_entry, err)
							time.Sleep(time.Second/20)
							continue
						}
						break
					}
				}
				r, err := stream.CloseAndRecv()
				if err != nil {
					fmt.Printf("\033[31;1m****** ERROR: %v problem receiving log sync reply %v:\n****** %v\n\033[0m", r_m.id, log_id, err)
					time.Sleep(time.Second/10)
					continue
				}
				fmt.Printf("\033[36m-----> SYNC REP -- %v - %v\n\033[0m", log_id, r.GetSyncComplete())
				if r.GetSyncComplete() {
					r_m.pasture[sheep_id].logs[log_id].ready_buff_chan <- true
					break
				}
			}
		}
	}
}

/*********************/
/* REMOTE CONTROLLER */
/*********************/

func (r_m *remote_muster) ApplyControl(stream pb.Control_ApplyControlServer) error {
	ctrl_ctr := 0
	var sheep_id string
	new_ctrls := make(map[string]uint64)
	for {
		req, err := stream.Recv()
		switch {
		case err == io.EOF:
			fmt.Printf("\033[35m-----> CTRL-REQ -- %v - %v\n\033[0m", sheep_id, new_ctrls)
			r_m.new_ctrl_chan <- control_request{sheep_id: sheep_id, ctrls: new_ctrls}
			<- r_m.pasture[sheep_id].ready_ctrl_chan
			fmt.Printf("\033[35m<----- CTRL REP -- %v - %v\n\033[0m", sheep_id, new_ctrls)
			return stream.SendAndClose(&pb.ControlReply{CtrlComplete: true})
		case err != nil:
			fmt.Printf("\033[31;1m****** ERROR: could not receive control request: %v\n\033[0m", err)
			return err
		default:
			sheep_id = req.GetSheepId()
			ctrl_id := req.GetCtrlEntry().GetCtrlId()
			ctrl_val := req.GetCtrlEntry().GetVal()
			new_ctrls[ctrl_id] = ctrl_val
			ctrl_ctr ++
		}
	}
}

func (r_m *remote_muster) start_controller() {
	fmt.Printf("\033[35;1m-- STARTING CONTROLLER :  %v\n\033[0m", r_m.id)
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.ctrl_server_port))
	if err != nil {
		fmt.Printf("\033[31;1m****** ERROR: %v failed to listen on control port: %v\n\033[0m", r_m.id, err)
	}
	s := grpc.NewServer()
	pb.RegisterControlServer(s, r_m)
	fmt.Printf("\033[35m---- Control server listening at %v - %v\n\033[0m", r_m.ip, lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("\033[31;1m****** ERROR: failed to serve: %v\n\033[0m", err)
	}
}


/**********************/
/* REMOTE COORDINATOR */
/**********************/

/* CoordinateLog: one type of coordination service
	-> when a CoordinateLog request is received, a remote muster should
	   signal the native logging wrapper to sync native logs with the
	   local muster mirror
	   --> the protocol by which native logs are synced is defined by
	   the remote muster's specialization, e.g:
	       	-> a bayopt remote muster is specialized to sync
		   only 1 log entry per CoordinateLog request
		-> a reinforcement learning remote muster is specialized
		   to sync all log entries until signalled to stop doing so
*/
func (r_m *remote_muster) CoordinateLog(ctx context.Context, in *pb.CoordinateLogRequest) (*pb.CoordinateLogReply, error) {
	sheep_id := in.GetSheepId()
	log_id := in.GetLogId()
	coordinate_cmd := in.GetCoordinateCmd()
	fmt.Printf("\033[34m---> COORD REQ %v -- %v -- %v -- %v\n\033[0m", r_m.id, sheep_id, log_id, coordinate_cmd)
	r_m.pasture[sheep_id].request_log_chan <- []string{log_id, coordinate_cmd}
	cmd_status := <- r_m.pasture[sheep_id].logs[log_id].ready_request_chan
	//fmt.Printf("\033[34m----DONE COORD REQ %v -- %v -- %v -- %v\n\033[0m", r_m.id, sheep_id, log_id, coordinate_cmd)
	return &pb.CoordinateLogReply{SheepId: sheep_id, LogId: log_id, Status: cmd_status, CoordinateCmd: coordinate_cmd}, nil
}

func (r_m *remote_muster) CoordinateCtrl(ctx context.Context, in *pb.CoordinateCtrlRequest) (*pb.CoordinateCtrlReply, error) {
	sheep_id := in.GetSheepId()
	ctrl_id := in.GetCtrlId()
	fmt.Printf("\033[34m---> COORD REQ %v -- %v -- %v\n\033[0m", r_m.id, sheep_id, ctrl_id)
	ctrl_val := r_m.pasture[sheep_id].controls[ctrl_id].value
//	r_m.pasture[sheep_id].request_ctrl_chan <- ctrl_id
//	<- r_m.pasture[sheep_id].controls[ctrl_id].ready_request_chan
	////fmt.Printf("\033[34m----DONE COORD REQ %v -- %v -- %v -- %v\n\033[0m", r_m.id, sheep_id, ctrl_id, ctrl_val)
	return &pb.CoordinateCtrlReply{SheepId: sheep_id, CtrlId: ctrl_id, CtrlVal: ctrl_val}, nil
}

func (r_m *remote_muster) start_coordinator() {
	fmt.Printf("\033[35;1m-- STARTING COORDINATOR :  %v\n\033[0m", r_m.id)
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.coordinate_server_port))
	if err != nil {
		fmt.Printf("\033[31;1m****** ERROR: %v failed to listen on control port: %v\n\033[0m", r_m.id, err)
	}
	s := grpc.NewServer()
	pb.RegisterCoordinateServer(s, r_m)
	fmt.Printf("\033[35m---- Coordination server listening at %v - %v\n\033[0m", r_m.ip, lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("\033[31;1m****** ERROR: failed to serve: %v\n\033[0m", err)
	}
}

//func (r_m *remote_muster) wait_done() {
//	total_ctr := 0
//	for sheep_id, _ := range(r_m.pasture) {
//		for i := 0; i < len(r_m.pasture[sheep_id].logs) ; i++ { total_ctr ++ }
//	}
//	done_ctr := 0
//	for {
//		select {
//		case ids := <- r_m.done_chan:
//			sheep_id := ids[0]
//			log_id := ids[1]
//			done_ctr ++
//			fmt.Println("********** DONE ********** ", sheep_id, log_id)
//			conn, err := grpc.Dial(*r_m.coordinate_server_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
//			if err != nil {
//				fmt.Printf("** ** ** ERROR: %v could not create connection to shepherd %s:\n** ** ** %v\n", r_m.id, *r_m.coordinate_server_addr, err)
//			}
//			c := pb.NewCoordinateClient(conn)
//			ctx, cancel := context.WithTimeout(context.Background(), exp_timeout)
//			fmt.Printf("-- Initialized coordinate client for %v\n", sheep_id)
//			r, err := c.CompleteRun(ctx, &pb.CompleteRunRequest{MusterId: r_m.id, SheepId: sheep_id})
//			fmt.Printf("-- Coordination complete:  %v\n", r)
//			conn.Close()
//			cancel()
//		}
//		if done_ctr == total_ctr { 
//			r_m.exit_chan <- true
//			return
//		}
//	}
//}

