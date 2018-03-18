package tpri

import (
	"unsafe"
	"../points"
	"errors"
	"strconv"
	"strings"
	"fmt"
)

func (this TPRIDB) ReadSnapshot(name string) (*points.PointValue, error) {

	names := make([]string, 1)
	names[0] = name

	ps, err := this.ReadSnapshots(names)
	if len(ps) == 1 {
		return &ps[0], err
	} else {
		return nil, err

	}
}

func (this TPRIDB) GetPoint(name string) *points.Point {
	name = strings.ToLower(name)
	index, ok := (*this.psIndex)[name]
	if ok {
		return &((*this.points)[index])
	} else {
		return nil
	}
}

func filterByStr(p *points.Point, str, xstr string) *points.Point {
	if len(xstr) == 0 {
		return p
	} else {
		if strings.Contains(strings.ToUpper(str), strings.ToUpper(xstr)) {
			return p
		} else {
			return nil
		}
	}
}

func filterByArray(p *points.Point, t int, ts []int) *points.Point {
	if len(ts) == 0 {
		return p
	} else {
		for _, x := range ts {
			if t == x {
				return p
			}
		}

		return nil
	}
}

func match(p *points.Point, name, desc string, pts, vts []int) bool {

	p = filterByStr(p, p.Name, name)
	if p == nil {
		return false
	}

	p = filterByStr(p, p.Desc, desc)
	if p == nil {
		return false
	}

	p = filterByArray(p, p.Type.Id, pts)
	if p == nil {
		return false
	}

	p = filterByArray(p, p.ValType.Id, vts)
	if p == nil {
		return false
	}

	return true
}

func (this TPRIDB) GetPoints(name, desc string, pts, vts []int) []points.Point {
	nps := make([]points.Point, 0)

	for _, point := range (*this.points) {
		if match(&point, name, desc, pts, vts) {
			nps = append(nps, point)
		}
	}

	return nps

}

func (this TPRIDB) ReadSnapshots(names []string) ([]points.PointValue, error) {

	conn := this.getConn()
	nsize := len(names)

	tagNames := points.BytePtr(names)

	dblvals := make([]float64, nsize)
	ltimes := make([]int32, nsize)
	snqas := make([]int16, nsize)
	lvals := make([]int32, nsize)
	ntypes := make([]int32, nsize)
	errs := make([]int16, nsize)

	ret, _, _ := DBPGetSnapshot.Call(conn.dwHandle,
		uintptr(unsafe.Pointer(&tagNames[0])),
		uintptr(unsafe.Pointer(&ltimes[0])),
		uintptr(unsafe.Pointer(&snqas[0])),
		uintptr(unsafe.Pointer(&dblvals[0])),
		uintptr(unsafe.Pointer(&lvals[0])),
		uintptr(unsafe.Pointer(&ntypes[0])),
		uintptr(unsafe.Pointer(&errs[0])),
		points.IntPtr(nsize))

	if ret == 0 {
		values := make([]points.PointValue, nsize)
		for i := 0; i < nsize; i++ {
			tn := points.BytePtrToString(tagNames[i])
			point := this.GetPoint(tn)
			pv := toPointValue(point, ltimes[i], snqas[i], dblvals[i], lvals[i], "", ntypes[i], errs[i])

			values[i] = pv
		}
		return values, nil
	} else {
		return nil, newErr(ret)
	}
}

const (
	vt_void    = iota
	vt_digital
	vt_int32
	vt_float32
	vt_int64
	vt_float64
	vt_string
	vt_blob
)

func toVal(fval float64, lval int32, sval string, tt int32) interface{} {
	switch tt {
	case vt_digital, vt_int32, vt_int64:
		return lval
	case vt_float32, vt_float64:
		return fval
	case vt_string:
		return sval
	default:
		return nil
	}
}

func toPointValue(p *points.Point, sec int32, quality int16, fval float64, lval int32, sval string, tt int32, err int16) points.PointValue {
	pv := points.NewPointValue(p, sec, toVal(fval, lval, sval, tt), quality, err)
	return pv
}

