package points

import (
	"io/ioutil"
	"os"
	"encoding/json"
	"strings"
	"log"
	"github.com/qjsoftcn/gutils"
)

func StorePoints(dir, dbName string, ps []Point) {
	filePath := dir + "/" + strings.ToLower(dbName) + ".json"
	if gutils.PathExists(filePath) {
		os.Remove(filePath)
	}
	jcs, err := json.Marshal(ps)
	if err == nil {
		ioutil.WriteFile(filePath, []byte(jcs), 0777)
	}
}

func LoadPoints(dir, dbName string) []Point {
	filePath := dir + "/" + strings.ToLower(dbName) + ".json"
	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}

	var ps []Point
	err = json.Unmarshal(bs, &ps)
	if err != nil {
		log.Println(err)
	}
	return ps
}
