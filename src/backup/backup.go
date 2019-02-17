package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Config struct {
	src           string
	dest          string
	ignoreFiles   []string
	ignoreFolders []string
	deleteFiles   []string
	deleteFolders []string
	clean         bool
	total         uint64
	count         uint64
}

func (c *Config) countFiles() {
	c.total = 0
	_ = filepath.Walk(c.src, func(pth string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			c.total++
		}
		if err != nil {
			return err
		} else {
			return nil
		}
	})
}
func (c *Config) updateConfigSrc() {
	p := ""
	defer func() {
		if r := recover(); r != nil {
			p = ""
		}
	}()
	for i := range os.Args {
		if os.Args[i] == "-f" {
			p = os.Args[i+1]
			break
		}
	}
	if filepath.IsAbs(p) {
		config.src = p
	} else {
		cwd, _ := os.Getwd()
		config.src = path.Join(cwd, p)
	}
}
func (c *Config) updateConfigDest() {
	p := ""
	defer func() {
		if r := recover(); r != nil {
			p = ""
		}
	}()
	for i := range os.Args {
		if os.Args[i] == "-d" {
			p = os.Args[i+1]
			break
		}
	}
	if filepath.IsAbs(p) {
		config.dest = p
	} else {
		cwd, _ := os.Getwd()
		config.dest = path.Join(cwd, p)
	}
	if !strings.HasSuffix(config.dest, path.Base(config.src)) {
		config.dest = path.Join(config.dest, path.Base(config.src))
	}

}

var config Config

func main() {
	config.ignoreFiles = []string{}
	config.ignoreFolders = []string{}
	config.deleteFiles = []string{}
	config.deleteFolders = []string{}
	var answer []byte
	var configPath string
	var configPathHome string
	var reader *bufio.Reader
	var cwd string
	var err error
	reader = bufio.NewReader(os.Stdin)
	cwd, _ = os.Getwd()
	configPath = path.Join(cwd, "backup.ini")
	configPathHome = path.Join(os.Getenv("HOME"), "backup.ini")
	if _, err = os.Stat(configPath); err == nil {
		if err = loadIniToConfig(configPath, &config); err != nil {
			log.Fatal(err)
		}
	} else if _, err = os.Stat(configPathHome); err == nil {
		if err = loadIniToConfig(configPathHome, &config); err != nil {
			log.Fatal(err)
		}
	}
	//else {
	//	log.Fatal("config file doesnt exist")
	//}
	if contains("-f", os.Args) {
		config.updateConfigSrc()
	}
	if contains("-d", os.Args) {
		config.updateConfigDest()
	}
	if _, err1 := os.Stat(config.src); err1 != nil {
		log.Fatal("invalid src dir")
	}
	if _, err2 := os.Stat(config.dest); err2 != nil {
		log.Fatal("invalid dest dir")
	}
	fmt.Println("Source     \t" + config.src)
	if contains("--clean", os.Args) || config.clean {
		fmt.Println("Cleaning selected directory")
	} else {
		fmt.Println("Destination\t" + config.dest)
	}
	fmt.Print("Do you want to continue? (Y/N) ")
	answer, _, _ = reader.ReadLine()
	if strings.ToUpper(string(answer)) == "Y" {
		if contains("--clean", os.Args) || config.clean {
			config.countFiles()
			clean(config.src)
		} else {
			if err = makeDestDir(config.dest); err != nil {
				log.Fatal(err)
			} else {
				config.countFiles()
				backup(config.src)
				fmt.Print("\nDo you want to clean dest dir? (Y/N) ")
				answer, _, _ = reader.ReadLine()
				if strings.ToUpper(string(answer)) == "Y" {
					cleanSrc(config.dest)
				} else {
					fmt.Println("\nBye")
					os.Exit(0)
				}
			}
		}
	} else {
		fmt.Println("Bye")
		os.Exit(0)
	}
}

