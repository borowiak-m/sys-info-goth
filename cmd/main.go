package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/borowiak-m/sys-info-goth/internal/hardware"
	"nhooyr.io/websocket"
)

type server struct {
	buffer      int
	mux         http.ServeMux
	subscribers map[*subscriber]struct{}
}

type subscriber struct {
	msgs chan []byte
}

func NewServer() *server {
	s := &server{
		buffer:      10,
		subscribers: make(map[*subscriber]struct{}),
	}

	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))

	return s
}

func (s *server) subscribeHandler(w http.ResponseWriter, req *http.Request) {
	err := s.subscribe(req.Context(), w, req)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *server) subscribe(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	var c *websocket.Conn
	subscriber := &subscriber{
		msgs: make(chan []byte, s.buffer),
	}
	s.addSubscriber(subscriber)

	c, err := websocket.Accept(w, req, nil)
	if err != nil {
		return err
	}
	defer c.CloseNow()

	ctx = c.CloseRead(ctx)
	for {
		// add sending messages to subs
	}
}

func (s *server) addSubscriber(sub *subscriber) // to do

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
