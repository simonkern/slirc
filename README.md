# slirc
description coming soonish


### Example Usage

NewBridge has the following signature:

`
func NewBridge(slackToken, slackChannel, ircServer, ircChannel, ircNick string, ircSSL, insecureSkipVerify bool) (bridge *Bridge) {
`

`
package main

import (
        "github.com/simonkern/slirc"
)


// Slack Chan without "#"-prefix
func main() {
        slirc.NewBridge("SLACKTOKEN",
                "slackChan", "IRC-SERVER", "IRCChannel", "IRCNick", false)

        select {}
}
`
