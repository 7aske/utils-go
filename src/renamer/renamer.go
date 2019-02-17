package main

import (
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ExifData struct {
	mk    string
	model string
	date  string
	tm    string
	ext   string
	p     string
	fp    string
}

func (d *ExifData) String() string {
	return fmt.Sprintf(
		"Make: %s\n"+
			"Model: %s\n"+
			"Date: %s\n"+
			"Time: %s\n"+
			"Ext: %s\n"+
			"Path: %s\n"+
			"FPath: %s\n",
		d.mk,
		d.model,
		d.date,
		d.tm,
		d.ext,
		d.p,
		d.fp)
}

var delAfterCopy = false

func main() {
	cwd, _ := os.Getwd()
	destFolder := path.Join(cwd, "_dest")
	images, count := listPhotos(path.Join(cwd, "_test"))
	fmt.Println(count)
	exifData := getExifData(&images)
	for _, f := range exifData {
		err := os.MkdirAll(path.Join(destFolder, f.mk, f.model), 0775)
		if err != nil {
			fmt.Println(err)
			continue
		}
		destFile := path.Join(destFolder, f.mk, f.model, f.fp)
		err = moveFile(f.p, destFile, delAfterCopy)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func getExifData(images *[]string) []ExifData {
	var data []ExifData

	for _, img := range *images {
		exifData := parseExifData(img)
		data = append(data, exifData)
	}
	return data
}

func parseExifData(p string) ExifData {
	var (
		model    string
		mk       string
		datetime string
		date     string
		tm       string
		ext      string
	)
	f, err := os.Open(p)
	if err != nil {
		fmt.Println("error opening " + p)
	}
	x, err := exif.Decode(f)
	xmodel, err := x.Get(exif.Model)
	if err != nil {
		model = "UNK-MODEL"
	} else {
		model, err = xmodel.StringVal()
		if err != nil {
			model = "UNK-MODEL"
		}
	}
	model = strings.ToUpper(model)
	switch model {
	case "FINEPIX X100":
		model = "X100"
	case "X-E1":
		model = "XE1"
	case "NIKON D3200":
		model = "D3200"
	case "MI A1":
		model = "MIA1"
	}
	xmk, err := x.Get(exif.Make)
	if err != nil {
		mk = "UNK-MAKE"
	} else {
		mk, err = xmk.StringVal()
		if err != nil {
			mk = "UNK-MAKE"
		}
	}
	mk = strings.ToUpper(mk)
	switch mk {
	case "NIKON CORPORATION":
		mk = "NIKON"
	}
	var xdatetime *tiff.Tag
	if xdatetime, err = x.Get(exif.DateTimeDigitized); err != nil {
		xdatetime, err = x.Get(exif.DateTime)
	}
	if err != nil {
		now := time.Now()
		datetime = fmt.Sprintf("%d:%s:%s %s:%s:%s", now.Year(), getMonth(now), addZero(now.Day()), addZero(now.Hour()), addZero(now.Minute()), addZero(now.Second()))
	} else {
		datetime, err = xdatetime.StringVal()
		if err != nil {
			now := time.Now()
			datetime = fmt.Sprintf("%d:%s:%s %s:%s:%s", now.Year(), getMonth(now), addZero(now.Day()), addZero(now.Hour()), addZero(now.Minute()), addZero(now.Second()))
		}
	}
	date = strings.Replace(strings.Split(datetime, " ")[0], ":", "-", -1)
	tm = strings.Replace(strings.Split(datetime, " ")[1], ":", "-", -1)
	ext = strings.ToUpper(filepath.Ext(p))
	fp := fmt.Sprintf("%s_%s_%s%s", date, tm, model, ext)
	return ExifData{mk, model, date, tm, ext, p, fp}
}

func listPhotos(p string) ([]string, int) {
	validExtensions := []string{".JPEG", ".JPG", ".NEF", ".RAF"}
	dir, _ := ioutil.ReadDir(p)
	var images []string
	var count = 0
	for _, file := range dir {
		ext := strings.ToUpper(filepath.Ext(file.Name()))
		if contains(ext, validExtensions) {
			images = append(images, path.Join(p, file.Name()))
			count++
		}
	}
	return images, count
}
func moveFile(src string, dest string, mv bool) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		fmt.Println(err)
	}
	dest = formatFilename(dest, 0)
	err = ioutil.WriteFile(dest, input, 0775)
	if err != nil {
		return err
	}
	if mv {
		err := os.Remove(src)
		if err != nil {
			return err
		}
	}
	return nil
}
func formatFilename(f string, c int) string {
	dir, filename := path.Split(f)
	ext := path.Ext(filename)
	var newFilename = f
	if c != 0 {
		if c == 1 {
			newFilename = path.Join(dir, filename[:len(filename)-len(ext)]+"-"+strconv.Itoa(c)+ext)
		} else if c < 11 {
			newFilename = path.Join(dir, filename[:len(filename)-len(ext)-2]+"-"+strconv.Itoa(c)+ext)
		} else {
			newFilename = path.Join(dir, filename[:len(filename)-len(ext)-3]+"-"+strconv.Itoa(c)+ext)
		}
	}
	c++
	if _, err := os.Stat(newFilename); err == nil {
		newFilename = formatFilename(newFilename, c)
	}
	return newFilename
}
func contains(search string, slice []string) bool {
	for _, s := range slice {
		if search == s {
			return true
		}
	}
	return false
}
func getMonth(date time.Time) string {
	switch date.Month() {
	case time.January:
		return "01"
	case time.February:
		return "02"
	case time.March:
		return "03"
	case time.April:
		return "04"
	case time.May:
		return "05"
	case time.June:
		return "06"
	case time.July:
		return "07"
	case time.August:
		return "08"
	case time.September:
		return "09"
	case time.October:
		return "10"
	case time.November:
		return "11"
	case time.December:
		return "12"
	}
	return "00"
}

func addZero(num int) string {
	if num < 10 {
		return "0" + strconv.Itoa(num)
	} else {
		return strconv.Itoa(num)
	}

}
