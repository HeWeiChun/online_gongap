package flowMap

import (
	"container/list"
	"github.com/Freddy/sctp_flowmap/database/PacketDB"
	"github.com/Freddy/sctp_flowmap/database/TimeFlow"
	"github.com/Freddy/sctp_flowmap/database/UEFlow"
	"strconv"
	"fmt"
	"time"
	"errors"
)
type semaphore struct {
	c         chan int8
	initCount int
	maxCount  int
}

func NewSemaphore(initCount int, maxCount int) (*semaphore, error) {
	if maxCount <= 0 || initCount < 0 || initCount > maxCount {
		return nil, errors.New("参数错误")
	}
	s := &semaphore{
		c: make(chan int8, maxCount),
	}
	for i := 1; i <= initCount; i++ {
		s.c <- 1
	}
	return s, nil
}

// P P操作
func (s *semaphore) P() {
	_ = <-s.c
}

// V V操作
func (s *semaphore) V() {
	s.c <- 1
}

var (
	FlowTable_UE   []*Flow
	FlowTable_Time []*Flow
	UE_SEM         []*semaphore
	TIME_SEM       []*semaphore
	flowCount_UE   uint64 = 0
	flowCount_Time uint64 = 0

	TimeMax        int64  = 300000000
	min_intervl    int64  = 1
	TimeChip       int64  = 5000000000
)

func init() {
	FlowTable_UE = make([]*Flow, TABLE_SIZE)
	FlowTable_Time = make([]*Flow, TABLE_SIZE)
	TIME_SEM = make([]*semaphore ,TABLE_SIZE)
	UE_SEM = make([]*semaphore, TABLE_SIZE)
	for i := 0; i< int(TABLE_SIZE); i ++{
		var err error
		TIME_SEM[i],err = NewSemaphore(1, 1)
		for err!=nil {
			fmt.Println(err)
			TIME_SEM[i],err = NewSemaphore(1, 1)
		}
		UE_SEM[i],err = NewSemaphore(1, 1)
		for err!=nil {
			fmt.Println(err)
			UE_SEM[i],err = NewSemaphore(1, 1)

		}
	}
	fmt.Println("cout")
	go UEFlowMapToStore()
	go TimeFlowMapToStore()
}

func Count_UE_ID(packet *Packet, taskid string) (uint64, bool) {
	RAN_UE_NGAP_ID := packet.RAN_UE_NGAP_ID
	if RAN_UE_NGAP_ID == -1 {
		return 0, false
	}
	flowID := FastTwoHash([]byte(strconv.FormatInt(RAN_UE_NGAP_ID, 10)), []byte(taskid))
	return flowID, true

}

func Count_Time_ID(packet *Packet, TimeFirst int64, taskid string) (uint64, bool) {
	RAN_UE_NGAP_ID := packet.RAN_UE_NGAP_ID
	if RAN_UE_NGAP_ID == -1 {
		return 0, false
	}
	Time := packet.ArriveTimeUs - TimeFirst
	Time = Time / TimeChip
	return FastTwoHash([]byte(strconv.FormatInt(Time, 10)), []byte(taskid)), true
}

