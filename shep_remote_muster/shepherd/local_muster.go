package main

import (
	"fmt"
	"context"
	"flag"
	"net"
	"io"
	"time"
	"strconv"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/************************************/

func (l_m *local_muster) init() {
	l_m.hb_chan = make(chan *pb.HeartbeatReply)
	var idx string = ""
	if l_m.ip_idx != -1 { idx = strconv.Itoa(int(l_m.ip_idx)) } 
	l_m.log_server_port = flag.Int("log_server_port_" + l_m.id + idx, l_m.log_sync_port, 
					"local muster log syncing server port")
	l_m.pulse_server_addr = flag.String("pulse_server_addr_" + l_m.id + idx, l_m.ip + ":" + strconv.Itoa(l_m.pulse_port),
						"address of one remote muster pulse server")
	l_m.ctrl_server_addr = flag.String("ctrl_server_addr_" + l_m.id + idx, l_m.ip + ":" + strconv.Itoa(l_m.ctrl_port),
						"address of one remote muster control server")
	l_m.coordinate_server_addr = flag.String("coordinate_server_addr_" + l_m.id + idx, l_m.ip + ":" + strconv.Itoa(l_m.coordinate_port),
							"address of remote muster  coordination server")
}


/************************/
/***** LOCAL PULSER *****/
/************************/

/* 
  local musters represent the client side of pulsing: where the server side
  is a remote mirror and a local pulse client checks responsiveness of 
  a remote pulse server. 
*/
func (l_m *local_muster) start_pulser() {
	fmt.Printf("\033[34;1m-- STARTING LOCAL PULSER :  %v\n\033[0m", l_m.id)
	conn, err := grpc.Dial(*l_m.pulse_server_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("****** ERROR: could not create local connection to remote muster %s:\n****** %v\n", l_m.id, err)
	}
	c := pb.NewPulseClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), exp_timeout)
	fmt.Printf("\033[36m---- %v -- Initialized pulse client\n\033[0m", l_m.id)
	go l_m.pulse(conn, c, ctx, cancel)
}

func (l_m *local_muster) pulse(conn *grpc.ClientConn, c pb.PulseClient, ctx context.Context, cancel context.CancelFunc) {
	defer conn.Close()
	defer cancel()
	var counter uint32 = 0
	err_count := 0
	for {
		counter += 1
		r, err := c.HeartBeat(ctx, &pb.HeartbeatRequest{ShepRequest: counter})  
		if err != nil {
			err_count ++
			if err_count == 30 {
				fmt.Printf("\033[31;1m***** LOST PULSE:  %v\n\033[0m", l_m.id)
				return
			}
		} else { 
			err_count = 0 
			l_m.hb_chan <- r
		}
		time.Sleep(time.Second/5)
	}
}


/************************/
/***** LOCAL LOGGER *****/
/************************/

/* These functions start a local muster thread that serves
   log sync requests from its mirror remote muster. 
   A remote muster sends a log sync request whenever a memory buffer
   of one of its sheeps' logs is full. 
   A remote muster will only continue logging into said memory buffer
   after log syncing with the local muster server is complete.
*/

func (l_m *local_muster) start_logger() {
	fmt.Printf("\033[34;1m-- STARTING LOCAL LOGGER :  %v\n\033[0m", l_m.id)
	go l_m.log()
}

func (l_m *local_muster) log() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *l_m.log_server_port))
	if err != nil {
		fmt.Printf("\033[31;1m****** ERROR: %v failed to listen at %v: %v\n\033[0m", l_m.id, *l_m.log_server_port, err)
	}
	s := grpc.NewServer()
	pb.RegisterLogServer(s, l_m)
	fmt.Printf("\033[36m---- %v -- Initialized log sync server listening at %v \n\033[0m", l_m.id, lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("\033[31;1m****** ERROR: %v failed to start log sync server: %v\n\033[0m", l_m.id, err)
	}
}

func (l_m *local_muster) SyncLogBuffers(stream pb.Log_SyncLogBuffersServer) error {
	buff_ctr := 0
	var log_id string
	var sheep_id string
	for {
		sync_req, err := stream.Recv()
		switch {
		case err == io.EOF:
			/* i.e. all log entries have been copied to mem_buff*/
			l_m.full_buff_chan <- []string{sheep_id, log_id}
			<- l_m.pasture[sheep_id].logs[log_id].ready_buff_chan
			fmt.Printf("\033[36m<----- SYNC-REP -- %v - %v - %v\n\033[0m", l_m.id, sheep_id, log_id)
			return stream.SendAndClose(&pb.SyncLogReply{SyncComplete:true})
		case err != nil:
			fmt.Printf("\033[31;1m****** ERROR: could not receive sync log request from stream\n\033[0m")
			return err
		default:
			sheep_id = sync_req.GetSheepId()
			log_id = sync_req.GetLogId()
			mem_buff := l_m.pasture[sheep_id].logs[log_id].mem_buff
			if buff_ctr == 0 { 
				fmt.Printf("\033[36m-----> SYNC-REQ -- %v - %v - %v\n\033[0m", l_m.id, sheep_id, log_id) 
				*(l_m.pasture[sheep_id].logs[log_id].mem_buff)  = make([][]uint64, 0)
			}	
			*mem_buff = append(*mem_buff, sync_req.GetLogEntry().GetVals())
			buff_ctr++
		}
	}
}


/************************/
/*** LOCAL CONTROLLER ***/
/************************/

