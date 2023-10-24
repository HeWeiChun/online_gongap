package main

import (
	"flag"
	"fmt"

	"github.com/Freddy/sctp_flowmap/database"
	"github.com/Freddy/sctp_flowmap/decoder"
)

func main() {
	pcapPath := flag.String("pcap_path", "", "Path to the pcap file")
	taskID := flag.String("taskid", "", "The task ID")
	mode := flag.String("mode", "", "The detect mode") // 在线检测(1), 离线检测(0)
	flag.Parse()
	_, err := database.Connect.Begin()
	database.CheckErr(err)

	if *pcapPath == "" {
		fmt.Println("pcap_path参数未提供")
		return
	}
	if *taskID == "" {
		fmt.Println("pcap_path参数未提供")
		return
	}
	if *mode == "" {
		fmt.Println("mode参数未提供")
		return
	}
	decoder.Decode(*pcapPath, *taskID, *mode)

	// fmt.Println("等待存储到数据库")
	// flowMap.UEFlowMapToStore()
	// flowMap.TimeFlowMapToStore(*taskID)
}
