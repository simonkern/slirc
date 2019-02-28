package slirc

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
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

type Config struct {
	SlackBotToken  string
	SlackUserToken string
	SlackChan      string

	IRCServer string
	IRCChan   string
	IRCNick   string
	IRCSSL    bool
	IRCAuth   *IRCAuth
}

// NewBridge instantiates a Bridge object and sets up the required irc and slack clients
func NewBridge(c *Config) (bridge *Bridge) {
	sc := slack.NewClient(c.SlackBotToken)

	sc.UserToken = c.SlackUserToken

	ircCfg := ircc.NewConfig(c.IRCNick, "slirc", "Powered by Slirc")
	ircCfg.QuitMessage = "Slack <-> IRC Bridge shutting down"
	ircCfg.Server = c.IRCServer
	ircCfg.NewNick = func(n string) string {
		if n != c.IRCNick && len(n) > len(c.IRCNick)+3 {
			return c.IRCNick
		}
		return n + "_"
	}
	if c.IRCSSL {
		ircCfg.SSL = true
		ircCfg.SSLConfig = &tls.Config{ServerName: c.IRCServer}
	}
	ic := ircc.Client(ircCfg)

	bridge = &Bridge{SlackChan: c.SlackChan, IRCChan: c.IRCChan, slack: sc, irc: ic}

	// IRC Handlers
	ic.HandleFunc(ircc.CONNECTED,
		func(conn *ircc.Conn, line *ircc.Line) {
			if c.IRCAuth != nil {
				log.Println("IRC Authentication")
				<-time.After(5 * time.Second)
				conn.Privmsg(c.IRCAuth.Target, c.IRCAuth.Msg)
				<-time.After(3 * time.Second)
			}
			conn.Join(c.IRCChan)
			bridge.slack.Send(bridge.SlackChan, "Connected to IRC.")
			log.Println("Connected to IRC.")
		})

	ic.HandleFunc(ircc.DISCONNECTED,
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

	ic.HandleFunc(ircc.PRIVMSG,
		func(conn *ircc.Conn, line *ircc.Line) {
			if line.Target() == bridge.IRCChan {
				msg := fmt.Sprintf("[%s]: %s", line.Nick, line.Text())
				bridge.slack.Send(bridge.SlackChan, msg)
			}
		})

	// thanks jn__
	ic.HandleFunc(ircc.ACTION,
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
			for {
				if err := sc.Connect(); err != nil {
					log.Println("Slack reconnect failed: ", err)
					log.Println("Trying again in 30 seconds...")
					time.Sleep(30 * time.Second)
					continue
				}
				// success
				break
			}
		})

	sc.HandleFunc("connected",
		func(sc *slack.Client, e *slack.Event) {
			bridge.irc.Privmsg(bridge.IRCChan, "Connected to Slack.")
			log.Println("Connected to Slack.")
		})

	sc.HandleFunc("admincommand",
		func(sc *slack.Client, e *slack.Event) {
			if e.Msg() == "die" {
				os.Exit(0)
			}
		})

	sc.HandleFunc("message",
		func(sc *slack.Client, e *slack.Event) {
			if e.Chan() == bridge.SlackChan && !sc.IsSelfMsg(e) && e.Text != "" {
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
		if err := ic.Connect(); err != nil {
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
