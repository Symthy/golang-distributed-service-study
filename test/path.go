package test

import (
	"path/filepath"
	"runtime"
)

func GetRootPath() string {
	_, b, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(b), "../")
}
