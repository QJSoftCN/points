package tpri

import (
	"../std"
	"unsafe"
	"time"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"../confs"
)

const (
	DBName = "TPRI"
)

type TPRIDriver struct {
	name string
}

type TPRIDB struct {
	connUrl  string
	connSize int
	conns    *[]TPRIConn
	driver   *TPRIDriver
	points   *[]points.Point
	psIndex  *map[string]int
}

type TPRIConn struct {
	dwHandle     uintptr
	createTime   time.Time
	count        int
	lastCallTime time.Time
}

func (this *TPRIConn) getDwHandle() uint32 {
	return uint32(this.dwHandle)
}

func (this *TPRIConn) calling() {
	this.count++
	this.lastCallTime = time.Now()

	fmt.Println("tpri conn call ", this.dwHandle, " count:", this.count)
}

func (this *TPRIDB) getConn() TPRIConn {
	//random get conn
	conn := (*this.conns)[rand.Intn(len(*this.conns))]
	conn.calling()
	return conn
}

func newTPRIConn(hptrs, uptrs, pptrs, ports, nsize uintptr) (*TPRIConn, error) {
	handle, _, _ := DBPCreate.Call(hptrs, uptrs, pptrs, ports, nsize)
	//conn dbp
	ret, _, _ := DBPConnect.Call(uintptr(handle))

	if ret != 0 {
		return nil, errors.New(fmt.Sprint("connect fail:", ret))
	}

	conn := TPRIConn{}
	conn.dwHandle = handle
	conn.createTime = time.Now()

	return &conn, nil
}

func (this TPRIDriver) Connect(connUrl string, connSize int) (points.RealDB, error) {
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

	if connSize <= 0 {
		connSize = 10
	}

	rdb := TPRIDB{}
	rdb.connSize = connSize
	rdb.connUrl = connUrl
	rdb.driver = &this

	conns := make([]TPRIConn, connSize)

	for i := 0; i < connSize; i++ {
		c, err := newTPRIConn(h, u, p, portr, n)
		if err == nil {
			conns[i] = *c
		} else {
			log.Println("new tpri conn err:", err)
		}
	}
	//设置连接池
	rdb.conns = &conns
	rdb.loadPoints()

	return rdb, nil
}

func (this *TPRIDB) loadPoints() {

	dir := confs.TempDir()
	ps := points.LoadPoints(dir, this.driver.name)

	if ps == nil {
		ps, err := this.ReadPoints()
		if err != nil {
			log.Println("load std err:", err)
			return
		}
		points.StorePoints(dir, this.driver.name, ps)
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
	tpriDriver := TPRIDriver{DBName}
	points.Reg(DBName, tpriDriver)
}
