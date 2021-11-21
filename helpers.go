package main

import "path/filepath"

func stripExtensionFromFileBaseName(baseName string) string {
	fileExtension := filepath.Ext(baseName)
	return baseName[0: len(baseName) - len(fileExtension)]
}

func getOutputDirectoryFromTmpFileName(tmpFileName string) string {
	baseName := filepath.Base(tmpFileName)
	return stripExtensionFromFileBaseName(baseName)
}


