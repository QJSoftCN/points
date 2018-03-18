package points

import (
	"time"
	"math/rand"
	"strings"
	"io/ioutil"
	"log"
	"encoding/json"
)

//point feature simulator
type PointFeatureSimulator interface {
	Val() interface{}
	IsNormal(val ...interface{}) bool
	Merge(pfs PointFeatureSimulator)
}

type SecSimulator struct {
	MaxTi int
	MinTi int
}

func (this *SecSimulator) Val() interface{} {
	return this.MinTi + rand.Intn(this.MaxTi-this.MinTi)
}

//para is time interval
func (this *SecSimulator) IsNormal(val ...interface{}) bool {
	ti := 0
	//this is ti
	sec, ok := val[0].(int32)
	if !ok {
		log.Println("sec simulator tell:", val[0], " not int32")
	}

	ti = int(sec)
	if ti >= this.MinTi && ti <= this.MaxTi {
		return true
	}
	return false
}

func (this *SecSimulator) Merge(simulator PointFeatureSimulator) {
	ss, ok := simulator.(*SecSimulator)
	if ok {
		if this.MaxTi < ss.MaxTi {
			this.MaxTi = ss.MaxTi
		}
		if this.MinTi > ss.MinTi {
			this.MinTi = ss.MinTi
		}
	}
}

type AnalogSimulator struct {
	Max float64 //real max
	Min float64 //real min
}

func (this *AnalogSimulator) Val() interface{} {
	return this.Min + (rand.Float64() * (this.Max - this.Min))
}

func (this *AnalogSimulator) IsNormal(val ...interface{}) bool {
	f, ok := val[0].(float64)
	if ok {
		return f >= this.Min && f <= this.Max
	} else {
		log.Println("ananlog simulator tell:", val, " not float64")
		return false
	}
}

func (this *AnalogSimulator) Merge(lawer PointFeatureSimulator) {
	avl, ok := lawer.(*AnalogSimulator)
	if ok {
		if (avl.Max > this.Max) {
			this.Max = avl.Max
		}

		if (avl.Min < this.Min) {
			this.Min = avl.Min
		}
	}
}

//digital simulator
type DigitalSimulator struct {
	//digital random dist array,must have random time law
	digs     []DigitalRate
	digRands []int32
}

type DigitalRate struct {
	Digital   int32
	Times     int
	TotalSecs int32
	rate      float64
}

func toDigRands(drs []DigitalRate) []int32 {
	var totalSecs int32
	times := 0
	for _, v := range drs {
		totalSecs += v.TotalSecs
		times += v.Times
	}

	for _, v := range drs {
		v.rate = float64(v.TotalSecs) / float64(totalSecs)
	}

	digRands := make([]int32, 0)
	for _, v := range drs {
		e := int(v.rate * 100)
		vds := make([]int32, e)
		for i := 0; i < e; i++ {
			vds[i] = v.Digital
		}
		digRands = append(digRands, vds...)
	}
	return digRands
}

func (this *DigitalSimulator) Val() interface{} {
	if this.digRands == nil {
		this.digRands = toDigRands(this.digs)
	}
	return this.digRands[rand.Intn(len(this.digRands))]
}

func (this *DigitalSimulator) IsNormal(val ...interface{}) bool {
	xd, ok := val[0].(int32)
	if ok {
		for _, d := range this.digs {
			if d.Digital == xd {
				return true
			}
		}
	}
	return false
}

func (this *DigitalRate) Merge(dr *DigitalRate) {
	this.TotalSecs += dr.TotalSecs
	this.Times += dr.Times
}

func (this DigitalSimulator) Get(d int32) *DigitalRate {
	for _, v := range this.digs {
		if v.Digital == d {
			return &v
		}
	}
	return nil
}

func (this *DigitalSimulator) Merge(simulator PointFeatureSimulator) {
	ds, ok := simulator.(*DigitalSimulator)
	if ok {
		for _, v := range ds.digs {
			dr := this.Get(v.Digital)
			if dr != nil {
				dr.Merge(&v)
			} else {
				this.digs = append(this.digs, v)
			}
		}
	}
}

/**
* monitoring and analysis std
 */
type HourFeature struct {
	Hour         int
	HisCount     int
	ValSimulator PointFeatureSimulator
	SecSimulator PointFeatureSimulator
	AvgTimeInterval float64
	MinTimeInterval int64
	MaxTimeInterval int64

}

func (hf *HourFeature) Sec(base int64) int64 {
	ti,_:=hf.SecSimulator.Val().(int32)
	return base + int64(ti)
}

