package slack

import (
	"sync"

	"github.com/gorilla/websocket"
)

type SlackClient struct {
	SlackToken string
	nextID     int64

	handlers map[string][]HandlerFunc

	self      Self
	users     []UserProfile
	userIDMap map[string]*UserProfile
	channels  []Channel
	chanIDMap map[string]*Channel
	chanMap   map[string]*Channel // lookup by channame

	quit chan struct{}
	in   chan *Event
	out  chan *Event

	mu        sync.RWMutex
	connected bool

	wg sync.WaitGroup
	ws *websocket.Conn
}

type Self struct {
	ID   string
	Name string
}

type HandlerFunc func(*SlackClient, *Event)

func (sc *SlackClient) HandleFunc(msgType string, hf HandlerFunc) {
	sc.handlers[msgType] = append(sc.handlers[msgType], hf)
}

func (sc *SlackClient) disPatchHandlers(event *Event) {
	if handlers, ok := sc.handlers[event.Type]; ok {
		for _, handler := range handlers {
			go handler(sc, event)
		}
	}
}

func NewSlackClient(token string) (sc *SlackClient) {
	sc = &SlackClient{SlackToken: token}
	sc.in = make(chan *Event, 3)
	sc.handlers = make(map[string][]HandlerFunc)
	return sc
}

func (sc *SlackClient) Send(target, msg string) {
	sc.send(&Event{Type: "message", Channelname: target, Text: msg})
}

func (sc *SlackClient) send(event *Event) {
	sc.in <- event
}

func (sc *SlackClient) updateUser(userProfile *UserProfile) {
	sc.userIDMap[userProfile.ID] = userProfile
}

func (sc *SlackClient) bookKeeping(apiResp *SlackAPIResponse) {
	// store self infos
	sc.self = apiResp.Self

	// store userInfo
	sc.users = apiResp.Users

	//store chanInfo
	sc.channels = apiResp.Channels

	// create map for User lookups by ID
	sc.userIDMap = make(map[string]*UserProfile)
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
