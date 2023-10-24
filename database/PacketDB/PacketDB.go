package PacketDB

import (
	"container/list"
	_ "container/list"
	"fmt"
	"github.com/Freddy/sctp_flowmap/database"
	"log"
)


func init() {
	creatPacketUETable()
	creatPacketTimeTable()
}

func InsertPacketUE(PacketList *list.List) {

	var sqlStr string = insertPacketUESQL

	tx, err := database.Connect.Begin()
	checkErr(err)
	stmt, err := tx.Prepare(sqlStr)
	checkErr(err)
	for info := PacketList.Front(); info != nil; info = info.Next(){
		packet_sql := info.Value.(*Packet)
		fmt.Println("db_ue_packet:",packet_sql.NgapType)
		if _, err := stmt.Exec(
			//packet_sql.PacketId,
			packet_sql.NgapType,
			packet_sql.NgapProcedureCode,
			packet_sql.RanUeNgapId,
			packet_sql.PacketLen,
			packet_sql.ArriveTimeUs,
			packet_sql.ArriveTime,
			packet_sql.TimeInterval,
			packet_sql.VerificationTag,
			packet_sql.SrcIP,
			packet_sql.DstIP,
			packet_sql.DirSeq,
			packet_sql.FlowUEID,
			packet_sql.FlowTimeID,
			packet_sql.StatusPacket,
		); err != nil {
			log.Fatal(err)
		}

	}
	checkErr(tx.Commit())
}

func InsertPacketTime(PacketList *list.List) {

	var sqlStr string = insertPacketTimeSQL

	tx, err := database.Connect.Begin()
	checkErr(err)
	stmt, err := tx.Prepare(sqlStr)
	checkErr(err)
	for info := PacketList.Front(); info != nil; info = info.Next(){
		packet_sql := info.Value.(*Packet)
		fmt.Println("db_time_packet:",packet_sql.NgapType)
		if _, err := stmt.Exec(
			//packet_sql.PacketId,
			packet_sql.NgapType,
			packet_sql.NgapProcedureCode,
			packet_sql.RanUeNgapId,
			packet_sql.PacketLen,
			packet_sql.ArriveTimeUs,
			packet_sql.ArriveTime,
			packet_sql.TimeInterval,
			packet_sql.VerificationTag,
			packet_sql.SrcIP,
			packet_sql.DstIP,
			packet_sql.DirSeq,
			packet_sql.FlowUEID,
			packet_sql.FlowTimeID,
			packet_sql.StatusPacket,
		); err != nil {
			log.Fatal(err)
		}

	}
	checkErr(tx.Commit())
}
/*
func QueryFlows() {
	var sqlStr string = queryPacketSQL

	rows, err := database.Connect.Query(sqlStr)
	checkErr(err)
	for rows.Next() {
		var packet Packet
		checkErr(rows.Scan(
			&packet.
			))

		fmt.Println(packet) //printf log
	}
}

*/



func creatPacketUETable() {
	var sqlStr string = creatPacketTableUESQL

	_, err := database.Connect.Exec(sqlStr)
	checkErr(err)
}

func creatPacketTimeTable() {
	var sqlStr string = creatPacketTableTimeSQL

	_, err := database.Connect.Exec(sqlStr)
	checkErr(err)
}

func dropflowLogTableUE() {
	if _, err := database.Connect.Exec(dropPacketTableUESQL); err != nil {
		log.Fatal(err)
	}
}
func dropflowLogTableTime() {
	if _, err := database.Connect.Exec(dropPacketTableTimeSQL); err != nil {
		log.Fatal(err)
	}
}
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}





