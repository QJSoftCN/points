package points

import (
	"time"
	"errors"
	"math"
)

type HistorysVisitor interface {
	Visit(index int, val *PointValue)
}

type SecVisitor struct {
	MaxInterval  int32
	MinInterVal  int32
	inited       bool
	visitedCount int
	upper        *PointValue
	tiSum        int32
	tiCount      int
}

func (this *SecVisitor) Visit(index int, val *PointValue) {
	this.visitedCount++
	if this.upper == nil {
		this.upper = val
		return
	}
	ti := val.sec - this.upper.sec
	if !this.inited {
		this.MinInterVal = ti
		this.MaxInterval = ti
	}
	if ti > 0 {
		if this.MaxInterval < ti {
			this.MaxInterval = ti
		}
		if this.MinInterVal > ti {
			this.MinInterVal = ti
		}
		this.tiSum += ti
		this.tiCount++
	}
}

func (this *SecVisitor) Long() (int32, error) {
	if this.visitedCount == 0 {
		return 0, errors.New("no data")
	}
	return this.MaxInterval, nil
}

func (this *SecVisitor) Short() (int32, error) {
	if this.visitedCount == 0 {
		return 0, errors.New("no data")
	}
	return this.MinInterVal, nil
}

func (this *SecVisitor) AvgInterval() (int32, error) {
	if this.visitedCount == 0 {
		return 0, errors.New("no data")
	}
	bs := this.tiSum / int32(this.tiCount)
	ys := this.tiSum % int32(this.tiCount)
	if ys > 0 {
		return bs + 1, nil
	} else {
		return bs, nil
	}
}

type DigitalCalculator struct {
	count       int
	Digital     int32
	ActionTimes int
	records     []DigitalRecord
	curRecord   *DigitalRecord
}

type DigitalRecord struct {
	Start        *PointValue
	End          *PointValue
	IsIncluedEnd bool
}

func (this *DigitalRecord) Secs() int32 {
	secs := this.End.sec - this.Start.sec
	if this.IsIncluedEnd {
		secs += 1
	}
	return secs
}

func NewDigitalCalculator() DigitalCalculator {
	dc := new(DigitalCalculator)
	dc.records = make([]DigitalRecord, 0)
	return *dc
}
func (this *DigitalCalculator) GetRecords() []DigitalRecord {
	if this.curRecord != nil {
		this.curRecord.IsIncluedEnd = true
		this.records = append(this.records, *this.curRecord)
		this.curRecord = nil
	}
	return this.records
}
func (this *DigitalCalculator) TotalSecs() int32 {
	var tss int32 = 0
	for _, dr := range this.GetRecords() {
		tss += dr.End.sec - dr.Start.sec
		if dr.IsIncluedEnd {
			tss += 1
		}
	}
	return tss
}

func (this *DigitalCalculator) Peek(upper, the *PointValue) {
	//action times
	this.ActionTimes += DigitalAction(this.Digital, upper, the)
	//secs
	curDig := the.IntVal()
	if this.curRecord == nil {
		if curDig == this.Digital {
			this.curRecord = new(DigitalRecord)
			this.curRecord.Start = the
			this.curRecord.End = the
		}
	} else {
		if curDig == this.Digital {
			this.curRecord.End = the
		} else {
			this.curRecord.End = the
			this.records = append(this.records, *this.curRecord)
			this.curRecord = nil
		}
	}
}

func DigitalAction(digital int32, upper, the *PointValue) int {
	curDig := the.IntVal()
	if upper.IntVal() == curDig {
		return 0
	} else {
		if curDig == digital {
			return 1
		} else {
			return 0
		}
	}
}

type DigitalVisitor struct {
	visitedCount int
	upper        *PointValue
	dcMap        map[int32]DigitalCalculator
}

func NewDigitalVisitor() DigitalVisitor {
	dv := DigitalVisitor{}
	dcMap := make(map[int32]DigitalCalculator)
	dv.dcMap = dcMap
	return dv
}

