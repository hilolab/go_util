package main

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"flag"
	"sync/atomic"
	"time"
	"strings"
	"path"
)

var cmd *exec.Cmd
var dir *string = flag.String("d","","work directory")
//var process *string = flag.String("p","","process")
var isBuilding int32

func main() {
	flag.Parse()

	if *dir == "" || !isDir(*dir){
		log.Fatalln("error: work dir is empty")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	os.Chdir(*dir)
	done := make(chan bool)

	var process string = path.Base(*dir)
	if strings.Contains(strings.ToLower(os.Getenv("GOOS")),"window") {
		process += ".exe"
	}

	go func() {
		for {
			select {
			case ev := <-watcher.Event:

				if strings.Contains(filepath.Base(ev.Name),process){
					continue
				}
				if !atomic.CompareAndSwapInt32(&isBuilding,0,1){
					continue
				}

				go func() {
					build(*dir+"/"+process)
					time.Sleep(time.Second*2)
					atomic.CompareAndSwapInt32(&isBuilding,1,0)
				}()

			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	watchFiles(*dir,watcher)
	if err != nil {
		log.Fatal(err)
	}

	<-done
	watcher.Close()
}

func watchFiles(dir string,wth *fsnotify.Watcher){
	files,err := filepath.Glob(dir)
	checkErr(err)
	for _,file := range files{
		if ck := isDir(file);ck{
			wth.Watch(file)
			watchFiles(file+"/*",wth)
		}
	}
}

func isDir(dir string) bool{
	stat,err := os.Stat(dir)
	checkErr(err)

	return stat.IsDir()
}

func checkErr(err error){
	if err != nil{
		log.Fatalln("err: ",err)
	}
}

func build(process string)  {

	fmt.Println("\n\n=============================== Process =====================================")
	icmd := exec.Command("go", "build")
	icmd.Stderr = os.Stderr
	_, err := icmd.Output()
	if err != nil {
		fmt.Println("[build] error: ", err)
		return
	}
	if cmd != nil && cmd.Process != nil {
		err = cmd.Process.Kill()

		if err != nil {
			fmt.Println("[restart] kill error: ", err)
		}
	}

	cmd = exec.Command(process)
	cmd.Stdout = os.Stdout
	go cmd.Run()
	fmt.Println("[build] done!")
}
