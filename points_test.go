package points

import (
	"testing"
	"fmt"
	"github.com/qjsoftcn/points/points"
)

func TestOpen(t *testing.T){
	//admin/admin@tpri
	db,err:=points.Open("tpri","admin/admin@192.168.101.248:12084",20)
	if err!=nil{
		fmt.Println(err)
	}

	fmt.Println(db.GetDBName())

}