func Put(packet *Packet, flowTable []*Flow, flowID string, taskid string, flowType string) bool {
	var flowInfo *FlowInfo
	var first = false // 是否流的首包
	numFlowId, _ := strconv.ParseUint(flowID, 10, 64)
	if flowType == "UE"{
		//fmt.Println("UE")
		//fmt.Println(numFlowId%TABLE_SIZE)
		UE_SEM[numFlowId%TABLE_SIZE].P()
	}
	if flowType == "Time"{
		//fmt.Println("TIME")
		//fmt.Println(numFlowId%TABLE_SIZE)
		TIME_SEM[numFlowId%TABLE_SIZE].P()
	}
	flowInfo, isExist := loadFlow(flowID, flowTable)
	if isExist {
		packet.TimeInterval = uint64(packet.ArriveTimeUs - flowInfo.EndTimeUs + min_intervl)
		flowInfo.EndTime = packet.ArriveTime
		flowInfo.EndTimeUs = packet.ArriveTimeUs
		flowInfo.PacketList.PushBack(packet)
		flowInfo.TotalNum = flowInfo.TotalNum + 1
		if flowInfo.SrcIP == packet.SrcIP {
			packet.DirSeq = 1
		} else {
			packet.DirSeq = -1
		}
	} else {
		// 首次接收，创建流info
		flowInfo = &FlowInfo{
			FlowID:          flowID,
			RAN_UE_NGAP_ID:  packet.RAN_UE_NGAP_ID,
			FlowType:        NGAPType,
			TotalNum:        1,
			VerificationTag: packet.VerificationTag,
			SrcIP:           packet.SrcIP,
			DstIP:           packet.DstIP,
			TimeID:          packet.TimeID,
			BeginTime:       packet.ArriveTime,
			TaskID:          taskid,
		}
		packet.DirSeq = 1
		packet.TimeInterval = uint64(min_intervl)
		flowInfo.EndTime = packet.ArriveTime
		flowInfo.EndTimeUs = packet.ArriveTimeUs
		flowInfo.PacketList = list.List{}
		flowInfo.PacketList.PushBack(packet)
		storeFlow(flowID, flowInfo, flowTable)
		if flowType == "UE"{
			flowCount_UE++
		}
		if flowType == "Time"{
			flowCount_Time++
		}

		first = true
	}
	if flowType == "UE"{
		UE_SEM[numFlowId%TABLE_SIZE].V()
		//fmt.Println("UE out")
		//fmt.Println(numFlowId%TABLE_SIZE)
	}
	if flowType == "Time"{
		TIME_SEM[numFlowId%TABLE_SIZE].V()
		//fmt.Println("TIME out")
		//fmt.Println(numFlowId%TABLE_SIZE)
	}
	return first
}



func UEFlowMapToStore()  {
	//fmt.Println("UE:")
	for true{
		var rubbishList = list.New()
		for i:=0;i<int(TABLE_SIZE);i++{
			UE_SEM[i].P()
			flow :=FlowTable_UE[i]
			if flow == nil {
				UE_SEM[i].V()
				continue
			}else{
				
				for cur := flow; cur != nil; cur = cur.next {
					flowInfo := cur.info
					fmt.Println("id:",flowInfo.FlowID)
					fmt.Println("UEFLOW:",i)
					nowTime := time.Now().UnixNano()
					if nowTime - flowInfo.EndTimeUs < 5e9{
						fmt.Println("flowId: ",flowInfo.FlowID,",  time: ",flowInfo.EndTimeUs)
						rubbishList.PushBack(flowInfo)
						fmt.Println(deleteFlow(flowInfo.FlowID,FlowTable_UE))
						flowCount_UE--
					}
	
				}
				UE_SEM[i].V()
			}
			
		}
		//fmt.Println("UEFlow")
		if rubbishList.Len() > 0{
			go UEflowStore(rubbishList)
		}
		
	}

}

func UEflowStore(rubbishList *list.List)  {
	var UeFlowList = list.New()
	var PacketList = list.New()
	for info := rubbishList.Front(); info != nil; info = info.Next(){
		flowInfo := info.Value.(*FlowInfo)
		fl := &UEFlow.UeFlow{
			FlowId: string(flowInfo.FlowID),
			RanUeNgapId: uint64(flowInfo.RAN_UE_NGAP_ID),
			TotalNum: uint32(flowInfo.TotalNum),
			BeginTime: flowInfo.BeginTime,
			LatestTime: flowInfo.EndTime,
			VerificationTag: uint64(flowInfo.VerificationTag),
			SrcIP: flowInfo.SrcIP,
			DstIP: flowInfo.DstIP,
			//TimeID          uint64
			StatusFlow: 0,
		}
		fmt.Println("ue_flow_info:",fl.FlowId)
		UeFlowList.PushBack(fl)
		for cur := flowInfo.PacketList.Front(); cur != nil; cur = cur.Next(){
			parse:= cur.Value.(*Packet)
			packet:= &PacketDB.Packet{
				//PacketId: FnvHash([]byte(string(parse.ArriveTimeUs))),
				NgapType: parse.NgapType,
				NgapProcedureCode: parse.NgapProcedureCode,
				RanUeNgapId: int64(parse.RAN_UE_NGAP_ID),
				PacketLen: uint32(parse.PacketLen),
				ArriveTimeUs: uint64(parse.ArriveTimeUs),
				ArriveTime: parse.ArriveTime,
				TimeInterval: uint64(parse.TimeInterval),
				VerificationTag: uint64(parse.VerificationTag),
				SrcIP: parse.SrcIP,
				DstIP: parse.DstIP,
				DirSeq: uint16(parse.DirSeq),
				FlowUEID: string(parse.FlowID),
				FlowTimeID: string(parse.TimeID),
				StatusPacket: 0,
			}
			fmt.Println("packet_ue_info:",packet.NgapType)
			PacketList.PushBack(packet)
		}

	}
	//fmt.Println("UEList")
	PacketDB.InsertPacketUE(PacketList)
	UEFlow.InsertUeFlow(UeFlowList)
	//fmt.Println("UEList2")

	
}

