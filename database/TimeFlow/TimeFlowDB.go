package TimeFlow

import (
	"container/list"
	"github.com/Freddy/sctp_flowmap/database"
	"log"
	"fmt"
)

func init() {
	creatTimeFlowTable()
}

func InsertTimeFlow(TimeFlowList *list.List) {
	var sqlStr string = insertTimeFlowSQL

	tx, err := database.Connect.Begin()
	checkErr(err)
	stmt, err := tx.Prepare(sqlStr)
	checkErr(err)
	for info := TimeFlowList.Front(); info != nil; info = info.Next() {
		fl := info.Value.(*TimeFlow)
		fmt.Println("db_flow_time:",fl.FlowId)
		if _, err := stmt.Exec(
			fl.FlowId,
			fl.RanUeNgapId,
			fl.TotalNum,
			fl.BeginTime,
			fl.LatestTime,
			fl.VerificationTag,
			fl.SrcIP,
			fl.DstIP,
			fl.StatusFlow,
			fl.TaskID,
		); err != nil {
			log.Fatal(err)
		}
	}
	checkErr(tx.Commit())
}

func creatTimeFlowTable() {
	var sqlStr string = createTimeFlowTableSQL

	_, err := database.Connect.Exec(sqlStr)
	checkErr(err)
}

func dropTimeFlowTable() {
	if _, err := database.Connect.Exec(dropTimeFlowTableSQL); err != nil {
		log.Fatal(err)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
