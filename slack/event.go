package slack

import "time"

type Event struct {
	ID          int64  `json:"id"` // Every event should have a unique (for that connection) positive integer ID.
	Error       Error  `json:"error,omitempty"`
	Type        string `json:"type"`
	ChannelID   string `json:"channel"`
	Channelname string `json:"-"`
	UserID      string `json:"user,omitempty"`
	Username    string `json:"-"`
	Text        string `json:"text,omitempty"`
	Presence    string `json:"presence,omitempty"` //active, away
	SubType     string `json:"subtype,omitempty"`
	Team        string `json:"team,omitempty"`
	Ts          string `json:"ts,omitempty"`
}

func (se *Event) Usernick() string {
	return se.Username
}

func (se *Event) Msg() string {
	return se.Text
}

func (se *Event) Chan() string {
	return se.Channelname
}

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
	Deleted  bool   `json:"deleted"`
	IsBot    bool   `json:"is_bot"`
	Presence string `json:"presence"` //active, away
	lastSeen time.Time
}

type Channel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsChannel  bool   `json:"is_channel"`
	Creator    string `json:"creator"`
	IsArchived bool   `json:"is_archived"`
}

type Error struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func (sc *SlackClient) idToName(e *Event) {

	channel, ok := sc.chanIDMap[e.ChannelID]
	if ok {
		e.Channelname = channel.Name
	}

	user, ok := sc.userIDMap[e.UserID]
	if ok {
		e.Username = user.Name
	}

}

func (sc *SlackClient) nameToID(e *Event) {
	// we only have to convert the channel, since user will be our slackbot anyway
	channel, ok := sc.chanMap[e.Channelname]
	if ok {
		e.ChannelID = channel.ID
	}

}

func (sc *SlackClient) IsSelfMsg(event *Event) bool {
	return event.UserID == sc.self.ID
}
