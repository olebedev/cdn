package cdn

import (
	"regexp"
	"strconv"
)

// it in list
func in(list []string, a string) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}
	return -1
}

// parse given params, example: '100x100' | '100'
func parseParams(s string) ([]int, error) {
	res := make([]int, 2)
	var err error
	re := regexp.MustCompile("\\d+")
	spl := re.FindAllString(s, 2)

	if len(spl) == 0 {
		return nil, err
	}

	for i, item := range spl {
		v, e := strconv.Atoi(item)
		if e != nil {
			err = e
			continue
		}
		res[i] = v
	}

	if len(spl) == 1 {
		res[1] = res[0]
	}

	return res, err
}
