package sim

import (
	"../std"
	"log"
	"github.com/qjsoftcn/confs"
)

//make sim from realdb
func MakeSim(supplier, driver string, db points.RealDB) {
	dir := confs.SimDir()
	psName := supplier + "_" + driver
	ps, err := db.ReadPoints()
	if err==nil{
		points.StorePoints(dir, psName, ps)
	}else{
		log.Println("make ",psName," err:",err)
	}

	//make point features
	points.BuildPFS()



}