func clean(p string) {
	dirSlice, err := ioutil.ReadDir(p)
	if err != nil {
		panic("cannot read " + p)
	}
	for _, f := range dirSlice {
		absPath := path.Join(p, f.Name())
		fi, err := os.Stat(absPath)
		if err != nil {
			panic("error reading" + absPath)
		}
		if fi.IsDir() {
			if toDelete(absPath) {
				err := os.RemoveAll(absPath)
				if err != nil {
					fmt.Println("error deleting " + absPath)
				}
				fmt.Println(absPath)
			} else {
				clean(absPath)
			}
		} else if toDelete(absPath) {
			err := os.Remove(absPath)
			if err != nil {
				fmt.Println("error deleting " + absPath)
			}
			fmt.Println(absPath)
		}
	}
}
func cleanSrc(p string) {
	dirSlice, _ := ioutil.ReadDir(p)
	for _, f := range dirSlice {
		absPath := path.Join(p, f.Name())
		relPath := path.Join(config.src, getRelPathRm(absPath))
		if fi, err := os.Stat(absPath); fi.IsDir() && err == nil {
			if _, err := os.Stat(relPath); err != nil {
				if err := os.RemoveAll(absPath); err != nil {
					fmt.Println("error removing " + absPath)
				}
			} else {
				cleanSrc(absPath)
			}
		} else if fi, err := os.Stat(absPath); !fi.IsDir() && err == nil {
			if _, err := os.Stat(relPath); err != nil {
				if err := os.Remove(absPath); err != nil {
					fmt.Println("error removing " + absPath)
				}
			}
		}

	}
}
func backup(p string) {
	dirSlice, err := ioutil.ReadDir(p)
	if err != nil {
		panic("cannot read " + p)
	}
	for _, f := range dirSlice {
		absPath := path.Join(p, f.Name())
		destPath := path.Join(config.dest, getRelPath(absPath))
		if f.IsDir() && !toIngore(absPath) {
			if _, err := os.Stat(destPath); err != nil {
				err := os.Mkdir(destPath, 0777)
				if err != nil {
					panic("error creating " + destPath)
				}
			}
			backup(absPath)
		} else if !toIngore(absPath) {
			if fi1, err := os.Stat(destPath); err == nil {
				if fi2, err := os.Stat(absPath); fi2.ModTime().Nanosecond() > fi1.ModTime().Nanosecond() && err == nil {
					copy2(absPath, destPath)
				}
			} else {
				copy2(absPath, destPath)
			}
			config.count++
		} else if !f.IsDir() {
			config.count++
		} else {
			config.count += countFiles(absPath)
		}
		progress(config.total, config.count)
	}
}

func loadIniToConfig(iniPath string, config *Config) error {
	if _, err := os.Stat(iniPath); err == nil {
		configIni, err := ini.Load(iniPath)
		if err != nil {
			return errors.New("unable to load ini file")
		}
		config.src = configIni.Section("linux").Key("src").String()
		config.dest = configIni.Section("linux").Key("dest").String()
		if !strings.HasSuffix(config.dest, path.Base(config.src)) {
			config.dest = path.Join(config.dest, path.Base(config.src))
		}
		config.ignoreFiles = strings.Split(configIni.Section("ignore").Key("files").String(), ", ")
		config.ignoreFolders = strings.Split(configIni.Section("ignore").Key("folders").String(), ", ")
		config.deleteFiles = strings.Split(configIni.Section("delete").Key("files").String(), ", ")
		config.deleteFolders = strings.Split(configIni.Section("delete").Key("folders").String(), ", ")
		config.clean = configIni.Section("").Key("clean").String() == "true"
		if len(config.src) == 0 || len(config.dest) == 0 {
			return errors.New("invalid config file")
		} else {
			return nil
		}
	} else {
		return errors.New("invalid ini path")
	}

}

func toIngore(p string) bool {
	base := path.Base(p)
	fi, err := os.Stat(p)
	if err != nil {
		panic("unable to read file info " + p)
	}
	if fi.IsDir() {
		if contains(base, config.ignoreFolders) {
			return true
		}
	} else {
		if contains(base, config.ignoreFiles) {
			return true
		}
	}
	return false

}
func toDelete(p string) bool {
	base := path.Base(p)
	fi, err := os.Stat(p)
	if err != nil {
		fmt.Println("error reading " + p)
	}
	if fi.IsDir() {
		if contains(base, config.deleteFolders) {
			return true
		}
	} else {
		if contains(base, config.deleteFiles) {
			return true
		}
	}
	return false
}
func countFiles(p string) uint64 {
	counter := uint64(0)
	_ = filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		} else {
			if !info.IsDir() {
				counter++
			}
			return nil
		}
	})
	return counter
}
func getRelPath(p string) string {
	return p[len(config.src):]
}
func getRelPathRm(p string) string {
	return p[len(config.dest):]
}
func contains(search string, slice []string) bool {
	for _, item := range slice {
		if search == item {
			return true
		}
	}
	return false
}
func makeDestDir(p string) error {
	if _, err := os.Stat(p); err != nil {
		err = os.MkdirAll(p, 0777)
		if err != nil {
			return errors.New("unable to create dest dir")
		}
	}
	return nil
}
func copy2(src string, dest string) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = ioutil.WriteFile(dest, input, 0777)
	if err != nil {
		fmt.Println("error creating", dest)
		fmt.Println(err)
		return
	}
}
func progress(total uint64, count uint64) {
	var barLen uint64 = 30
	bar := ""
	fillLen := int(barLen * count / total)
	for i := 0; i < fillLen; i++ {
		bar += "#"
	}
	for j := 0; j < int(barLen-uint64(fillLen)); j++ {
		bar += "."
	}
	fmt.Printf("[%s] Files: %d/%d", bar, count, total)
	fmt.Print("\r")
}
