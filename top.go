package points

import "time"

/**
	topDB is common used RealDB
	topDB must be inited at application starting
	topDB can use "Open()" method init
	et:
		dbName := confs.GetString("top", "RealDB", "Name")
		connUrl := confs.GetString("top", "RealDB", "Url")
		size := confs.GetInt("top", "RealDB", "Size")

		db, err := points.Open(dbName, connUrl, size)
		if err != nil {
			//record err,tell user,use sim replace.
			//this is a good method
			panic(err)
		} else {
			points.SetTopDB(db)
		}
 */

var topDB RealDB
func SetTopDB(opendDB RealDB) {
	topDB = opendDB
}

func ReadHistory(name string, start, end time.Time, way HistorysComplementWay) (*HistoryVals, error) {
	return topDB.ReadHistory(name, toSecs(start), toSecs(end), way)
}

func GetPoint(name string) *Point {
	return topDB.GetPoint(name)
}

func GetPoints(name, desc string, pts, vts []int) []Point {
	return topDB.GetPoints(name, desc, pts, vts)
}

func ReadSnapshot(name string) (*PointValue, error) {
	return topDB.ReadSnapshot(name)
}

func ReadSnapshots(names []string) ([]PointValue, error) {
	return topDB.ReadSnapshots(names)
}

func toSecs(t time.Time) int32 {
	nt := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	return int32(nt.Unix())
}

func InterVal(name string, t time.Time, way InterWay) (*PointValue, error) {
	return topDB.InterVal(name, toSecs(t), way)
}

func InterVals(name []string, t time.Time, way InterWay) ([]PointValue, error) {
	return topDB.InterVals(name, toSecs(t), way)
}
