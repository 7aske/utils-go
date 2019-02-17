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

var config Config
var wg = sync.WaitGroup{}
var errorStrings = []string{"Changes to be committed", "Changes not staged for commit", "Untracked files"}
var statusOutput []string
var counter = 0
var notCounter = 0
var ignoredFolders = []string{"_test", "_others"}

func main() {
	runtime.GOMAXPROCS(100)
	var err error
	var cwd string
	var configPath string
	var configPathHome string
	cwd, _ = os.Getwd()
	configPath = path.Join(cwd, "gitstatus.ini")
	configPathHome = path.Join(os.Getenv("HOME"), "gitstatus.ini")
	if _, err = os.Stat(configPath); err == nil {
		if err = loadIniToConfig(configPath, &config); err != nil {
			log.Fatal(err)
		}
	} else if _, err = os.Stat(configPathHome); err == nil {
		if err = loadIniToConfig(configPathHome, &config); err != nil {
			log.Fatal(err)
		}
	}
	if contains("-f", &os.Args) {
		config.updateConfigSrc()
	}
	if _, err = os.Stat(config.src); err != nil {
		log.Fatal("invalid src dir")
	}
	fmt.Printf("Path: \t%s\n\n", config.src)
	wg.Add(1)
	go getGits(config.src)
	wg.Wait()
	for _, repo := range statusOutput {
		fmt.Println(repo)
	}
	fmt.Println("\nRepositories checked: " + strconv.Itoa(counter))
	fmt.Println("Not up-to-date: " + strconv.Itoa(notCounter))
}
func getGits(src string) {
	dir, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatal("unable to read src dir")
	}
	for _, f := range dir {
		absPath := path.Join(src, f.Name())
		if err != nil {
			fmt.Println("error reading " + absPath)
		}
		if f.IsDir() && !contains(f.Name(), &ignoredFolders) {
			if f.Name() == ".git" {
				wg.Add(1)
				go gitStatus(src)
			} else if containsDir(".git", absPath) {
				wg.Add(1)
				go gitStatus(absPath)
			}
			fsDir, err := ioutil.ReadDir(absPath)
			if err != nil {
				log.Fatal("unable to read src dir")
			}
			for _, fs := range fsDir {
				fsAbsPath := path.Join(absPath, fs.Name())
				if fs.IsDir() {
					if containsDir(".git", fsAbsPath) {
						wg.Add(1)
						go gitStatus(fsAbsPath)
					}
				}
			}
		}
	}
	//wg.Wait()
	wg.Done()
}
func gitStatus(p string) {
	status := exec.Command("git", "-C", p, "status")
	out, err := status.Output()
	if err != nil {
		fmt.Println("error checking status of " + path.Base(p))
	}
	if !isCommited(string(out)) {
		statusOutput = append(statusOutput, p)
		notCounter++
	}
	counter++
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
