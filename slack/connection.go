package slack

import (
	"encoding/json"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// APIResp represents the API response of rtm.start
// See https://api.slack.com/methods/rtm.start
type APIResp struct {
	Ok       bool      `json:"ok"`
	Self     Self      `json:"self"`
	Error    string    `json:"error"`
	Users    []User    `json:"users"`
	Channels []Channel `json:"channels"`
	URL      string    `json:"url"`
}

type EventType struct {
	Type string `json:"type"`
}

const slackAPIEndpoint = "https://slack.com/api/rtm.start"

func (sc *Client) Connected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.connected
}

func (sc *Client) Connect() (err error) {
	err = sc.connect()
	return err
}

func (sc *Client) connect() (err error) {
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

	sc.wg.Add(2)
	go sc.readLoop()
	go sc.writeLoop()

	connectedEvent := &Event{Type: "connected"}
	sc.disPatchHandlers(connectedEvent)

	return nil
}

func (sc *Client) readLoop() {
	defer sc.wg.Done()
	sc.ws.SetReadDeadline(time.Now().Add(pongWait))
	sc.ws.SetPongHandler(func(string) error {
		sc.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

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

		// unmarshal to temp struct and check whether it is a bookkeeping event, or a regular event
		var et EventType
		if err := json.Unmarshal(msg, &et); err != nil {
			log.Printf("Failed to unmarshal the following rawEvent with messageType: %v \n", messageType)
			log.Println(string(msg))
			continue
		}

		if et.Type == "file_public" {
			if sc.UserToken != "" {
				var fe FileEvent
				if err := json.Unmarshal(msg, &fe); err != nil {
					log.Printf("Failed to unmarshal the following rawEvent with messageType: %v \n", messageType)
					log.Println(string(msg))
					continue
				}
				go sc.shareFile(fe.FileID)
			}

		}

		// bookkeeping event
		if et.Type == "user_change" || et.Type == "team_join" {
			var ue UserEvent
			if err := json.Unmarshal(msg, &ue); err != nil {
				log.Printf("Failed to unmarshal the following rawEvent with messageType: %v \n", messageType)
				log.Println(string(msg))
				continue
			}
			sc.updateUser(ue.User)
			continue
		}

		// normal event
		var event Event
		if err := json.Unmarshal(msg, &event); err != nil {
			log.Printf("Failed to unmarshal the following rawEvent with messageType: %v \n", messageType)
			log.Println(string(msg))
			continue
		}
		event.Text = html.UnescapeString(bracketRe.ReplaceAllStringFunc(event.Text, sc.unSlackify))
		sc.idToName(&event)

		// is this a command?
		if strings.HasPrefix(event.Text, fmt.Sprint("@", sc.self.Name)) {
			event.Type = "command"
			if len(event.Text) > len(sc.self.Name)+1 {
				event.Text = strings.TrimSpace(event.Text[len(sc.self.Name)+1:])
			}
			user, ok := sc.userIDMap[event.UserID]
			if ok && user.IsAdmin {
				log.Println("admin-command found: ", event.Text)
				event.Type = "admincommand"
			}
		}
		go sc.disPatchHandlers(&event)
	}
}

func (sc *Client) writeLoop() {
	ticker := time.NewTicker(pingPeriod)

	defer ticker.Stop()
	defer sc.wg.Done()

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

func (sc *Client) handleDisconnect() {
	sc.mu.Lock()

	if !sc.connected {
		// release mutex immediately and return
		sc.mu.Unlock()
		return
	}
	sc.connected = false
	sc.mu.Unlock()
	sc.close()

	// Send disconnected event after all goroutines have been stopped.
	sc.wg.Wait()
	log.Println("Slack: stopped all Goroutines.")
	dcEvent := &Event{Type: "disconnected"}
	sc.disPatchHandlers(dcEvent)
}

// startRTM() calls Slack
func (sc *Client) startRTM() (wsAddr string, err error) {
	resp, err := http.PostForm(slackAPIEndpoint, url.Values{"token": {sc.BotToken}})
	if err != nil {
		return "", fmt.Errorf("Failed to obtain websocket address: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read API response: %v", err)
	}

	var apiResp APIResp
	if err = json.Unmarshal(body, &apiResp); err != nil {
		return
	}

	if !apiResp.Ok {
		return "", fmt.Errorf("start.RTM failed: %v", apiResp.Error)
	}
	sc.bookKeeping(&apiResp)
	return apiResp.URL, nil
}

func (sc *Client) connectWS(wsAddr string) (err error) {
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

func (sc *Client) Close() {
	// Announce shutdown in progress
	shutdownEvent := &Event{Type: "shutdown"}
	sc.disPatchHandlers(shutdownEvent)
	sc.close()
}

func (sc *Client) close() {
	if sc.ws != nil {
		sc.ws.Close()
	}
	close(sc.quit)
}