func newErr(ret uintptr) error {
	return errors.New(strconv.Itoa(int(ret)))
}

const (
	once_maxsize = 86400
)

const (
	his_raw   = iota
	his_inter
	his_plot
)

const (
	err_no_tag = 9
)

func (this TPRIDB) readHistory(p *points.Point, tn, ivt, ht uintptr, s, e int32, vals []points.PointValue) ([]points.PointValue, error, bool) {

	conn := this.getConn()
	isEnd := true

	var nsize, pt, rsize int32
	nsize = int32(e - s + 1)
	if nsize > once_maxsize {
		nsize = once_maxsize
		isEnd = false
	}

	dblvals := make([]float64, nsize)
	ltimes := make([]int32, nsize)
	snqas := make([]int16, nsize)
	lvals := make([]int32, nsize)

	ret, _, _ := DBPGetHisVal.Call(conn.dwHandle,
		tn,
		uintptr(s),
		uintptr(e),
		ivt,
		ht,
		uintptr(unsafe.Pointer(&dblvals[0])),
		uintptr(unsafe.Pointer(&lvals[0])),
		uintptr(unsafe.Pointer(&ltimes[0])),
		uintptr(unsafe.Pointer(&snqas[0])),
		uintptr(nsize),
		uintptr(unsafe.Pointer(&pt)),
		uintptr(unsafe.Pointer(&rsize)))

		fmt.Println(s,e,ret)

	if ret == 0 {
		if rsize > 0 {
			values := make([]points.PointValue, rsize)
			for i := 0; i < int(rsize); i++ {
				values[i] = toPointValue(p, ltimes[i], snqas[i], dblvals[i], lvals[i], "", pt, 0)
			}

			x := 0

			valsLen := len(vals)
			if valsLen > 0 {
				valsEnd := vals[valsLen-1].GetSec()
				for index, p := range values {
					if p.GetSec() > valsEnd {
						x = index
						break
					}
				}
			}

			vals = append(vals, values[x:]...)

		}

		if isEnd {
			return vals, nil, isEnd
		} else {
			lastTime := ltimes[rsize-1]

			if lastTime < e {
				return this.readHistory(p, tn, ivt, ht, lastTime, e, vals)
			} else {
				return vals, nil, true
			}

		}

		return vals, nil, isEnd
	} else {

		return vals, newErr(ret), true
	}

}

func (this TPRIDB) ReadHistory(name string, start, end int32, way points.HistorysComplementWay) (*points.HistoryVals, error) {
	p := this.GetPoint(name)
	if p == nil {
		return nil, newErr(err_no_tag)
	}

	nptr, ivt, ht := uintptr(unsafe.Pointer(points.StringToINT8Ptr(name))), uintptr(0), uintptr(his_raw)

	var vals []points.PointValue
	vals, err, _ := this.readHistory(p, nptr, ivt, ht, start, end, vals)

	hval := points.NewHistoryVals(&vals)
	hval.Set(start, end, way)

	return &hval, err

}

func (this TPRIDB) InterVal(name string, t int32, way points.InterWay) (*points.PointValue, error) {

	names := make([]string, 1)
	names[0] = name

	ps, err := this.InterVals(names,t,way)
	if len(ps) == 1 {
		return &ps[0], err
	} else {
		return nil, err
	}
}


func (this TPRIDB) InterVals(names []string, t int32, way points.InterWay) ([]points.PointValue, error) {

	conn := this.getConn()
	nsize := len(names)

	tagNames := points.BytePtr(names)

	dblvals := make([]float64, nsize)
	ltimes := make([]int32, nsize)
	snqas := make([]int16, nsize)
	lvals := make([]int32, nsize)
	ntypes := make([]int32, nsize)
	errs := make([]int16, nsize)

	for i:=0;i<nsize;i++{
		ltimes[i]=t
	}

	ret, _, _ := DBPGetMultiPointHisVal.Call(conn.dwHandle,
		uintptr(int32(way)),
		uintptr(unsafe.Pointer(&tagNames[0])),
		uintptr(unsafe.Pointer(&ltimes[0])),
		uintptr(unsafe.Pointer(&ntypes[0])),
		uintptr(unsafe.Pointer(&dblvals[0])),
		uintptr(unsafe.Pointer(&lvals[0])),
		uintptr(unsafe.Pointer(&snqas[0])),
		uintptr(unsafe.Pointer(&errs[0])),
		points.IntPtr(nsize))

	if ret == 0 {
		values := make([]points.PointValue, nsize)
		for i := 0; i < nsize; i++ {
			tn := points.BytePtrToString(tagNames[i])
			point := this.GetPoint(tn)
			pv := toPointValue(point, ltimes[i], snqas[i], dblvals[i], lvals[i], "", ntypes[i], errs[i])

			values[i] = pv
		}
		return values, nil
	} else {
		return nil, newErr(ret)
	}
}

