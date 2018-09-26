# slirc

[![Build Status](https://travis-ci.org/simonkern/slirc.svg)](https://travis-ci.org/simonkern/slirc)

Slirc links an IRC and a Slack channel.

## Example Usage

NewBridge has the following signature:

```go
func NewBridge(slackBotToken, slackUserToken, slackChannel, ircServer, ircChannel, ircNick string, ircSSL bool, tlsConfig *tls.Config, ircAuth *IRCAuth) (bridge *Bridge)
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
        ircAuth := &slirc.IRCAuth{Target: "NickServ", Msg: "IDENTIFY FooUser BarPassword"}
        slirc.NewBridge("SlackKBotToken", "SlackUserToken",
                "slackChan", "irc.freenode.net", "IRCChannel", "IRCNick", true,  &tls.Config{ServerName: "irc.freenode.net"}, ircAuth)

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
        slirc.NewBridge("SlackKBotToken", "SlackUserToken",
                "slackChan", "irc.freenode.net", "IRCChannel", "IRCNick", true, &tls.Config{ServerName: "irc.freenode.net"}, nil)

        select {}
}
```
