package sim

import (
	"strconv"
	"strings"
)

const (
	Slash = '/'
	At    = '@'
	Sem   = ';'
	Colon = ':'
)

//parse conn url
func parseConnURL(connStr string) (
	hosts []string, ports []int, users []string, pwds []string) {

	sl := len(connStr)

	if sl == 0 {
		return hosts, ports, users, pwds
	}

	ss := strings.Split(connStr, ";")

	size := len(ss)

	hosts = make([]string, size)
	ports = make([]int, size)
	users = make([]string, size)
	pwds = make([]string, size)

	startIndex := 0
	aIndex := 0

	for index, r := range connStr {
		switch r {
		case Slash:
			users[aIndex] = connStr[startIndex:index]
			startIndex = index + 1
		case At:
			pwds[aIndex] = connStr[startIndex:index]
			startIndex = index + 1
		case Colon:
			hosts[aIndex] = connStr[startIndex:index]
			startIndex = index + 1
		case Sem:
			ports[aIndex], _ = strconv.Atoi(connStr[startIndex:index])
			startIndex = index + 1
			aIndex++
		default:
			continue
		}
	}

	if startIndex < sl {
		ports[aIndex], _ = strconv.Atoi(connStr[startIndex:sl])
	}

	return hosts, ports, users, pwds

}
