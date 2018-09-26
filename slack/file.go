package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type FileApiResp struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	File  *File  `json:"file,omitempty"`
}

type File struct {
	ID       string   `json:"id"` // Every event should have a unique (for that connection) positive integer ID.
	UserID   string   `json:"user,omitempty"`
	PubPerma string   `json:"permalink_public,omitempty"`
	Channels []string `json:"channels,omitempty"`
}

func (sc *Client) shareFile(fileID string) {
	sc.sharemu.Lock()
	defer sc.sharemu.Unlock()
	done, ok := sc.shared[fileID]
	if done || ok {
		return
	}
	const fileSharePublicURL = ""

	// Enable public sharing URL
	client := &http.Client{}
	payload := []byte(fmt.Sprintf(`{"token": "%s", "file": "%s"}`, sc.UserToken, fileID))

	req, err := http.NewRequest("POST", "https://slack.com/api/files.sharedPublicURL", bytes.NewBuffer(payload))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("charset", "UTF-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", sc.UserToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read files.sharedPublicURL API response: %v", err)
	}

	var f FileApiResp
	if err = json.Unmarshal(body, &f); err != nil {
		log.Println(err)
		return
	}

	if !f.OK {
		log.Println("Image sharing failed:")
		if f.Error != "" {
			log.Println(f.Error)
		}
		return
	}
	if len(f.File.Channels) > 0 {

		for _, channelID := range f.File.Channels {
			msg := fmt.Sprintf("has shared a file: %s", f.File.PubPerma)
			event := &Event{Type: "message", UserID: f.File.UserID, ChannelID: channelID, Text: msg}
			sc.idToName(event)
			sc.disPatchHandlers(event)
		}
	}
	sc.shared[fileID] = true
}