func TimeFlowMapToStore() {
	//fmt.Println("time:")
	for true{
		var rubbishList = list.New()
		for i:=0;i<int(TABLE_SIZE);i++{
			//fmt.Println("TIMEFLOW:",i)
			TIME_SEM[i].P()
			flow := FlowTable_Time[i]
			//fmt.Println("judge time:")
			if flow == nil {
				TIME_SEM[i].V()
				continue
			}else{
				nowTime := time.Now().UnixNano()
				for cur := flow; cur != nil; cur = cur.next {
					flowInfo := cur.info
					fmt.Println("id:",flowInfo.FlowID)
					fmt.Println("TIMEFLOW:",i)
					time.Sleep(30)
					if nowTime - flowInfo.EndTimeUs < 5e9{
						fmt.Println("flowId: ",flowInfo.FlowID,",  time: ",flowInfo.EndTimeUs)
						rubbishList.PushBack(flowInfo)
						fmt.Println(deleteFlow(flowInfo.FlowID,FlowTable_Time))
						flowCount_Time--
					}
				}
				TIME_SEM[i].V()
			}
			
		}
		//fmt.Println("TimeFlow")
		if rubbishList.Len() > 0{
			go TimeflowStore(rubbishList)
		}
		
	}

}

func TimeflowStore(rubbishList *list.List)  {
	var TimeFlowList = list.New()
	var PacketList = list.New()
	for info := rubbishList.Front(); info != nil; info = info.Next(){
		flowInfo := info.Value.(*FlowInfo)
		
		fl := &TimeFlow.TimeFlow{
			FlowId: string(flowInfo.FlowID),
			RanUeNgapId: uint64(flowInfo.RAN_UE_NGAP_ID),
			TotalNum: uint32(flowInfo.TotalNum),
			BeginTime: flowInfo.BeginTime,
			LatestTime: flowInfo.EndTime,
			VerificationTag: uint64(flowInfo.VerificationTag),
			SrcIP: flowInfo.SrcIP,
			DstIP: flowInfo.DstIP,
			//TimeID          uint64
			StatusFlow: 0,
		}
		fmt.Println("time_flow_info:",fl.FlowId)
		TimeFlowList.PushBack(fl)


		for cur := flowInfo.PacketList.Front(); cur != nil; cur = cur.Next(){
			parse:= cur.Value.(*Packet)
			
			packet:= &PacketDB.Packet{
				//PacketId: FnvHash([]byte(string(parse.ArriveTimeUs))),
				NgapType: parse.NgapType,
				NgapProcedureCode: parse.NgapProcedureCode,
				RanUeNgapId: int64(parse.RAN_UE_NGAP_ID),
				PacketLen: uint32(parse.PacketLen),
				ArriveTimeUs: uint64(parse.ArriveTimeUs),
				ArriveTime: parse.ArriveTime,
				TimeInterval: uint64(parse.TimeInterval),
				VerificationTag: uint64(parse.VerificationTag),
				SrcIP: parse.SrcIP,
				DstIP: parse.DstIP,
				DirSeq: uint16(parse.DirSeq),
				FlowUEID: string(0),
				FlowTimeID: string(parse.TimeID),
				StatusPacket: 0,
			}
			fmt.Println("packet_time_info:",packet.NgapType)
			PacketList.PushBack(packet)
		}
	}
	//fmt.Println("TimeList")
	PacketDB.InsertPacketTime(PacketList)
	TimeFlow.InsertTimeFlow(TimeFlowList)

}
