package handler

import (
	"math"
	"path/filepath"
)

func stripExtensionFromFileBaseName(baseName string) string {
	fileExtension := filepath.Ext(baseName)
	return baseName[0: len(baseName) - len(fileExtension)]
}

func getOutputDirectoryFromTmpFileName(tmpFileName string) string {
	baseName := filepath.Base(tmpFileName)
	return stripExtensionFromFileBaseName(baseName)
}

func bytesToMb(size int64) float64 {
	return float64(size) / math.Pow(10, 6)
}