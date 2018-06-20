package slack

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Represents the API response of rtm.start
// See https://api.slack.com/methods/rtm.start
type SlackAPIResponse struct {
	Ok       bool      `json:"ok"`
	Self     Self      `json:"self"`
	Error    string    `json:"error"`
	Users    []User    `json:"users"`
	Channels []Channel `json:"channels"`
	Url      string    `json:"url"`
}

const slackAPIEndpoint = "https://slack.com/api/rtm.start"

func (sc *SlackClient) Connected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.connected
}

func (sc *SlackClient) Send(target, msg string) {
	event := &Event{Type: "message", Channelname: target, Text: msg}
	sc.send(event)
}

func (sc *SlackClient) send(event *Event) {
	sc.in <- event
}

func (sc *SlackClient) dispatchLoop() {
	defer sc.wg.Done()

	for {
		select {

		case <-sc.quit:
			return

		case event := <-sc.out:
			sc.disPatchHandlers(event)
		}

	}
}

func (sc *SlackClient) readLoop() {
	defer sc.wg.Done()

	sc.ws.SetReadDeadline(time.Now().Add(pongWait))
	sc.ws.SetPongHandler(func(string) error { sc.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {

		messageType, r, err := sc.ws.NextReader()
		if err != nil {
			log.Println(err)
			// If we do not start a seperate Goroutine and return,
			// we will never decrease our wg counter
			go sc.handleDisconnect()
			return
		}

		msg, err := ioutil.ReadAll(r)
		if err != nil {
			log.Println(err)
			go sc.handleDisconnect()
			return
		}

		var event Event
		err = json.Unmarshal(msg, &event)
		if err != nil {
			log.Printf("Failed to unmarshal the following rawEvent with messageType: %v \n", messageType)
			log.Println(string(msg))
			continue
		}
		event.Text = html.UnescapeString(bracketRe.ReplaceAllStringFunc(event.Text, sc.unSlackify))

		sc.idToName(&event)
		sc.out <- &event
	}
}

func (sc *SlackClient) writeLoop() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		sc.wg.Done()
	}()

	for {
		select {
		case <-sc.quit:
			return
		case event := <-sc.in:

			// replace Channel Name with ID
			channel, ok := sc.chanMap[event.Chan()]
			if !ok {
				log.Printf("Unknown Channel %s \n", event.Chan())
				continue
			}
			event.ChannelID = channel.ID
			// set event's ID
			event.ID = sc.nextID

			err := sc.ws.WriteJSON(&event)
			if err != nil {
				log.Println(err)
				// If we do not start a seperate Goroutine and return,
				// we will never decrease our wg counter
				go sc.handleDisconnect()
				return
			}
			sc.nextID++
		case <-ticker.C:
			if err := sc.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				go sc.handleDisconnect()
				return
			}
		}
	}
}

func (sc *SlackClient) handleDisconnect() {
	sc.mu.Lock()

	if !sc.connected {
		// release mutex immediately and return
		sc.mu.Unlock()
		return
	}

	sc.Close()
	sc.connected = false
	close(sc.quit)
	sc.mu.Unlock()

	// Announce shutdown in progress
	shutdownEvent := &Event{Type: "shutdown"}
	sc.disPatchHandlers(shutdownEvent)

	// Send disconnected event after all goroutines have been stopped.
	sc.wg.Wait()
	log.Println("Slack: stopped all Goroutines.")
	dcEvent := &Event{Type: "disconnected"}
	sc.disPatchHandlers(dcEvent)

}

func (sc *SlackClient) connect() (err error) {
	// create quit chan, on which we broadcast goroutine shutdowns
	sc.quit = make(chan struct{})

	for {
		wsAddr, err := sc.startRTM()
		if err != nil {
			log.Println("SlackRTMStart failed: ", err)
			log.Println("Trying again in 30 seconds...")
			time.Sleep(30 * time.Second)
			continue
		}
		err = sc.connectWS(wsAddr)
		if err != nil {
			log.Println("SlackWS reconnect failed: ", err)
			log.Println("Trying again in 30 seconds...")
			time.Sleep(30 * time.Second)
			continue
		}
		// success
		break
	}

	sc.mu.Lock()
	sc.connected = true
	sc.mu.Unlock()

	sc.wg.Add(3)
	go sc.dispatchLoop()
	go sc.readLoop()
	go sc.writeLoop()

	connectedEvent := &Event{Type: "connected"}
	sc.disPatchHandlers(connectedEvent)

	return nil
}

// startRTM() calls Slack
func (sc *SlackClient) startRTM() (wsAddr string, err error) {
	resp, err := http.PostForm(slackAPIEndpoint, url.Values{"token": {sc.SlackToken}})
	if err != nil {
		return "", fmt.Errorf("Failed to obtain websocket address: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read API response: %v", err)
	}

	var apiResp SlackAPIResponse
	if err = json.Unmarshal(body, &apiResp); err != nil {
		return
	}

	if !apiResp.Ok {
		return "", fmt.Errorf("start.RTM failed: %v", apiResp.Error)
	}
	sc.bookKeeping(&apiResp)
	return apiResp.Url, nil
}

func (sc *SlackClient) connectWS(wsAddr string) (err error) {
	log.Println("Connecting to Websocket at address:")
	log.Println(wsAddr)
	ws, _, err := websocket.DefaultDialer.Dial(wsAddr, nil)
	if err != nil {
		return err
	}
	sc.ws = ws
	// set ID for next send on this connection to 1
	sc.nextID = 1

	var event Event
	err = sc.ws.ReadJSON(&event)
	if err != nil {
		return fmt.Errorf("Failed to read initial hello from Websocket: %v", err)
	}
	// First read should yield {"type":"hello"}
	if event.Type != "hello" {
		return fmt.Errorf("Expected to get hello, but got %v", event.Type)
	}
	return
}

func (sc *SlackClient) Close() {
	if sc.ws != nil {
		sc.ws.Close()
	}
}
