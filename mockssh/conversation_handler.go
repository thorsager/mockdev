package mockssh

import (
	"github.com/gliderlabs/ssh"
	"github.com/sirupsen/logrus"
	"github.com/thorsager/mockdev/logging"
	"golang.org/x/crypto/ssh/terminal"
	"regexp"
	"strings"
)

type Handler struct {
	Conversations []Conversation
	Log           logging.Logger
	Users         map[string]Credentials
	DefaultPrompt string
	MOTD          string
}

func (h *Handler) handle(s ssh.Session) {
	term := terminal.NewTerminal(s, h.DefaultPrompt)
	_, err := term.Write(append([]byte(h.MOTD), '\n'))
	if err != nil {
		h.Log.Errorf("while writing motd: %s", err)
		return
	}
	for {
		line, err := term.ReadLine()
		if err != nil {
			h.Log.Errorf("while reading: %s", err)
			break
		}
		conv := h.findConversation(line)

		if conv == nil {
			_, err := term.Write([]byte("I'm no a teapot\n"))
			if err != nil {
				h.Log.Errorf("while writing: %s", err)
				break
			}
			term.SetPrompt(h.DefaultPrompt)
		} else {
			if conv.Response.Prompt != "" {
				term.SetPrompt(conv.Response.Prompt)
			} else {
				term.SetPrompt(h.DefaultPrompt)
			}
			if len(conv.Response.Body) > 0 {
				_, err := term.Write(append([]byte(conv.Response.Body), '\n'))
				if err != nil {
					h.Log.Errorf("while writing: %s", err)
					break
				}
			}
			if conv.Response.TerminateConnection {
				h.Log.Info("connection terminated by user.")
				break
			}
		}

	}
	h.Log.Info("terminal closed")
}

func (h *Handler) findConversation(line string) *Conversation {
	convkey := strings.TrimSpace(line)
	for _, conv := range h.Conversations {
		matcher := regexp.MustCompile(conv.RequestMatcher)
		if matcher.MatchString(convkey) {
			h.Log.Infof("matched conv: %s", conv.Name)
			return &conv
		}
	}
	return nil
}

func (h *Handler) passwordHandler(ctx ssh.Context, password string) bool {
	if cred, found := h.Users[ctx.User()]; found {
		if password == cred.Password {
			logrus.Infof("user %s authenticated using password", ctx.User())
			return true
		}
		logrus.Infof("password authentication failed for %s@%s", ctx.User(), ctx.RemoteAddr())
	}
	return false
}

func (h *Handler) publicKeyHandler(ctx ssh.Context, pk ssh.PublicKey) bool {
	if cred, found := h.Users[ctx.User()]; found {
		if cred.AuthorizedKey == "" {
			return false
		}
		ak, _, _, _, err := ssh.ParseAuthorizedKey([]byte(cred.AuthorizedKey))
		if err != nil {
			logrus.Warn("unable to parse authorized key: %s", err)
			return false
		}
		if ssh.KeysEqual(ak, pk) {
			logrus.Infof("user %s authenticated using public key", ctx.User())
			return true
		} else {
			logrus.Infof("public-key authentication failed for %s@%s", ctx.User(), ctx.RemoteAddr())
		}
	}
	return false // allow all keys, or use ssh.KeysEqual() to compare against known keys
}
