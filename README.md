# slirc

[![Build Status](https://travis-ci.org/simonkern/slirc.svg)](https://travis-ci.org/simonkern/slirc)

Slirc links an IRC and a Slack channel.

## Example Usage

NewBridge has the following signature:

```go
func NewBridge(slackToken, slackChannel, ircServer, ircChannel, ircNick string, ircSSL, insecureSkipVerify bool, ircAuth *IRCAuth) (bridge *Bridge)
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
        slirc.NewBridge("SLACKTOKEN",
                "slackChan", "IRC-SERVER", "IRCChannel", "IRCNick", true, true, ircAUTH)

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
        slirc.NewBridge("SLACKTOKEN",
                "slackChan", "IRC-SERVER", "IRCChannel", "IRCNick", true, true, nil)

        select {}
}
```
