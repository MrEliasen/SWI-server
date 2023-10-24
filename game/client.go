package game

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/mreliasen/swi-server/game/settings"
	"github.com/mreliasen/swi-server/internal/logger"
	"github.com/pterm/pterm"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type Client struct {
	Game          *Game
	Player        *Entity
	Authenticated bool
	UserId        uint64
	UUID          string
	UserType      uint8
	CombatLogging bool
	Connection    *websocket.Conn
	Send          chan protoreflect.ProtoMessage
	Mu            sync.Mutex
}

func (c *Client) SendEvent(msg protoreflect.ProtoMessage) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("Trid to send to closed client channel:", err)
		}
	}()

	/* val, err := proto.Marshal(msg)
	if err != nil {
		logger.Logger.Error(fmt.Sprintf("Error marshaling: %s", err.Error()))
		return
	} */

	if c.Connection == nil {
		return
	}

	c.Send <- msg
}

func (c *Client) handleOutput() {
	ticker := time.NewTicker(settings.PingPeriod)
	defer func() {
		defer func() {
			if err := recover(); err != nil {
				log.Println("Connection closure issue:", err)
			}
		}()

		ticker.Stop()

		c.Mu.Lock()
		defer c.Mu.Unlock()

		if c.Connection == nil {
			return
		}

		c.Connection.Close()
		c.Connection = nil
		c.Game.Logout <- c
	}()

	for {
		select {
		// send any mssages to the client which are available
		case msg, ok := <-c.Send:
			if c.Connection == nil {
				return
			}

			c.Connection.SetWriteDeadline(time.Now().Add(settings.WriteWait))
			if !ok {
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Connection.NextWriter(websocket.BinaryMessage)
			if err != nil {
				return
			}

			m, err := anypb.New(msg)
			if err != nil {
				logger.Logger.Error(err.Error())
				return
			}

			wire, err := proto.Marshal(m)
			if err != nil {
				logger.Logger.Error(err.Error())
				return
			}

			w.Write(wire)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})

				m, err := anypb.New(<-c.Send)
				if err != nil {
					logger.Logger.Error(err.Error())
					continue
				}

				wire, err := proto.Marshal(m)
				if err != nil {
					logger.Logger.Error(err.Error())
					continue
				}

				w.Write(wire)
			}

			if err := w.Close(); err != nil {
				return
			}

		// keep alive, check client still connected
		case <-ticker.C:
			if c.Connection == nil {
				return
			}

			c.Connection.SetWriteDeadline(time.Now().Add(settings.WriteWait))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleInput() {
	defer func() {
		c.Mu.Lock()
		c.Connection.Close()
		c.Connection = nil
		c.Mu.Unlock()

		c.Game.Logout <- c
	}()

	c.Connection.SetReadLimit(settings.MaxMessageSize)
	c.Connection.SetReadDeadline(time.Now().Add(settings.PongWait))
	c.Connection.SetPongHandler(func(string) error {
		c.Connection.SetReadDeadline(time.Now().Add(settings.PongWait))
		return nil
	})

	for {
		if c.Connection == nil {
			return
		}

		_, msg, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Logger.Warn(fmt.Sprintf("WS Error: %s", err))
			}
			break
		}

		if c.Player != nil {
			logger.Logger.Trace(pterm.Sprintf("%s: %s", c.Player.Name, msg))
		}

		ExecuteCommand(c, string(msg))
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWsClient(g *Game, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Logger.Trace(pterm.Sprint(err))
		return
	}

	client := &Client{
		Game:          g,
		Connection:    conn,
		UUID:          uuid.New().String(),
		Authenticated: false,
		Send:          make(chan protoreflect.ProtoMessage),
	}

	logger.Logger.Trace("New connection.")
	time.Sleep(500 * time.Millisecond)

	go client.handleInput()
	go client.handleOutput()
	g.SendMOTD(client)
}
