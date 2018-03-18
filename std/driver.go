package points

type InterWay uint

const (
	IW_Not      = iota
	IW_Before
	IW_After
	IW_Accurate
)

type HistorysComplementWay struct {
	Start    bool
	End      bool
	StartWay InterWay
	EndWay   InterWay
}

var HCW_Demo = NewHCW(false, false, IW_Not, IW_Not)

func NewHCW(s, e bool, sw, ew InterWay) HistorysComplementWay {
	return HistorysComplementWay{Start: s, End: e, StartWay: sw, EndWay: sw}
}

type RealDB interface {
	GetDBName() string
	GetConnUrl() string
	GetConnSize() int

	Close() error

	GetPoint(name string) *Point
	GetPoints(name, desc string, pts, vts []int) []Point

	ReadSnapshot(name string) (*PointValue, error)
	ReadSnapshots(name []string) ([]PointValue, error)
	ReadHistory(name string, start, end int32, way HistorysComplementWay) (*HistoryVals, error)
	InterVal(name string, t int32, way InterWay) (*PointValue, error)
	InterVals(name []string, t int32, way InterWay) ([]PointValue, error)
	ReadPoints() ([]Point, error)
}

type RealDBDriver interface {
	Connect(url string, size int) (RealDB, error)
}
