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
	destFolder := path.Join(cwd, "_dest")
	images, count := exifdata.ListPhotos(path.Join(cwd, "_test"))
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
		utils.Progress(count, i + 1)
	}
}
