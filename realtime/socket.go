package realtime

import (
	"log"

	socketio "github.com/googollee/go-socket.io"
)

var Server *socketio.Server

func InitSocket() {
	var err error
	Server = socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	Server.OnConnect("/", func(c socketio.Conn) error {
		log.Println("connected:", c.ID())
		return nil
	})

	go func() {
		if err := Server.Serve(); err != nil {
			log.Fatal("socket listen error:", err)
		}
	}()
}
