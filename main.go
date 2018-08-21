package main

import (
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
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
						out, err := os.Create("/Volumes/MICROBIT//microbit.hex")
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
						log.Println("complete:", fname)
						os.Remove(event.Name)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	usr, _ := user.Current()
	err = watcher.Add(filepath.Join(usr.HomeDir, "Downloads"))
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
