package mslice

import (
	"math/rand"
	"reflect"
	"strconv"
	"time"
)

// MakeRangeSlice [start,end)
func MakeRangeSlice(start, end int) []int {
	lens := end - start
	if lens < 0 {
		return []int{}
	}
	list := make([]int, 0, lens)
	for i := start; i < end; i++ {
		list = append(list, i)
	}
	return list
}

// MakeRangeSliceStr [start,end)
func MakeRangeSliceStr(start, end int) []string {
	lens := end - start
	if lens < 0 {
		return []string{}
	}
	list := make([]string, 0, lens)
	for i := start; i < end; i++ {
		list = append(list, strconv.Itoa(i))
	}
	return list
}

// MakeRandRangeSlice [start,end)
func MakeRandRangeSlice(start, end int) []int {
	lens := end - start
	if lens < 0 {
		return []int{}
	}
	list := make([]int, 0, lens)
	for i := start; i < end; i++ {
		list = append(list, i)
	}
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(list) - 1; i > 0; i-- {
		randNum := rd.Intn(i)
		list[i], list[randNum] = list[randNum], list[i]
	}
	return list
}

func RandSlice(x any) {
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	of := reflect.ValueOf(x)
	rd.Shuffle(of.Len(), func(i, j int) {
		tmp := reflect.New(of.Index(i).Type()).Elem()
		tmp.Set(of.Index(i))
		of.Index(i).Set(of.Index(j))
		of.Index(j).Set(tmp)
	})
}
