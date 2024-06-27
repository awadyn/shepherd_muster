package main

import (
	"fmt"
	"flag"
	"context"
//	"strconv"
	"time"
	"net"
//	"io"

	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

/*********************************************/

/* An ep-based remote muster watches over k sheep (i.e. cores) whereby 
   each sheep (i.e. core) can produce a list of logs and 
   each sheep (i.e. core) can be controlled by a list of controls
*/
func (r_m *remote_muster) init(n_ip string, n_cores int, pulse_server_port int, ctrl_server_port int, log_server_port string, coordinate_server_port string) {
	r_m.pulse_server_port = flag.Int("pulse_port", pulse_server_port, "remote_muster pulse serving port")
	r_m.ctrl_server_port = flag.Int("ctrl_port", ctrl_server_port, "remote_muster ctrl serving port")
	r_m.log_server_addr = flag.String("log_server_addr_" + r_m.id, 
					  mirror_ip + ":" + log_server_port, 
					  "address of mirror local_muster log sync server")
	r_m.coordinate_server_addr = flag.String("coordinate_server_addr", 
					         "localhost:" + coordinate_server_port, 
					         "address of shepherd's coordinate server")
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
	fmt.Printf("-- %v :  Log syncing server waiting.. \n", r_m.id)
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

///*********************/
///* REMOTE CONTROLLER */
///*********************/
//
//func (r_m *test_muster) check_ctrl_alive(sheep_id string, ctrls map[string]uint64) bool {
//	fmt.Println(r_m.done_log_map[sheep_id])
//	sheep := r_m.pasture[sheep_id]
//	c_str := strconv.Itoa(int(sheep.core))
//	ctrl_dvfs_id := "ctrl-dvfs-" + c_str + "-" + r_m.ip
//	ctrl_itr_id := "ctrl-itr-" + c_str + "-" + r_m.ip
//	dvfs_val := ctrls[ctrl_dvfs_id]
//	itr_val := ctrls[ctrl_itr_id]
//	dvfs_str := fmt.Sprintf("0x%x", dvfs_val)
//	itr_str := strconv.Itoa(int(itr_val))
//	log_fname := "/home/tanneen/shepherd_muster/mcd_logs/linux.mcd.dmesg.0_" + c_str + "_" + itr_str + "_" + dvfs_str + "_135_200000.ep.csv.ep"
//	select {
//	case <- r_m.done_log_map[sheep_id][log_fname]:
//		fmt.Println("!!!!!!!!!!! SKIPPING NEW CTRL - ", sheep_id)
//		select {
//		case r_m.done_log_map[sheep_id][log_fname] <- true:
//		default:
//		}
//		return true 
//	default:
//		return false
//	}
//}
//
//
//func (r_m *test_muster) handle_new_ctrl() {
//	for {
//		select {
//		case new_ctrl_req := <- r_m.new_ctrl_chan:
//			sheep := r_m.pasture[new_ctrl_req.sheep_id]
//			new_ctrls := new_ctrl_req.ctrls
//			// check if new ctrl is not applicable - i.e. when logging for that ctrl is completed
//			skip_ctrl := r_m.check_ctrl_alive(sheep.id, new_ctrls)
//			if skip_ctrl {
//				sheep.done_kill_chan <- false
//				<- sheep.ready_ctrl_chan
//			} else {
//				// kill logging of current ctrl
//				sheep.kill_log_chan <- true
//				// once shepherd is told that new ctrl is acknowledged and old ctrl logging is killed..
//				<- sheep.ready_ctrl_chan
//				// apply new ctrls 
//				for ctrl_id, ctrl_val := range(new_ctrls) {
//					sheep.controls[ctrl_id].value = ctrl_val
//					//sheep.controls[ctrl_id].dirty = true
//				}
//				// then restart logging
//				for log_id, _ := range(sheep.logs) {
//					go r_m.simulate_remote_log(sheep.id, log_id, sheep.core)
//				}
//			}
//		}
//	}
//}
//
//func (r_m *test_muster) ApplyControl(stream pb.Control_ApplyControlServer) error {
//	var sheep_id string
//	new_ctrls := make(map[string]uint64)
//	for {
//		req, err := stream.Recv()
//		switch {
//		case err == io.EOF:
//			fmt.Println("NEW CTRL - ", sheep_id, new_ctrls)
//
//			r_m.new_ctrl_chan <- ctrl_req{sheep_id: sheep_id, ctrls: new_ctrls}
//			done_kill := <- r_m.pasture[sheep_id].done_kill_chan
//			r_m.pasture[sheep_id].ready_ctrl_chan <- true
//
//			fmt.Printf("------------ COMPLETED CTRL-REQ -- %v\n", sheep_id)
//			return stream.SendAndClose(&pb.ControlReply{CtrlComplete: done_kill})
//		case err != nil:
//			fmt.Printf("** ** ** ERROR: could not receive control request: %v\n", err)
//			return err
//		default:
//			fmt.Printf("------------ CTRL-REQ -- %v\n", sheep_id)
//			sheep_id = req.GetSheepId()
//			ctrl_id := req.GetCtrlEntry().GetCtrlId()
//			ctrl_val := req.GetCtrlEntry().GetVal()
//			new_ctrls[ctrl_id] = ctrl_val
//		}
//	}
//}
//
//func (r_m *test_muster) start_controller() {
//	fmt.Printf("-- %v :  Ctrl syncing client waiting for heartbeats.. \n", r_m.id)
//	<- r_m.hb_chan
//	fmt.Printf("-- STARTING REMOTE MUSTER CONTROLLER :  %v\n", r_m.id)
//	flag.Parse()
//	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *r_m.ctrl_server_port))
//	if err != nil {
//		fmt.Printf("** ** ** ERROR: %v failed to listen on control port: %v\n", r_m.id, err)
//	}
//	s := grpc.NewServer()
//	pb.RegisterControlServer(s, r_m)
//	fmt.Printf("-- %v -- Control server listening at %v ... ... ...\n", r_m.id, lis.Addr())
//	if err := s.Serve(lis); err != nil {
//		fmt.Printf("** ** ** ERROR: failed to serve: %v\n", err)
//	}
//}




