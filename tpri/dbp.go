package tpri

import (
	"syscall"
)

const (
	DllName = "dbpapi_x64.dll"
)

var (
	DBPAPI                 = syscall.NewLazyDLL(DllName)
	DBPCreate              = DBPAPI.NewProc("DBPCreate")
	DBPConnect             = DBPAPI.NewProc("DBPConnect")
	DBPIsConnect           = DBPAPI.NewProc("DBPIsConnect")
	DBPGetSnapshot         = DBPAPI.NewProc("DBPGetSnapshot")
	DBPQueryTagFromDbp2    = DBPAPI.NewProc("DBPQueryTagFromDbp2")
	DBPEnumTagAttr         = DBPAPI.NewProc("DBPEnumTagAttr")
	DBPGetHisVal           = DBPAPI.NewProc("DBPGetHisVal")
	DBPGetMultiPointHisVal = DBPAPI.NewProc("DBPGetMultiPointHisVal")
)
