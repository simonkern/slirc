package slirc

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"
	"time"

	ircc "github.com/fluffle/goirc/client"

	"github.com/simonkern/slirc/slack"
)

// Bridge links an irc and a slack channel
type Bridge struct {
	SlackChan string
	IRCChan   string
	slack     *slack.Client
	irc       *ircc.Conn
}

type messager interface {
	Usernick() string
	Msg() string
	Chan() string
}

// IRCAuth stores authentification target and the message that needs to be send in order to auth
// e.g. "NickServ" and "IDENTIFY fooBarPassword"
type IRCAuth struct {
	Target string
	Msg    string
}

// NewBridge instantiates a Bridge object and sets up the required irc and slack clients
func NewBridge(slackBotToken, slackUserToken, slackChannel, ircServer, ircChannel, ircNick string, ircSSL bool, tlsConfig *tls.Config, ircAuth *IRCAuth) (bridge *Bridge) {
	sc := slack.NewClient(slackBotToken)

	sc.UserToken = slackUserToken

	ircCfg := ircc.NewConfig(ircNick, "slirc", "Powered by Slirc")
	ircCfg.QuitMessage = "Slack <-> IRC Bridge shutting down"
	ircCfg.Server = ircServer
	ircCfg.NewNick = func(n string) string {
		if n != ircNick && len(n) > len(ircNick)+2 {
			return ircNick
		}
		return n + "_"
	}
	if ircSSL {
		ircCfg.SSL = true
		if tlsConfig != nil {
			ircCfg.SSLConfig = tlsConfig
		}
	}
	c := ircc.Client(ircCfg)

	bridge = &Bridge{SlackChan: slackChannel, IRCChan: ircChannel, slack: sc, irc: c}

	// IRC Handlers
	c.HandleFunc(ircc.CONNECTED,
		func(conn *ircc.Conn, line *ircc.Line) {
			if ircAuth != nil {
				log.Println("IRC Authentication")
				<-time.After(5 * time.Second)
				conn.Privmsg(ircAuth.Target, ircAuth.Msg)
				<-time.After(3 * time.Second)
			}
			conn.Join(ircChannel)
			bridge.slack.Send(bridge.SlackChan, "Connected to IRC.")
			log.Println("Connected to IRC.")
		})

	c.HandleFunc(ircc.DISCONNECTED,
		func(conn *ircc.Conn, line *ircc.Line) {
			bridge.slack.Send(bridge.SlackChan, "Disconnected from IRC. Reconnecting...")
			log.Println("Disconnected from IRC. Reconnecting...")
			for {
				if err := conn.Connect(); err != nil {
					log.Println("IRC reconnect failed: ", err)
					log.Println("Trying again in 30 seconds...")
					time.Sleep(30 * time.Second)
					continue
				}
				// success
				break
			}
		})

	c.HandleFunc(ircc.PRIVMSG,
		func(conn *ircc.Conn, line *ircc.Line) {
			if line.Target() == bridge.IRCChan {
				msg := fmt.Sprintf("[%s]: %s", line.Nick, line.Text())
				bridge.slack.Send(bridge.SlackChan, msg)
			}
		})

	// thanks jn__
	c.HandleFunc(ircc.ACTION,
		func(conn *ircc.Conn, line *ircc.Line) {
			if line.Target() == bridge.IRCChan {
				msg := fmt.Sprintf(" * %s %s", line.Nick, line.Text())
				bridge.slack.Send(bridge.SlackChan, msg)
			}
		})

	// Slack Handlers
	sc.HandleFunc("shutdown",
		func(sc *slack.Client, e *slack.Event) {
			bridge.irc.Privmsg(bridge.IRCChan, "Shutting down slack client")
			log.Println("Shutting down slack client")

		})

	sc.HandleFunc("disconnected",
		func(sc *slack.Client, e *slack.Event) {
			bridge.irc.Privmsg(bridge.IRCChan, "Disconnected from Slack. Reconnecting...")
			log.Println("Disconnected from Slack. Reconnecting...")
			sc.Connect()

		})

	sc.HandleFunc("connected",
		func(sc *slack.Client, e *slack.Event) {
			bridge.irc.Privmsg(bridge.IRCChan, "Connected to Slack.")
			log.Println("Connected to Slack.")
		})

	sc.HandleFunc("message",
		func(sc *slack.Client, e *slack.Event) {
			if e.Chan() == bridge.SlackChan && !sc.IsSelfMsg(e) && e.Text != "" && e.Usernick() != "rwthirc" {
				msg := fmt.Sprintf("[%s]: %s", e.Usernick(), e.Msg())
				// IRC has problems with newlines, therefore we split the message
				for _, line := range strings.SplitAfter(msg, "\n") {
					// we do not want to send empty lines...
					if strings.TrimSpace(line) != "" {
						bridge.irc.Privmsg(bridge.IRCChan, line)
					}
				}
			}

		})

	go func() {
		if err := c.Connect(); err != nil {
			log.Fatal("Could not connect to IRC: ", err)
		}
	}()
	go func() {
		if err := sc.Connect(); err != nil {
			log.Fatal("Could not connect to Slack: ", err)
		}
	}()
	return bridge
}