/* 
  local musters represent the client side of ctrl: where the server side
  is a remote mirror and a local ctrl client sends new ctrl requests to 
  a remote pulse server. 
*/
func (l_m *local_muster) start_controller() {
	fmt.Printf("\033[34;1m-- STARTING LOCAL CONTROLLER :  %v\n\033[0m", l_m.id)
	conn, err := grpc.Dial(*l_m.ctrl_server_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("\033[31;1m****** ERROR: could not create local connection to remote controller %s:\n****** %v\n\033[0m", l_m.id, err)
		panic(err)
	}
	c := pb.NewControlClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), exp_timeout)
	fmt.Printf("\033[36m---- %v -- Initialized control client \n\033[0m", l_m.id)
	go l_m.control(conn, c, ctx, cancel)
}

func (l_m *local_muster) control(conn *grpc.ClientConn, c pb.ControlClient, ctx context.Context, cancel context.CancelFunc) {
	<- l_m.hb_chan
	defer conn.Close()
	defer cancel()
	var done_ctrl bool
	for {
		select {
		case new_ctrl_req := <- l_m.new_ctrl_chan:
			go func() {
				new_ctrl_req := new_ctrl_req
				sheep_id := new_ctrl_req.sheep_id
				new_ctrls := new_ctrl_req.ctrls
				for {
					stream, err := c.ApplyControl(ctx)
					if err != nil { time.Sleep(time.Second/5); continue }
					for ctrl_id, ctrl_val := range(new_ctrls) {
						err = stream.Send(&pb.ControlRequest{SheepId: sheep_id, CtrlEntry: &pb.ControlEntry{CtrlId: ctrl_id, Val: ctrl_val}})
						if err != nil { time.Sleep(time.Second/5); continue }
					}
					r, err := stream.CloseAndRecv()
					done_ctrl = r.GetCtrlComplete()
					if !done_ctrl { fmt.Printf("\033[31;1m****** CTRL WAS NOT APPLIED - %v - %v\n\033[0m", l_m.id, sheep_id) }
					if err != nil { time.Sleep(time.Second/5); continue }
					break
				}
				new_ctrl_reply := control_reply{ctrls: new_ctrls, done: done_ctrl}
				l_m.pasture[sheep_id].done_ctrl_chan <- new_ctrl_reply
				fmt.Printf("\033[35m-------> CTRL REP --  %v - %v - %v\n\033[0m", l_m.id, sheep_id, new_ctrls)
			}()
		}
	}
}

/*************************/
/*** LOCAL COORDINATOR ***/
/*************************/
func (l_m *local_muster) start_coordinator() {
	fmt.Printf("\033[34;1m-- STARTING LOCAL COORDINATOR :  %v\n\033[0m", l_m.id)
	conn, err := grpc.Dial(*l_m.coordinate_server_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("\033[31;1m****** ERROR: could not create local connection to remote muster %s:\n****** %v\n \033[0m", l_m.id, err)
	}
	c := pb.NewCoordinateClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), exp_timeout)
	fmt.Printf("\033[36m---- %v -- Initialized coordinate client \n\033[0m", l_m.id)
	go l_m.coordinate(conn, c, ctx, cancel)
}

func (l_m *local_muster) coordinate(conn *grpc.ClientConn, c pb.CoordinateClient, ctx context.Context, cancel context.CancelFunc) {
	<- l_m.hb_chan
	defer conn.Close()
	defer cancel()
	for {
		select {
		case req := <- l_m.request_log_chan:
			go func() {
				req := req
				sheep_id := req[0]
				log_id := req[1]
				coordinate_cmd := req[2]

				<- l_m.pasture[sheep_id].logs[log_id].ready_request_chan
				//fmt.Printf("\033[34m<--- COORD REQ -- %v -- %v -- %v\n\033[0m", sheep_id, log_id, coordinate_cmd)
				for {
					r, err := c.CoordinateLog(ctx, &pb.CoordinateLogRequest{SheepId: sheep_id, LogId: log_id, CoordinateCmd: coordinate_cmd})  
					if err != nil { time.Sleep(time.Second/2); continue } 
					fmt.Printf("\033[34m---> COORD LOG REP -- %v\n\033[0m", r)
					l_m.pasture[sheep_id].logs[log_id].ready_request_chan <- true
					return
				}
			} ()
		case req := <- l_m.request_ctrl_chan:
			go func() {
				req := req
				sheep_id := req[0]
				ctrl_id := req[1]
//				<- l_m.pasture[sheep_id].controls[ctrl_id].ready_request_chan
				//fmt.Printf("\033[34m<--- COORD REQ -- %v -- %v\n\033[0m", sheep_id, ctrl_id)
				for {
					r, err := c.CoordinateCtrl(ctx, &pb.CoordinateCtrlRequest{SheepId: sheep_id, CtrlId: ctrl_id})  
					if err != nil { time.Sleep(time.Second/2); continue } 
					ctrl_id_ret := r.GetCtrlId()
					if ctrl_id_ret != ctrl_id { panic(errors.New("RPC GONE WRONG!!")) }
					ctrl_val := r.GetCtrlVal()
					l_m.pasture[sheep_id].controls[ctrl_id].value = ctrl_val
					fmt.Printf("\033[34m---> COORD CTRL REP -- %v\n\033[0m", r)
					l_m.pasture[sheep_id].controls[ctrl_id].ready_request_chan <- true
					return
				}

			} ()
		}
	}
}




