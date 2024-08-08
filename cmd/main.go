package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/borowiak-m/sys-info-goth/internal/hardware"
)

type server struct {
	buffer int
	mux    http.ServeMux
}

func NewServer() *server {
	s := &server{
		buffer: 10,
	}

	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))

	return s
}

func main() {
	fmt.Println("Starting system monitor...")
	go func() {
		for {
			systemSection, err := hardware.GetSystemSection()
			if err != nil {
				fmt.Println(err)
			}
			diskSection, err := hardware.GetDiskSection()
			if err != nil {
				fmt.Println(err)
			}
			cpuSection, err := hardware.GetCPUSection()
			if err != nil {
				fmt.Println(err)
			}

			fmt.Println(systemSection)
			fmt.Println(diskSection)
			fmt.Println(cpuSection)

			time.Sleep(3 * time.Second)
		}
	}()
	srv := NewServer()
	err := http.ListenAndServe(":8080", &srv.mux)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
