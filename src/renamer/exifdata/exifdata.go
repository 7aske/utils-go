package exifdata

import (
	"../dateutils"
	"../utils"
	"fmt"
	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	dt "time"
)

type ExifData struct {
	maker string
	model string
	date  string
	time  string
	ext   string
	name  string
	abs   string
	rel   string
}

func (d *ExifData) Date() string  { return d.date }
func (d *ExifData) Make() string  { return d.maker }
func (d *ExifData) Model() string { return d.model }
func (d *ExifData) Time() string  { return d.time }
func (d *ExifData) Ext() string   { return d.ext }
func (d *ExifData) Rel() string   { return d.rel }
func (d *ExifData) Abs() string   { return d.abs }
func (d *ExifData) Name() string  { return d.name }
func (d *ExifData) String() string {
	return fmt.Sprintf(
		"Make: %s\n"+
			"Model: %s\n"+
			"Date: %s\n"+
			"Time: %s\n"+
			"Ext: %s\n"+
			"Path: %s\n"+
			"FPath: %s\n",
		d.maker,
		d.model,
		d.date,
		d.time,
		d.ext,
		d.name,
		d.abs)
}
func GetExifData(images *[]string) []ExifData {
	var data []ExifData

	for _, img := range *images {
		if exifData, ok := parseExifData(img); ok {
			data = append(data, exifData)
		}
	}
	return data
}

func parseExifData(p string) (ExifData, bool) {
	var (
		model    string
		maker    string
		datetime string
		date     string
		time     string
		ext      string
	)
	f, err := os.Open(p)
	if err != nil {
		fmt.Println("error opening " + p)
		return ExifData{}, false
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
	case "FINEPIX X100S":
		model = "X100S"
	case "X-E1":
		model = "XE1"
	case "NIKON D3200":
		model = "D3200"
	case "MI A1":
		model = "MIA1"
	}
	xmk, err := x.Get(exif.Make)
	if err != nil {
		maker = "UNK-MAKE"
	} else {
		maker, err = xmk.StringVal()
		if err != nil {
			maker = "UNK-MAKE"
		}
	}
	maker = strings.ToUpper(maker)
	switch maker {
	case "NIKON CORPORATION":
		maker = "NIKON"
	case "SONY-ERICSSON":
		maker = "SONY"
	}

	var xdatetime *tiff.Tag
	if xdatetime, err = x.Get(exif.DateTimeDigitized); err != nil {
		xdatetime, err = x.Get(exif.DateTime)
	}
	if err != nil {
		datetime = dateutils.DateToString(dt.Now())
	} else {
		datetime, err = xdatetime.StringVal()
		if err != nil {
			datetime = dateutils.DateToString(dt.Now())
		}
	}
	date = strings.Replace(strings.Split(datetime, " ")[0], ":", "-", -1)
	time = strings.Replace(strings.Split(datetime, " ")[1], ":", "-", -1)
	ext = strings.ToUpper(filepath.Ext(p))
	name := fmt.Sprintf("%s_%s_%s%s", date, time, model, ext)
	//img.Make(), img.Model(), fmt.Sprintf("%s_%s", img.Date(), img.Model()
	rel := path.Join(maker, model, fmt.Sprintf("%s_%s", date, model))
	return ExifData{maker, model, date, time, ext, name, p, rel}, true
}

func ListPhotos(p string) ([]string, int) {
	validExtensions := []string{".JPEG", ".JPG", ".NEF", ".RAF"}
	dir, _ := ioutil.ReadDir(p)
	var images []string
	var count = 0
	for _, file := range dir {
		ext := strings.ToUpper(filepath.Ext(file.Name()))
		if utils.Contains(ext, validExtensions) {
			images = append(images, path.Join(p, file.Name()))
			count++
		}
	}
	return images, count
}
