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
		
		ircc "github.com/fluffle/goirc/client"
)

// Example with IRC Authentication
// Slack Chan without "#"-prefix
func main() {

	postConnect := func(ic *ircc.Conn, c *slirc.Config) {
		log.Println("IRC PostConnect Action")
		<-time.After(5 * time.Second)
		log.Println("Authenticating with Nickserv")
		ic.Privmsg("NickServ", "IDENTIFY BotNick Password")
		<-time.After(3 * time.Second)
	}

	conf := &slirc.Config{
		SlackBotToken:  "xoxb-0123456789-012345678901-5abCdefGhIjkLmN2OpqRSTuV",
		SlackUserToken: "xoxp-0123456789-012345678901-012345678901-2abcd3e45678901234fg5678901234hi",
		SlackChan:      "slackChan", //without # prefix

		IRCServer: "irc.freenode.net",
		IRCChan:   "#ircChanToLink",
		IRCNick:   "IRCNickname",
		IRCSSL:    true,
		IRCPostConnect: postConnect,
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
		
		ircc "github.com/fluffle/goirc/client"
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
		IRCPostConnect: nil,
	}

	slirc.NewBridge(conf)

	select {}
}
```
