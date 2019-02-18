package utils

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

func Contains(search string, slice []string) bool {
	for _, s := range slice {
		if search == s {
			return true
		}
	}
	return false
}
func CopyFile(src string, dest string) error {
	dest = FormatFilename(dest, 0)
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dest)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		_ = d.Close()
		return err
	}
	return d.Close()
}
func FormatFilename(f string, c int) string {
	if _, err := os.Stat(f); err != nil {
		return f
	} else {
		var newFilename = f
		dir, filename := path.Split(f)
		ext := path.Ext(filename)
		if c == 0 {
			newFilename = path.Join(dir, filename[:len(filename)-len(ext)]+"-"+strconv.Itoa(c)+ext)
		} else {
			pad := len(strconv.Itoa(c)) + 1
			newFilename = path.Join(dir, filename[:len(filename)-len(ext)-pad]+"-"+strconv.Itoa(c)+ext)
		}
		c++
		return FormatFilename(newFilename, c)
	}
}
func Progress(total int, current int) {
	const barsize = 30
	var ratio = int(current * barsize / total)
	bar := ""
	for i := 0; i < ratio; i++ {
		bar += "#"
	}
	for j := 0; j < barsize - ratio; j++ {
		bar += "."
	}
	fmt.Printf("\r[%s] Files: %d/%d", bar, current, total)
}
