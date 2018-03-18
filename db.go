package points

import (
	"log"
	"strings"
	"errors"
)

var rdrs map[string]RealDBDriver

func Open(dbName, connUrl string, connSize int) (RealDB, error) {

	dbName = strings.ToUpper(dbName)
	dr, ok := rdrs[dbName]
	if ok {
		db, err := dr.Connect(connUrl, connSize)
		if err != nil {
			log.Println("driver:", dbName, "connect:", connUrl, " err:", err)
		}
		return db, err
	} else {
		log.Println(dbName, "not exist,please check driver")
		return nil, errors.New("no " + dbName + " driver")
	}
}

func Reg(dbName string, db RealDBDriver) {
	dbName = strings.ToUpper(dbName)
	rdrs[dbName] = db
	log.Println("register driver:", dbName, " success")
}

func init() {
	rdrs = make(map[string]RealDBDriver)
	log.Println("init rdrs")
	//init point features info

}
