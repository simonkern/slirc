package slack

import (
	"fmt"
	"regexp"
	"strings"
)

// Slack usually escapes < and >, however, highlights, links etc.
// are wrapped in unescaped < >
var bracketRe = regexp.MustCompile("(<.+?>)")

func (sc *Client) nickForUserID(userID string) string {
	user, ok := sc.userIDMap[userID]
	if ok {
		if user.Profile.DisplayName == "" {
			return user.Name
		}
		return user.Profile.DisplayName
	}
	return userID
}

func (sc *Client) unSlackify(str string) string {
	// Links e.g. <http://heise.de|heise.de>, <http://heise.de>
	if strings.HasPrefix(str, "<http") {
		endpos := strings.IndexRune(str, '|')
		if endpos != -1 {
			return str[1:endpos]
		}
		return str[1 : len(str)-1]
	}
	// Highlights e. g. <@U02A2A2A2>
	if strings.HasPrefix(str, "<@U") {
		userID := str[2 : len(str)-1]
		user, ok := sc.userIDMap[userID]
		if ok {
			if user.Profile.DisplayName != "" {
				return fmt.Sprint("@", user.Profile.DisplayName)
			}
			if user.Name != "" {
				return fmt.Sprint("@", user.Name)
			}
		}
		// if we do not have a match, just return the ID.
		return str[1 : len(str)-1]
	}
	// Mail addresses e.g. <mailto:foo@bar.com|foo@bar.com>, <mailto:test@example.org>
	if strings.HasPrefix(str, "<mailto:") {
		endpos := strings.IndexRune(str, '|')
		if endpos != -1 {
			return str[8:endpos]
		}
		return str[8 : len(str)-1]
	}
	// Channels <#C02A2A2A2>
	if strings.HasPrefix(str, "<#C") {
		chanID := str[2 : len(str)-1]
		channel, ok := sc.chanIDMap[chanID]
		if ok {
			return fmt.Sprintf("#%v", channel.Name)
		}
		return str[1 : len(str)-1]
	}
	return str
}
