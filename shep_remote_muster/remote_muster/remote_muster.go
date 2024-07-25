package main

import (
	"fmt"
	"flag"
	"context"
//	"strconv"
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
func (r_m *remote_muster) init(n_ip string, n_cores int, pulse_server_port int, ctrl_server_port int, 
				log_server_port string, coordinate_server_port int) {
	r_m.log_server_addr = flag.String("log_server_addr_" + r_m.id, 
					  mirror_ip + ":" + log_server_port, 
					  "address of mirror local_muster log sync server")
	r_m.pulse_server_port = flag.Int("pulse_port", pulse_server_port, 
						"remote_muster pulse server port")
	r_m.ctrl_server_port = flag.Int("ctrl_port", ctrl_server_port, 
						"remote_muster ctrl server port")
	r_m.coordinate_server_port = flag.Int("coordinate_port", coordinate_server_port, 
						"remote muster coordinate server port")
}


/*****************/
/* REMOTE LOGGER */
/*****************/

func (r_m *remote_muster) start_logger() {
	<- r_m.hb_chan

	fmt.Printf("-- STARTING REMOTE MUSTER LOGGER :  %v\n", r_m.id)
	conn, err := grpc.Dial(*r_m.log_server_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v could not create connection to local muster %s:\n** ** ** %v\n", r_m.id, *r_m.log_server_addr, err)
	}
	c := pb.NewLogClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), exp_timeout)
	fmt.Printf("-- %v -- Initialized log client \n", r_m.id)
	r_m.log(conn, c, ctx, cancel)
}

/* This function represents a single remote_muster logger thread
   which is a client of the shepherd server, which in turn performs
   all log processing and control computation relevant to this 
   remote_muster as well as all other remote_musters the shepherd
   is in charge of. 
*/
func (r_m *remote_muster) log(conn *grpc.ClientConn, c pb.LogClient, ctx context.Context, cancel context.CancelFunc) {
	fmt.Printf("-- %v :  Log syncing client waiting.. \n", r_m.id)
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
					fmt.Printf("** ** ** ERROR: %v could not call log sync server for %v:\n** ** ** %v\n", r_m.id, log_id, err)
					time.Sleep(time.Second/10)
					continue
				}
				for _, log_entry := range *(r_m.pasture[sheep_id].logs[log_id].mem_buff) {
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
					fmt.Printf("** ** ** ERROR: %v problem receiving log sync reply %v:\n** ** ** %v\n", r_m.id, log_id, err)
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

/*****************/
/* REMOTE PULSER */
/*****************/
var hb_counter int = 0

func (r_m *remote_muster) HeartBeat(ctx context.Context, in *pb.HeartbeatRequest) (*pb.HeartbeatReply, error) {
	if hb_counter % 3 == 0 { fmt.Printf("------HB-REQ-- %v\n", in.GetShepRequest()) }
	select {
	case r_m.hb_chan <- true:
	default:
	}
	hb_counter ++
	return &pb.HeartbeatReply{MusterReply: r_m.id, ShepRequest: in.GetShepRequest()}, nil
}

func (r_m *remote_muster) start_pulser() {
	fmt.Printf("-- STARTING REMOTE MUSTER PULSER :  %v\n", r_m.id)
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.pulse_server_port))
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
	ctrl_ctr := 0
	var sheep_id string
	new_ctrls := make(map[string]uint64)
	for {
		req, err := stream.Recv()
		switch {
		case err == io.EOF:
			r_m.new_ctrl_chan <- ctrl_req{sheep_id: sheep_id, ctrls: new_ctrls}
			<- r_m.pasture[sheep_id].ready_ctrl_chan
			fmt.Printf("------------ COMPLETED CTRL-REQ -- %v - %v\n", sheep_id, new_ctrls)
			return stream.SendAndClose(&pb.ControlReply{CtrlComplete: true})
		case err != nil:
			fmt.Printf("** ** ** ERROR: could not receive control request: %v\n", err)
			return err
		default:
			sheep_id = req.GetSheepId()
			if ctrl_ctr == 0 { 
				fmt.Printf("------------ CTRL-REQ -- %v\n", sheep_id)
			}
			ctrl_id := req.GetCtrlEntry().GetCtrlId()
			ctrl_val := req.GetCtrlEntry().GetVal()
			new_ctrls[ctrl_id] = ctrl_val
			ctrl_ctr ++
		}
	}
}

func (r_m *remote_muster) start_controller() {
	<- r_m.hb_chan

	fmt.Printf("-- STARTING REMOTE MUSTER CONTROLLER :  %v\n", r_m.id)
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.ctrl_server_port))
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
	fmt.Printf("------COORDINATE-REQ %v ------ %v -- %v -- %v\n", r_m.id, sheep_id, log_id, coordinate_cmd)

	r_m.pasture[sheep_id].logs[log_id].request_log_chan <- coordinate_cmd
//	r_m.pasture[sheep_id].logs[log_id].do_log_chan <- true
	<- r_m.pasture[sheep_id].logs[log_id].done_log_chan

	fmt.Printf("------DONE-COORDINATE-REQ %v ------ %v -- %v -- %v\n", r_m.id, sheep_id, log_id, coordinate_cmd)
	return &pb.CoordinateLogReply{SheepId: sheep_id, LogId: log_id, CoordinateCmd: coordinate_cmd}, nil
}

func (r_m *remote_muster) start_coordinator() {
	<- r_m.hb_chan

	fmt.Printf("-- STARTING REMOTE MUSTER COORDINATOR :  %v\n", r_m.id)
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.coordinate_server_port))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to listen on control port: %v\n", r_m.id, err)
	}
	s := grpc.NewServer()
	pb.RegisterCoordinateServer(s, r_m)
	fmt.Printf("-- %v -- Coordination server listening at %v ... ... ...\n", r_m.id, lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("** ** ** ERROR: failed to serve: %v\n", err)
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