func (hf *HourFeature) Val() interface{} {
	return hf.ValSimulator.Val()
}

// point features
type PointOneDayFeature struct {
	hfs map[int]HourFeature
}

func (pf *PointOneDayFeature) Make(base time.Time) (int32, interface{}, int16, int16) {
	h := base.Hour()
	hf, ok := pf.hfs[h]

	if ok {
		return int32(hf.Sec(base.Local().Unix())), hf.Val(), 0, 0
	} else {
		return int32(base.Unix() + int64(rand.Intn(10))), rand.Float64() * 1000, 0, 0
	}
}

var pfs map[string]*PointOneDayFeature

func GetPointFeature(name string) *PointOneDayFeature {
	name = strings.ToLower(name)
	pf, _ := pfs[name]
	return pf
}

func LoadPointFeatures(dbName string) {
	path := "ps/" + dbName + "_pfs.json"
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("read ", dbName, " pfs err:", err)
	}

	err = json.Unmarshal(bs, &pfs)
	if err != nil {
		log.Println("unmarshal ", dbName, " pfs json err:", err)
	}
}

func StorePointFeatures(dbName string, m *map[string]PointOneDayFeature) {
	path := "ps/" + dbName + "_pfs.json"
	bs, err := json.Marshal(m)
	if err != nil {
		log.Println("marshal ", dbName, " pfs json err:", err)
	}
	ioutil.WriteFile(path, bs, 0777)
}

func BuildPFS(db RealDB, ps []Point) map[string]*PointOneDayFeature {
	pf := make(map[string]*PointOneDayFeature)
	for index, p := range ps {
		go func() {
			pf[strings.ToLower(p.Name)] = BuildPF(db, &p)
			log.Println("No.", index, p.Name, " pf builded ok")
		}()
	}
	return pf
}

func BuildPF(db RealDB, p *Point) *PointOneDayFeature {
	//read ten days history analysis feature
	now := time.Now()

	end := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	start := time.Date(now.Year(), now.Month(), now.Day()-10, now.Hour(), 0, 0, 0, now.Location())

	m := make(map[int][]HourFeature)

	for {
		hf := BuildHF(db, p, start, start.Add(time.Hour-time.Second))
		if hf != nil {
			m[hf.Hour] = append(m[hf.Hour], *hf)
		}
		start = start.Add(time.Hour)
		if !start.Before(end) {
			break
		}
	}

	hfs := make(map[int]HourFeature)
	for key, v := range m {
		hfs[key] = toOneHF(key, &v)
	}

	podf := new(PointOneDayFeature)
	podf.hfs = hfs
	return podf
}

func toOneHF(hour int, hfs *[]HourFeature) HourFeature {
	hf := (*hfs)[0]

	atiSum := hf.AvgTimeInterval
	hcSum := hf.HisCount

	l := len(*hfs)
	vNum := 0
	for i := 1; i < l; i++ {
		v := (*hfs)[i]

		if v.HisCount == 0 {
			continue
		}

		//hf.ValLawer.Merge(v.ValLawer)

		if v.MinTimeInterval < hf.MinTimeInterval {
			hf.MinTimeInterval = v.MinTimeInterval
		}
		if v.MaxTimeInterval > hf.MaxTimeInterval {
			hf.MaxTimeInterval = v.MaxTimeInterval
		}

		atiSum += v.AvgTimeInterval
		hcSum += v.HisCount
		vNum++
	}

	//hf.AvgTimeInterval = atiSum / vNum
	hf.HisCount = hcSum / vNum

	return hf
}

func BuildHF(db RealDB, p *Point, start, end time.Time) *HourFeature {
	s := toSecs(start)
	e := toSecs(end)
	hisVal, err := db.ReadHistory(p.Name, s, e, HCW_Demo)
	if err != nil {
		log.Println("BuildHF read histroy err:", err)
		return nil
	}
	switch p.ValType.Id {
	case vt_float32, vt_float64:
		av := NewAnalogVistor()
		hisVal.Accept(&av)
	case vt_digital:
		dv := NewDigitalVisitor()
		hisVal.Accept(&dv)
	case vt_int32, vt_int64:
		av := NewAnalogVistor()
		dv := NewDigitalVisitor()
		hisVal.Accept(&av, &dv)
	default:
		return nil
	}

	//avl := AnalogValLawer{}

	hf := HourFeature{}
	hf.Hour = start.Hour()
	//hf.ValLawer = &avl

	return nil
}
