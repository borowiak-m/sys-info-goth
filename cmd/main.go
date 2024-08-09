package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/borowiak-m/sys-info-goth/internal/hardware"
	"nhooyr.io/websocket"
)

type server struct {
	msgsBuffer  int
	mux         http.ServeMux
	subsMutex   sync.Mutex
	subscribers map[*subscriber]struct{}
}

type subscriber struct {
	msgs chan []byte
}

func NewServer() *server {
	s := &server{
		msgsBuffer:  10,
		subscribers: make(map[*subscriber]struct{}),
	}

	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)

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
	var socConn *websocket.Conn // create websocket connection
	subscriber := &subscriber{
		msgs: make(chan []byte, s.msgsBuffer),
	}
	s.addSubscriber(subscriber) // add client to a map of subs

	socConn, err := websocket.Accept(w, req, nil)
	if err != nil {
		return err
	}
	defer socConn.CloseNow()

	ctx = socConn.CloseRead(ctx) //io closer
	for {                        // fetch messages from chan until ctx.Done
		select {
		case msg := <-subscriber.msgs:
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			err := socConn.Write(ctx, websocket.MessageText, msg) // write msg to websocket
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *server) addSubscriber(sub *subscriber) {
	s.subsMutex.Lock()
	s.subscribers[sub] = struct{}{}
	s.subsMutex.Unlock()
	fmt.Println("Added subscriber", sub)
}

func (s *server) broadcast(msg []byte) {
	s.subsMutex.Lock()
	for subscriber := range s.subscribers {
		subscriber.msgs <- msg
	}
	s.subsMutex.Unlock()
}

func main() {
	fmt.Println("Starting system monitor...")
	srv := NewServer()
	go func(s *server) {
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
			timestamp := time.Now().Format("2006-01-02 15:04:05")

			html := `
			<div hx-swap-oob="innerHTML:#update-timestamp"> ` + timestamp + `</div>
			<div hx-swap-oob="innerHTML:#system-data"> ` + systemSection + `</div>
			<div hx-swap-oob="innerHTML:#disk-data"> ` + diskSection + `</div>
			<div hx-swap-oob="innerHTML:#cpu-data"> ` + cpuSection + `</div>
			`
			s.broadcast([]byte(html))

			time.Sleep(3 * time.Second)
		}
	}(srv)
	err := http.ListenAndServe(":8080", &srv.mux)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