func (this *DigitalVisitor) Visit(index int, val *PointValue) {
	this.visitedCount++

	if this.upper == nil {
		this.upper = val
	}

	dig := val.IntVal()

	_, ok := this.dcMap[dig]
	if !ok {
		dc := NewDigitalCalculator()
		dc.Digital = dig
		this.dcMap[dig] = dc
	}

	for _, c := range this.dcMap {
		c.Peek(this.upper, val)
	}
}

func (this *DigitalVisitor) ActionTimes(digital int32) int {
	dc, ok := this.dcMap[digital]
	if ok {
		return dc.ActionTimes
	} else {
		return 0
	}
}

func (this *DigitalVisitor) TotalSecs(digital int32) int32 {
	dc, ok := this.dcMap[digital]
	if ok {
		return dc.TotalSecs()
	} else {
		return 0
	}
}

func (this *DigitalVisitor) GetRecords(digital int32) []DigitalRecord {
	dc, ok := this.dcMap[digital]
	if ok {
		return dc.GetRecords()
	} else {
		return nil
	}
}

func (this *DigitalVisitor) Digitals() []int32 {
	dcMLen := len(this.dcMap)
	if dcMLen == 0 {
		return nil
	}

	ds := make([]int32, dcMLen)
	i := 0
	for d, _ := range this.dcMap {
		ds[i] = d
	}
	return ds
}

//analog supports avg avg2 max/maxtime min/mintime
type AnalogVisitor struct {
	visitedCount int

	firstVal *PointValue
	lastVal  *PointValue
	maxVal   *PointValue
	minVal   *PointValue
	//avg2
	valSum float64
	//avg
	upper    *PointValue
	valTWSum float64
	secs     time.Duration
	goodSecs time.Duration
	//total2
	sgSegs []SGSegment
	curSeg *SGSegment
}

func NewAnalogVistor() AnalogVisitor {
	av := AnalogVisitor{}
	sgs := make([]SGSegment, 0)
	av.sgSegs = sgs
	return av
}

//total2 segment
type SGSegment struct {
	Start *PointValue
	End   *PointValue
	reSet *SGSReset
}

//total2 reset point
type SGSReset struct {
	Before *PointValue
	The    *PointValue
	After  *PointValue
}

func (this *AnalogVisitor) Avg() (float64, error) {
	switch {
	case this.visitedCount <= 0:
		return 0, errors.New("no data")
	case this.visitedCount == 1:
		return this.upper.FloatVal(), nil
	default:
		return this.valTWSum / float64(this.goodSecs), nil
	}
}

func (this *AnalogVisitor) Avg2() (float64, error) {
	if this.visitedCount <= 0 {
		return 0, errors.New("no data")
	}

	return this.valSum / float64(this.visitedCount), nil
}

func (this *AnalogVisitor) Total1() (float64, error) {
	avg, err := this.Avg()
	if err != nil {
		return 0, err
	}
	return avg * (float64(this.secs) / 3600), nil
}

func (this *AnalogVisitor) Total2() (float64, error) {
	if this.visitedCount <= 0 {
		return 0, errors.New("no data")
	}

	total2 := 0.0
	for _, seg := range this.sgSegs {
		total2 += seg.End.FloatVal() - seg.Start.FloatVal()
	}

	if this.curSeg != nil {
		total2 += this.curSeg.End.FloatVal() - this.curSeg.Start.FloatVal()
	}

	return total2, nil
}

func (this *AnalogVisitor) Max() (*PointValue, error) {
	if this.visitedCount <= 0 {
		return nil, errors.New("no data")
	}
	return this.maxVal, nil
}

func (this *AnalogVisitor) Min() (*PointValue, error) {
	if this.visitedCount <= 0 {
		return nil, errors.New("no data")
	}
	return this.minVal, nil
}

