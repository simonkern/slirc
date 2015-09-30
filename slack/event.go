package slack

type Event struct {
	Id        int64  `json:"id"` // Every event should have a unique (for that connection) positive integer ID.
	Error     Error  `json:"error,omitempty"`
	Type      string `json:"type"`
	ChannelID string `json:"channel"`
	Channel   string `json:"-"`
	UserID    string `json:"user,omitempty"`
	User      string `json:"-"`
	Text      string `json:"text,omitempty"`
	Presence  string `json:"presence,omitempty"` //active, away
	SubType   string `json:"subtype,omitempty"`
	Team      string `json:"team,omitempty"`
	Ts        string `json:"ts,omitempty"`
}

type Error struct {
	Code int    `json:"code,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func (se *Event) Username() string {
	return se.User
}

func (se *Event) Msg() string {
	return se.Text
}

func (se *Event) Chan() string {
	return se.Channel
}

func (sc *SlackClient) idToName(e *Event) {

	channel, ok := sc.chanIDMap[e.ChannelID]
	if ok {
		e.Channel = channel.Name
	}

	user, ok := sc.userIDMap[e.UserID]
	if ok {
		e.User = user.Name
	}

}

func (sc *SlackClient) nameToID(e *Event) {
	// we only have to convert the channel, since user will be our slackbot anyway
	channel, ok := sc.chanMap[e.Channel]
	if ok {
		e.ChannelID = channel.Id
	}

}

func (sc *SlackClient) IsSelfMsg(event *Event) bool {
	return event.UserID == sc.self.Id
}
