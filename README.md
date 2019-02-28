# slirc

[![Build Status](https://travis-ci.org/simonkern/slirc.svg)](https://travis-ci.org/simonkern/slirc)

Slirc links an IRC and a Slack channel.

## Example Usage

NewBridge has the following signature:

```go
func NewBridge(conf *slirc.Config) (bridge *Bridge)
```

### Example with IRC authentication

```go
package main

import (
        "github.com/simonkern/slirc"
)

// Example with IRC Authentication
// Slack Chan without "#"-prefix
func main() {
	ircAuth := &slirc.IRCAuth{Target: "NickServ", Msg: "IDENTIFY BotNick Password"}

	conf := &slirc.Config{
		SlackBotToken:  "xoxb-0123456789-012345678901-5abCdefGhIjkLmN2OpqRSTuV",
		SlackUserToken: "xoxp-0123456789-012345678901-012345678901-2abcd3e45678901234fg5678901234hi",
		SlackChan:      "slackChan", //without # prefix

		IRCServer: "irc.freenode.net",
		IRCChan:   "#ircChanToLink",
		IRCNick:   "IRCNickname",
		IRCSSL:    true,
		IRCAuth:   ircAuth,
	}

	slirc.NewBridge(conf)

	select {}
}
```

### Example without IRC authentication

```go
package main

import (
        "github.com/simonkern/slirc"
)

// Example without IRC Authentication
// Slack Chan without "#"-prefix
func main() {
	
	conf := &slirc.Config{
		SlackBotToken:  "xoxb-0123456789-012345678901-5abCdefGhIjkLmN2OpqRSTuV",
		SlackUserToken: "xoxp-0123456789-012345678901-012345678901-2abcd3e45678901234fg5678901234hi",
		SlackChan:      "slackChan", //without # prefix

		IRCServer: "irc.freenode.net",
		IRCChan:   "#ircChanToLink",
		IRCNick:   "IRCNickname",
		IRCSSL:    true,
		IRCAuth:   nil,
	}

	slirc.NewBridge(conf)

	select {}
}
```
