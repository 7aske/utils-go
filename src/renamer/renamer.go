package main

import (
	"./exifdata"
	"./utils"
	"fmt"
	"os"
	"path"
)

func main() {
	cwd, _ := os.Getwd()
	destFolder := path.Join(cwd)
	srcFolder := path.Join(cwd, "_INGEST")
	if utils.Contains("-f", os.Args) {
		arg := utils.ParseArgument("-f", os.Args)
		if path.IsAbs(arg) {
			srcFolder = arg
		} else {
			srcFolder = path.Join(cwd, arg)
		}
	}
	if utils.Contains("-d", os.Args) {
		arg := utils.ParseArgument("-d", os.Args)
		if path.IsAbs(arg) {
			destFolder = arg
		} else {
			destFolder = path.Join(cwd, arg)
		}
	}
	images, count := exifdata.ListPhotos(srcFolder)
	fmt.Println(count)
	exifData := exifdata.GetExifData(&images)
	for i, img := range exifData {
		destPath := path.Join(destFolder, img.Rel())
		err := os.MkdirAll(destPath, 0775)
		if err != nil {
			fmt.Println("error creating path for", img.Name())
			continue
		}
		err = utils.CopyFile(img.Abs(), path.Join(destPath, img.Name()))
		if err != nil {
			fmt.Println("error copying", img.Name())
			continue
		}
		utils.Progress(count, i+1)
	}
}
