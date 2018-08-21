package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	var microbitPath = ""
	usr, _ := user.Current()
	var downloadPath = filepath.Join(usr.HomeDir, "Downloads")

	switch runtime.GOOS {
	case "darwin":
		microbitPath = "/Volumes/MICROBIT"
	case "linux":
		microbitPath = ""
	case "windows":
		microbitPath = ""
	default:
		microbitPath = ""
	}

	flag.StringVar(&microbitPath, "microbit", microbitPath, "microbit dir")
	flag.StringVar(&downloadPath, "download", downloadPath, "downloads dir")
	flag.Parse()

	if microbitPath == "" {
		flag.PrintDefaults()
		os.Exit(2)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("failed to notify watcher: ", err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Create == fsnotify.Create {
					fname := filepath.Base(event.Name)
					if strings.HasPrefix(fname, "microbit-") && filepath.Ext(fname) == ".hex" {
						in, err := os.Open(event.Name)
						if err != nil {
							log.Fatal(err)
						}
						out, err := os.Create(filepath.Join(microbitPath, "microbit.hex"))
						if err != nil {
							log.Println("skip:", fname)
							os.Remove(event.Name)
							break
						}
						if _, err := io.Copy(out, in); err != nil {
							log.Println("write error:", fname, err)
							break
						}
						out.Close()
						if runtime.GOOS == "darwin" {
							exec.Command("diskutil", "unmountDisk", microbitPath).Run()
						}
						log.Println("complete:", fname)
						os.Remove(event.Name)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(downloadPath)
	if err != nil {
		log.Fatalf("failed to watch %s : %s", downloadPath, err)
	}
	<-done
}
