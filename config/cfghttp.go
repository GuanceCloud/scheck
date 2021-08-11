package config

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const (
	MB = 1024 * 1024
)

func toTestMem() {
	blocks := make([][MB]byte, 0)
	log.Println("Child pid is", os.Getpid())

	for i := 0; ; i++ {
		blocks = append(blocks, [MB]byte{})
		printMemUsage()
		time.Sleep(time.Second)
	}
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tSys = %v MiB \n", bToMb(m.Sys))
}

func bToMb(b uint64) uint64 {
	return b / MB
}

func RunCGroupForTest() {
	l.Infof("启动http 可通过 localhost:12345/mem 检测内存超过限制之后的情况")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("进入 http func")
		blocks := make([][MB]byte, 0)
		for i := 0; ; i++ {
			blocks = append(blocks, [MB]byte{})
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			fmt.Println("------ReadMemStats(&m)")
			alloc := fmt.Sprintf("Alloc = %v MiB \n", bToMb(m.Alloc))
			fmt.Println(alloc)
			sys := fmt.Sprintf("Sys = %v MiB \n", bToMb(m.Sys))
			fmt.Println(sys)
			time.Sleep(time.Second * 2)
		}
	})
	http.HandleFunc("/mem", handler)
	if err := http.ListenAndServe(":4321", nil); err != nil {
		l.Warnf("开通端口错误 err=%v", err)
	}
}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("进入 http func")
	blocks := make([][MB]byte, 0)
	for i := 0; ; i++ {
		blocks = append(blocks, [MB]byte{})
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Println("------ReadMemStats(&m)")
		fmt.Printf("Alloc = %v MiB \n", bToMb(m.Alloc))
		fmt.Printf("Sys = %v MiB \n", bToMb(m.Sys))
		fmt.Println("----http-----")
		time.Sleep(time.Second)
	}
}
