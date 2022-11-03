package trace

import (
	"fmt"
	"path"
	"runtime"
)

type Caller struct {
	FunctionName string
	FileName     string
	Line         int
}

func Call(skip int) Caller {
	type monitor struct {
		Alloc,
		TotalAlloc,
		Sys,
		Mallocs,
		Frees,
		LiveObjects,
		PauseTotalNs uint64
		NumGC        uint32
		NumGoroutine int
	}
	var m monitor
	m.NumGoroutine = runtime.NumGoroutine()
	fileName, funcName, line := f1(skip)
	return Caller{
		FunctionName: funcName,
		FileName:     fileName,
		Line:         line,
	}
}

func getLocation(skip int) (fileName, funcName string, line int) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		fmt.Println("caller get info failed")
		return
	}
	fileName = path.Base(file)
	funcName = runtime.FuncForPC(pc).Name()
	return
}
func f1(skip int) (fileName, funcName string, line int) {
	fileName, funcName, line = getLocation(skip)
	return
}
