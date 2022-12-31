package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRootPath(t *testing.T) {
	rootPath := GetRootPath()
	dir, _ := os.Getwd()
	assert.Equal(t, filepath.Join(dir, "../"), rootPath)
}
