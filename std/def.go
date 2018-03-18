package points

import (
	"fmt"
	"time"
	"github.com/qjsoftcn/confs"
	"github.com/qjsoftcn/gutils"
)

type PointType struct {
	Id   int
	Name string
}

type ValType struct {
	Id   int
	Name string
}

type Point struct {
	Id         int
	Name       string
	Desc       string
	Unit       string
	Type       *PointType
	ValType    *ValType
	DB         string
	DBName     string
	DefaultFmt string
}

type PointValue struct {
	point   *Point
	val     interface{}
	sec     int32
	quality int16
	err     int16
}

type HistoryVals struct {
	Vals  *[]PointValue
	Start int32
	End   int32
	Way   HistorysComplementWay
}

func (this *HistoryVals) Accept(vs ...HistorysVisitor) {
	for index, val := range *this.Vals {
		for _, v := range vs {
			v.Visit(index, &val)
		}
	}
}

func (this *HistoryVals) Set(start, end int32, way HistorysComplementWay) {
	this.Start = start
	this.End = end
	this.Way = way
}

func (this *HistoryVals) GetEndVal() *PointValue {
	size := len(*this.Vals)
	if size == 0 {
		return nil
	} else {
		return &(*this.Vals)[size-1]
	}
}

func (this *PointValue) StringVal() string {
	s, ok := this.val.(string)
	if ok {
		return s
	} else {
		return ""
	}
}

func (this *PointValue) IntVal() int32 {
	s, ok := this.val.(int32)
	if ok {
		return s
	} else {
		return 0
	}
}

func (this *PointValue) FloatVal() float64 {
	s, ok := this.val.(float64)
	if ok {
		return s
	} else {
		return 0
	}
}

func (this *PointValue) Format(decimalNum int) string {
	switch value := this.val.(type) {
	case string, int32, int16, int64, int:
		return fmt.Sprint(value)
	case float64, float32:
		f := fmt.Sprint("%.", decimalNum, "f")
		return fmt.Sprintf(f, value)
	default:
		return fmt.Sprint(value)
	}
}

func (this *PointValue) GetSec() int32 {
	return this.sec
}

func (this *PointValue) GetTime() time.Time {
	return time.Unix(int64(this.sec), 0)
}

func (this *PointValue) GetTimeString() string {
	return gutils.Format(this.GetTime(),gutils.TF_Sec)
}

func (this *PointValue) GetQuality() int16 {
	return this.quality
}

func (this *PointValue) GetErr() int16 {
	return this.err
}

func (this *PointValue) IsOK() bool {
	if this.err == 0 {
		return QualityIsGood(this.GetQuality())
	}
	return false
}

func (this *PointValue) ValToString() string {
	switch value := this.val.(type) {
	case string, int32, int16, int64, int:
		return fmt.Sprint(value)
	case float64, float32:
		return fmt.Sprint(value)
	default:
		return fmt.Sprint(value)
	}
}

func QualityIsGood(quality int16) bool {
	return quality == 0
}

const (
	t_Real   = iota
	t_Manual
	t_Pre
	t_Curve
)

const (
	conf_group   = "point"
	conf_type    = "type"
	conf_valtype = "valType"
)

var (
	tR = &PointType{t_Real, confs.GetString(conf_group, conf_type, "real", confs.Locale())}
	tM = &PointType{t_Manual, confs.GetString(conf_group, conf_type, "manual", confs.Locale())}
	tP = &PointType{t_Pre, confs.GetString(conf_group, conf_type, "pre", confs.Locale())}
	tC = &PointType{t_Curve, confs.GetString(conf_group, conf_type, "curve", confs.Locale())}
)

func GetType(t int) *PointType {
	switch t {
	case t_Real:
		return tR
	case t_Manual:
		return tM
	case t_Pre:
		return tP
	case t_Curve:
		return tC
	default:
		return tR
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

var (
	vt_v   = &ValType{vt_void, confs.GetString(conf_group, conf_valtype, "void", confs.Locale())}
	vt_d   = &ValType{vt_digital, confs.GetString(conf_group, conf_valtype, "digital", confs.Locale())}
	vt_i32 = &ValType{vt_int32, confs.GetString(conf_group, conf_valtype, "int32", confs.Locale())}
	vt_f32 = &ValType{vt_float32, confs.GetString(conf_group, conf_valtype, "float32", confs.Locale())}
	vt_i64 = &ValType{vt_int64, confs.GetString(conf_group, conf_valtype, "int64", confs.Locale())}
	vt_f64 = &ValType{vt_float64, confs.GetString(conf_group, conf_valtype, "float64", confs.Locale())}
	vt_s   = &ValType{vt_string, confs.GetString(conf_group, conf_valtype, "string", confs.Locale())}
	vt_b   = &ValType{vt_blob, confs.GetString(conf_group, conf_valtype, "blob", confs.Locale())}
)

func GetValType(t int) *ValType {
	switch t {
	case vt_void:
		return vt_v
	case vt_digital:
		return vt_d
	case vt_int32:
		return vt_i32
	case vt_float32:
		return vt_f32
	case vt_int64:
		return vt_i64
	case vt_float64:
		return vt_f64
	case vt_string:
		return vt_s
	case vt_blob:
		return vt_b
	default:
		return vt_f32
	}
}

const conf_defaultFmt = "defaultFmt"

var default_fmt = confs.GetString(conf_group, conf_defaultFmt)

func GetDefaultFmt() string {
	return default_fmt
}

func NewPointValue(point *Point, sec int32, val interface{}, quality, err int16) PointValue {
	pv := PointValue{}
	pv.err = err
	pv.quality = quality
	pv.point = point
	pv.sec = sec
	pv.val = val
	return pv
}

func NewHistoryVals(vals *[]PointValue) HistoryVals {
	hvals := HistoryVals{}
	hvals.Vals = vals
	return hvals
}

func (this Point) String() string {
	return fmt.Sprint("ID:", this.Id, " Name:", this.Name, " Desc:", this.Desc, " Unit:", this.Unit, " Type:", this.Type.Name, " ValType:", this.ValType.Name,
		" DB:", this.DB, " DBName:", this.DBName, " Fmt:", this.DefaultFmt)
}

func (this PointValue) String() string {
	return fmt.Sprint("Name:", this.point.Name, " Time:", this.GetTimeString(), " Val:", this.val, " Quality:", this.quality, " Err:", this.err)
}
