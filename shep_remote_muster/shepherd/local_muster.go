package main

import (
	"fmt"
	"context"
	"flag"
	"net"
	"io"
	"time"

	"google.golang.org/grpc"
	pb "github.com/awadyn/shep_remote_muster/shep_remote_muster"
)

/************************************/

/************************/
/***** LOCAL PULSER *****/
/************************/

func (l_m *local_muster) pulse(conn *grpc.ClientConn, c pb.PulseClient, ctx context.Context, cancel context.CancelFunc) {
	defer conn.Close()
	defer cancel()
	var counter uint32 = 0
	for {
		counter += 1
		r, err := c.HeartBeat(ctx, &pb.HeartbeatRequest{ShepRequest: counter})  
		if err != nil {
//			fmt.Printf("** ** ** ERROR: %v could not send heartbeat request to remote:\n** ** ** %v\n", l_m.id, err)
			time.Sleep(time.Second/5)
			continue
		}
		l_m.hb_chan <- r
		time.Sleep(time.Second/5)
	}
}

/************************/
/***** LOCAL LOGGER *****/
/************************/

/* This function starts a local muster thread that serves
   log sync requests from its mirror remote muster. A remote
   muster sends a log sync request whenever one of its log
   memory buffers are full. A remote muster can only continue
   logging into its memory buffers after log syncing with 
   the local muster is complete.
*/
func (l_m *local_muster) log() {
	// begin logging protocol once heartbeats are established
	<- l_m.hb_chan
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *l_m.log_sync_port))
	if err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to listen at %v: %v\n", l_m.id, *l_m.log_sync_port, err)
	}
	s := grpc.NewServer()
	pb.RegisterLogServer(s, l_m)
	fmt.Printf("-- %v -- Log sync server listening at %v ... ... ...\n", l_m.id, lis.Addr())
	if err := s.Serve(lis); err != nil {
		fmt.Printf("** ** ** ERROR: %v failed to start log sync server: %v\n", l_m.id, err)
	}
}

func (l_m *local_muster) SyncLogBuffers(stream pb.Log_SyncLogBuffersServer) error {
	l_buff_ctr := 0
	var log_id string
	var sheep_id string
	for {
		log_sync_req, err := stream.Recv()
		switch {
		case err == io.EOF:
			fmt.Printf("-------- COMPLETED-SYNC-REQ -- %v - %v - %v\n", l_m.id, sheep_id, log_id)
			// signal shepherd to start processing synced log
			l_m.process_buff_chan <- []string{sheep_id, log_id}
			return stream.SendAndClose(&pb.SyncLogReply{SyncComplete:true})
		case err != nil:
			fmt.Printf("** ** ** ERROR: could not receive sync log request from stream\n")
			return err
		default:
			sheep_id = log_sync_req.GetSheepId()
			log_id = log_sync_req.GetLogId()
			mem_buff := l_m.pasture[sheep_id].logs[log_id].l_buff
			if l_buff_ctr == 0 { 
				fmt.Printf("-------- SYNC-REQ -- %v - %v\n", sheep_id, log_id) 
				// local log buffer requested for sync should not be in-use by shepherd's log processing thread
				<- l_m.pasture[sheep_id].logs[log_id].ready_buff_chan
			}	
			// copy sync log data into local memory buffer 
			(*mem_buff)[l_buff_ctr] = log_sync_req.GetLogEntry().GetVals()
			l_buff_ctr++
		}
	}
}

/************************/
/*** LOCAL CONTROLLER ***/
/************************/

func (l_m *local_muster) control(conn *grpc.ClientConn, c pb.ControlClient, ctx context.Context, cancel context.CancelFunc) {
	// begin control protocol once heartbeats are established
	<- l_m.hb_chan
	defer conn.Close()
	defer cancel()
	for {
		select {
		case sheep_id := <- l_m.ready_ctrl_chan:
			for {
				stream, err := c.ApplyControl(ctx)
				if err != nil { 
					fmt.Printf("** ** ** ERROR: %v could not send control request for %v:\n** ** ** %v\n", l_m.id, sheep_id, err)
					time.Sleep(time.Second/20)
					continue 
				}
				for _, ctrl := range(l_m.pasture[sheep_id].controls) {
					if !ctrl.dirty { continue }
					err = stream.Send(&pb.ControlRequest{SheepId: sheep_id, CtrlEntry: &pb.ControlEntry{CtrlId: ctrl.id, Val: ctrl.value}})
					if err != nil { 
						fmt.Printf("** ** ** ERROR: %v %v could not send dvfs control entry:\n** ** **%v\n", l_m.id, sheep_id, err)
						continue 
					}
				}
				_, err = stream.CloseAndRecv()
				if err != nil { 
					fmt.Printf("** ** ** ERROR: %v problem receiving control reply %v:\n** ** ** %v\n", l_m.id, sheep_id, err)
					time.Sleep(time.Second/20)
					continue
				}
				// reset dirty bit of applied controls
				for _, ctrl := range(l_m.pasture[sheep_id].controls) {
					if ctrl.dirty { ctrl.dirty = false }
				}
				break
			}
		}
	}
}


