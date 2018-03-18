package points

import (
	"strings"
	"strconv"
)

type PVCalcMethod string

const (
	PVCM_Value  = "val"
	PVCM_Time   = "time"
	PVCM_Max    = "max"
	PVCM_Min    = "min"
	PVCM_Avg    = "avg"
	PVCM_Total1 = "total1"
	PVCM_Total2 = "total2"
	PVCM_Diff   = "diff"
	PVCM_Avg2   = "avg2"
)

type PointVar struct {
	PointName  string
	CalcMethod *string
	InterWay   InterWay
}

func (this *PointVar) Var() string {
	return "'" + this.String() + "'"
}

func (this PointVar) String() string {
	if this.CalcMethod == nil {
		return this.PointName
	} else {
		return this.PointName + "." + *this.CalcMethod
	}
}

func (this *PointVar) GetMethod() string {
	if this.CalcMethod == nil {
		return PVCM_Value
	} else {
		return *this.CalcMethod
	}
}

func (this *PointVar) GetInterWay() InterWay {
	if this.InterWay == 0 {
		//default is accurate
		return IW_Accurate
	} else {
		return this.InterWay
	}
}
func isValidMethod(cm string) bool {
	switch cm {
	case PVCM_Avg, PVCM_Avg2, PVCM_Diff, PVCM_Max, PVCM_Min, PVCM_Time,
		PVCM_Total1, PVCM_Total2, PVCM_Value:
		return true
	default:
		return false
	}
}

func ParseVar(pvar string) *PointVar {
	if pvar[0] == '\'' {
		pvar = pvar[1:]
	}
	pl := len(pvar)

	if pvar[pl-1] == '\'' {
		pvar = pvar[:pl-1]
	}

	v := new(PointVar)
	ldi := strings.LastIndex(pvar, ".")
	if ldi == -1 {
		v.PointName = pvar
	} else {
		cm := pvar[ldi+1:]
		fdi := strings.Index(cm, ",")
		m := cm
		if fdi != -1 {
			m = cm[:fdi]
		}

		if isValidMethod(m) {
			v.PointName = pvar[:ldi]
			v.CalcMethod = &m

			if fdi != -1 {
				iwStr := cm[fdi+1:]
				iw, err := strconv.Atoi(iwStr)
				if err == nil {
					v.InterWay = InterWay(iw)
				}
			}

		} else {
			v.PointName = pvar
		}
	}

	return v

}
