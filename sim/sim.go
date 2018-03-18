package sim

import (
	"log"
	"strings"
	"sync"
	"time"
	"fmt"
	"unsafe"
	"../std"
	"math/rand"
)

const (
	DB_Name = "SIM"
)

type SIMDriver struct {
	name string
}

type SIMDB struct {
	connUrl  string
	connSize int
	driver   *SIMDriver
	sync.Mutex
	conns    *[]SIMConn
	points   *[]points.Point
	psIndex  *map[string]int
}

type SIMConn struct {
	dwHandle     uintptr
	createTime   time.Time
	count        int
	lastCallTime time.Time
}

func (this *SIMConn) getDwHandle() uint32 {
	return uint32(this.dwHandle)
}

func (this *SIMConn) calling() {
	this.count++
	this.lastCallTime = time.Now()

	fmt.Println("sim conn call ", this.dwHandle, " count:", this.count)
}

func (this *SIMDB) getConn() SIMConn {
	//random get conn
	conn := (*this.conns)[rand.Intn(len(*this.conns))]
	conn.calling()
	return conn
}

func newTPRIConn(hptrs, uptrs, pptrs, ports, nsize uintptr) (*SIMConn, error) {
	conn := SIMConn{}
	conn.dwHandle = uintptr(rand.Intn(5))
	conn.createTime = time.Now()

	return &conn, nil
}

func (this SIMDriver) Connect(connUrl string, connSize int) (points.RealDB, error) {
	//admin/admin@192.168.101.248:12084?20;admin/admin@192.168.101.248:12084?20;
	hosts, ports, users, pwds := parseConnURL(connUrl)

	hptrs := points.BytePtr(hosts)
	uptrs := points.BytePtr(users)
	pptrs := points.BytePtr(pwds)
	nsize := len(hosts)

	h := uintptr(unsafe.Pointer(&hptrs[0]))
	u := uintptr(unsafe.Pointer(&uptrs[0]))
	p := uintptr(unsafe.Pointer(&pptrs[0]))
	portr := uintptr(unsafe.Pointer(&ports[0]))
	n := points.IntPtr(nsize)

	rdb := SIMDB{}
	rdb.connSize = connSize
	rdb.connUrl = connUrl
	rdb.driver = &this

	conns := make([]SIMConn, connSize)

	for i := 0; i < connSize; i++ {
		c, err := newTPRIConn(h, u, p, portr, n)
		if err == nil {
			conns[i] = *c
		} else {
			log.Println("new tpri conn err:", err)
		}
	}

	rdb.conns = &conns
	rdb.loadPoints()

	return rdb, nil
}

func (this *SIMDB) loadPoints() {

	ps, err := this.ReadPoints()
	if err != nil {
		log.Println("load std err:", err)
		return
	}

	pointMap := make(map[string]int)
	for index, p := range ps {
		key := strings.ToLower(p.Name)
		pointMap[key] = index
	}

	this.points = &ps
	this.psIndex = &pointMap

	log.Println("load std size:", len(ps))
}

func init() {
	tpriDriver := SIMDriver{DB_Name}
	points.Reg(DB_Name, tpriDriver)

}
