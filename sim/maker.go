package sim

import (
	"../std"
	"time"
)

func makeSnapshot(point *points.Point)points.PointValue{
	pf:=points.GetPointFeature(point.Name)
	points.NewPointValue(point,pf.MakeSec(time.Now()),pf.MakeValue(),0,0)
}
