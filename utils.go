package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

func savePid() {
	if withPid := os.Getenv("WITH_PID"); len(withPid) > 0 {
		fmt.Printf("Start with pid %d.\n", os.Getpid())
		outfile, err := os.Create(os.Getenv("WITH_PID"))
		if err != nil {
			panic(err)
		}
		outfile.Write([]byte(fmt.Sprintf("%d", os.Getpid())))
		outfile.Close()
	}
}

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
func parseParams(s string) []int {
	res := make([]int, 2)
	re := regexp.MustCompile("\\d+")
	spl := re.FindAllString(s, 2)

	if len(spl) == 0 {
		return nil
	}

	for i, item := range spl {
		v, err := strconv.Atoi(item)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		res[i] = v
	}

	if len(spl) == 1 {
		res[1] = res[0]
	}

	return res
}
