package main

import (
	"errors"
	"fmt"
	"github.com/go-ini/ini"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type Config struct {
	src string
}

func (c *Config) updateConfigSrc(f string) {
	p := ""
	defer func() {
		if r := recover(); r != nil {
			p = ""
		}
	}()
	for i := range os.Args {
		if os.Args[i] == f {
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

var config Config
var wg = sync.WaitGroup{}
var errorStrings = []string{"Changes to be committed", "Changes not staged for commit", "Untracked files"}
var counter = 0
var notCounter = 0
var ignoredFolders = []string{"_test", "_others"}

func main() {
	runtime.GOMAXPROCS(100)
	var err error
	ex, err := os.Executable()
	if err != nil {
		log.Fatal("cannot locate executable file")
	}
	exPath := filepath.Dir(ex)
	configPath := path.Join(exPath, "gitstatus.ini")
	configPathHome := path.Join(os.Getenv("HOME"), "gitstatus.ini")
	if contains("-f", &os.Args) {
		config.updateConfigSrc("-f")
	} else if _, err = os.Stat(configPath); err == nil {
		if err = loadIniToConfig(configPath, &config); err != nil {
			log.Fatal("cannot load $PWD ini file")
		}
	} else if _, err = os.Stat(configPathHome); err == nil {
		if err = loadIniToConfig(configPathHome, &config); err != nil {
			log.Fatal("cannot load $HOME ini file")
		}
	} else {
		log.Fatal("cannot load src dir")
	}

	if _, err = os.Stat(config.src); err != nil {
		log.Fatal("invalid src dir")
	}
	fmt.Printf("Path: \t%s\n\n", config.src)
	getGits(config.src)
	fmt.Println("\nRepositories checked: " + strconv.Itoa(counter))
	fmt.Println("Not up-to-date: " + strconv.Itoa(notCounter))
}
func getGits(src string) {
	dir, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatal("unable to read dir " + src)
	}
	for _, f := range dir {
		if f.IsDir() && !contains(f.Name(), &ignoredFolders) {
			absPath := path.Join(src, f.Name())
			if f.Name() == ".git" {
				wg.Add(1)
				go gitStatus(src)
			} else if containsDir(".git", absPath) {
				wg.Add(1)
				go gitStatus(absPath)
			} else {
				fsDir, err := ioutil.ReadDir(absPath)
				if err != nil {
					log.Fatal("unable to read dir " + absPath)
				}
				for _, fs := range fsDir {
					if fs.IsDir() {
						fsAbsPath := path.Join(absPath, fs.Name())
						if containsDir(".git", fsAbsPath) {
							wg.Add(1)
							go gitStatus(fsAbsPath)
						}
					}
				}
			}
		}
	}
	wg.Wait()
}
func gitStatus(p string) {
	status := exec.Command("git", "-C", p, "status")
	out, err := status.Output()
	if err != nil {
		fmt.Println("error checking status of " + path.Dir(p) + "/" + path.Base(p))
	}
	counter++
	if !isCommited(string(out)) {
		slice := strings.Split(p, filepath.Base(config.src))
		fmt.Println(slice[len(slice)-1])
		notCounter++
	}
	wg.Done()
}
func contains(search string, slice *[]string) bool {
	for _, item := range *slice {
		if search == item {
			return true
		}
	}
	return false
}

func isCommited(s string) bool {
	for _, e := range errorStrings {
		if strings.Contains(s, e) {
			return false
		}
	}
	return true
}

func containsDir(search string, p string) bool {
	dir, err := ioutil.ReadDir(p)
	if err != nil {
		fmt.Println("error reading " + p)
	}
	for _, item := range dir {
		if search == item.Name() {
			return true
		}
	}
	return false
}
func loadIniToConfig(iniPath string, config *Config) error {
	if _, err := os.Stat(iniPath); err == nil {
		configIni, err := ini.Load(iniPath)
		if err != nil {
			return errors.New("unable to load ini file")
		}
		config.src = configIni.Section("path").Key("src").String()
		if len(config.src) == 0 {
			return errors.New("invalid config file")
		} else {
			return nil
		}
	} else {
		return errors.New("invalid ini path")
	}

}