func (this TPRIDB) GetDBName() string {
	return this.driver.name
}
func (this TPRIDB) GetConnUrl() string {
	return this.connUrl
}
func (this TPRIDB) GetConnSize() int {
	return this.connSize
}

func (this TPRIDB) Close() error {

	return nil
}

func (this TPRIDB) isConnected() bool {
	conn := this.getConn()
	var isConn int
	isConn = -1

	ret, _, _ := DBPIsConnect.Call(conn.dwHandle, uintptr(unsafe.Pointer(&isConn)))
	if ret != 0 {
		return false
	}

	return isConn != 0
}

func (this TPRIDB) ReadPoints() ([]points.Point, error) {

	if this.points != nil {
		return *this.points, nil
	}

	conn := this.getConn()
	tagSize := 0
	ret, _, err := DBPQueryTagFromDbp2.Call(conn.dwHandle,
		uintptr(0),
		uintptr(0),
		points.IntPtr(-1),
		points.IntPtr(-1),
		uintptr(unsafe.Pointer(&tagSize)))

	if ret != 0 {
		return nil, err
	}

	var sname, sdbtagname, sdesc [80]byte
	var sdbname, sunit [16]byte
	var tagid, vt, ltagflag int32
	var errcode int16

	ps := make([]points.Point, tagSize)

	for i := 0; i < tagSize; i++ {
		ret, _, err = DBPEnumTagAttr.Call(conn.dwHandle,
			uintptr(unsafe.Pointer(&sname[0])),
			uintptr(unsafe.Pointer(&sdbname[0])),
			uintptr(unsafe.Pointer(&sdbtagname[0])),
			uintptr(unsafe.Pointer(&sdesc[0])),
			uintptr(unsafe.Pointer(&sunit[0])),
			uintptr(unsafe.Pointer(&tagid)),
			uintptr(unsafe.Pointer(&vt)),
			uintptr(unsafe.Pointer(&errcode)),
			uintptr(unsafe.Pointer(&ltagflag)),
			points.IntPtr(i))

		if ret == 0 {
			tagName := points.BytePtrToString((*byte)(unsafe.Pointer(&sname[0])))
			dbName := points.BytePtrToString((*byte)(unsafe.Pointer(&sdbname[0])))
			dbTagName := points.BytePtrToString((*byte)(unsafe.Pointer(&sdbtagname[0])))
			tagDesc := points.BytePtrToString((*byte)(unsafe.Pointer(&sdesc[0])))
			vU := points.BytePtrToString((*byte)(unsafe.Pointer(&sunit[0])))
			p := toPoint(tagid, tagName, tagDesc, vU, dbTagName, dbName, vt, ltagflag, errcode)

			if p != nil {
				ps[i] = *p
			}

		}

	}

	return ps, nil
}

func toPoint(id int32, tn, td, vU, dbtn, dbn string, vt, tt int32, errCode int16) *points.Point {
	if errCode == 0 {
		point := points.Point{}
		point.Id = int(id)
		point.Name = tn
		point.Desc = td
		point.Unit = vU
		point.DB = dbn
		point.DBName = dbtn
		point.Type = points.GetType(int(tt))
		point.ValType = points.GetValType(int(vt))
		point.DefaultFmt = points.GetDefaultFmt()
		return &point
	} else {
		return nil
	}

}
