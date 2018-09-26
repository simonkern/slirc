package slack

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	BotToken  string
	UserToken string
	nextID    int64

	handlers map[string][]HandlerFunc

	self  Self
	users []User

	userIDMap map[string]*User
	channels  []Channel
	chanIDMap map[string]*Channel
	chanMap   map[string]*Channel // lookup by channame

	quit chan struct{}
	in   chan *Event
	out  chan *Event

	sharemu sync.Mutex
	shared  map[string]bool

	mu        sync.RWMutex
	connected bool

	wg sync.WaitGroup
	ws *websocket.Conn
}

type Self struct {
	ID   string
	Name string
}

type HandlerFunc func(*Client, *Event)

func (sc *Client) HandleFunc(msgType string, hf HandlerFunc) {
	sc.handlers[msgType] = append(sc.handlers[msgType], hf)
}

func (sc *Client) disPatchHandlers(event *Event) {
	if handlers, ok := sc.handlers[event.Type]; ok {
		for _, handler := range handlers {
			go handler(sc, event)
		}
	}
}

func NewClient(botToken string) (sc *Client) {
	sc = &Client{BotToken: botToken}
	sc.in = make(chan *Event, 3)
	sc.handlers = make(map[string][]HandlerFunc)
	sc.shared = make(map[string]bool)
	return sc
}

func (sc *Client) Send(target, msg string) {
	sc.send(&Event{Type: "message", Channelname: target, Text: msg})
}

func (sc *Client) send(event *Event) {
	sc.in <- event
}

func (sc *Client) updateUser(user *User) {
	sc.userIDMap[user.ID] = user
}

func (sc *Client) bookKeeping(apiResp *APIResp) {
	// store self infos
	sc.self = apiResp.Self

	// store userInfo
	sc.users = apiResp.Users

	//store chanInfo
	sc.channels = apiResp.Channels

	// create map for User lookups by ID
	sc.userIDMap = make(map[string]*User)
	// populate map
	for i, user := range sc.users {
		sc.userIDMap[user.ID] = &sc.users[i]
	}

	//create map for Chan lookups by ID
	sc.chanIDMap = make(map[string]*Channel)
	//create map for Chan lookups by Name
	sc.chanMap = make(map[string]*Channel)
	// populate maps
	for i := range sc.channels {
		channel := &sc.channels[i]
		sc.chanIDMap[channel.ID] = channel
		sc.chanMap[channel.Name] = channel
	}

}