func (this *AnalogVisitor) First() (*PointValue, error) {
	if this.visitedCount <= 0 {
		return nil, errors.New("no data")
	}
	return this.firstVal, nil
}

func (this *AnalogVisitor) Last() (*PointValue, error) {
	if this.visitedCount <= 0 {
		return nil, errors.New("no data")
	}
	return this.lastVal, nil
}

func (this *AnalogVisitor) calcMax(val *PointValue) {
	if this.maxVal == nil {
		this.maxVal = val
	} else {
		if this.maxVal.FloatVal() < val.FloatVal() {
			this.maxVal = val
		}
	}
}

func (this *AnalogVisitor) calcMin(val *PointValue) {
	if this.minVal == nil {
		this.minVal = val
	} else {
		if this.minVal.FloatVal() > val.FloatVal() {
			this.minVal = val
		}
	}
}

func (this *AnalogVisitor) calcAvg2(val *PointValue) {
	this.valSum += val.FloatVal()
}

func (this *AnalogVisitor) calcAvg(val *PointValue) {
	if this.upper == nil {
		this.upper = val
	}

	this.secs += time.Duration(val.sec - this.upper.sec)
	this.goodSecs += calcGoodSecs(this.upper, val)
	this.valTWSum += calcMid(this.upper, val)

	this.upper = val
}

func IsReset(befoe, the, after *PointValue) bool {

	if after.FloatVal() >= befoe.FloatVal() {
		return false
	}

	abChange := after.FloatVal() - befoe.FloatVal()
	abRate := math.Abs(abChange / befoe.FloatVal())
	if abRate <= 0.1 {
		//not more than 10%
		return false
	}

	tbChange := the.FloatVal() - befoe.FloatVal()
	tbRate := math.Abs(tbChange / befoe.FloatVal())
	if tbRate <= 0.1 {
		//not more than 10%
		return false
	}

	return true
}

func (this *AnalogVisitor) calcTotal2(val *PointValue) {
	if this.curSeg == nil {
		this.curSeg = new(SGSegment)
		this.curSeg.Start = val
		this.curSeg.End = val
	} else {
		if this.curSeg.reSet == nil {
			if this.curSeg.End.FloatVal() <= val.FloatVal() {
				//keep growing
				this.curSeg.End = val
			} else {
				//judge reset point
				r := SGSReset{}
				r.Before = this.curSeg.End
				r.The = val

				this.curSeg.reSet = &r
			}
		} else {
			this.curSeg.reSet.After = val

			if IsReset(this.curSeg.reSet.Before, this.curSeg.reSet.The, this.curSeg.reSet.After) {
				//reset
				r := this.curSeg.reSet
				this.curSeg.reSet = nil
				this.sgSegs = append(this.sgSegs, *this.curSeg)

				this.curSeg = new(SGSegment)
				this.curSeg.Start = r.The
				this.curSeg.End = r.After

			} else {
				//not reset
				this.curSeg.End = val
				this.curSeg.reSet = nil
			}
		}

	}

}

func calcMid(upper, after *PointValue) float64 {
	secs := after.sec - upper.sec
	if upper.IsOK() {
		if after.IsOK() {
			return ((upper.FloatVal() + after.FloatVal()) / 2) * float64(secs)
		} else {
			return upper.FloatVal() * float64(secs)
		}
	} else {
		return 0
	}

}

func calcGoodSecs(upper, after *PointValue) time.Duration {
	if upper.IsOK() {
		return time.Duration(after.sec - upper.sec)
	}
	return 0
}

func (this *AnalogVisitor) Visit(index int, val *PointValue) {
	this.visitedCount++

	if this.firstVal == nil {
		this.firstVal = val
	}

	this.calcMax(val)
	this.calcMin(val)
	this.calcAvg2(val)
	this.calcAvg(val)
	this.calcTotal2(val)

	this.lastVal = val

}
