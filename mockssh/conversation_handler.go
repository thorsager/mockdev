package mockssh

import (
	"bufio"
	"github.com/gliderlabs/ssh"
	"github.com/sirupsen/logrus"
	"github.com/thorsager/mockdev/logging"
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

var crlf = []byte{'\r', '\n'}

const keyEnter = '\r'
const newLine = '\n'

func (h *Handler) handle(s ssh.Session) {
	_, err := s.Write(append([]byte(h.MOTD), crlf...))
	if err != nil {
		h.Log.Error(err)
		return
	}

	_, _ = s.Write([]byte(h.DefaultPrompt))

	br := bufio.NewReader(s)

	for {
		var line []byte
		var err error
		for {
			var b byte
			b, err = br.ReadByte()
			if err != nil {
				break
			}
			if b == keyEnter || b == newLine {
				_, _ = s.Write(crlf)
				break
			} else {
				_, _ = s.Write([]byte{b})
				line = append(line, b)
				h.Log.Tracef("line: %s", line)
			}
		}
		if err != nil {
			h.Log.Errorf("while reading: %s", err)
			break
		}
		h.Log.Debugf("Got a full line: %s", line)

		conv := h.findConversation(string(line))

		if conv == nil {
			h.Log.Warn("no conv, teapot?")
			_, err = s.Write([]byte("i'm no a teapot\n"))
			_, _ = s.Write([]byte(h.DefaultPrompt))
			if err != nil {
				h.Log.Errorf("while writing: %s", err)
				break
			}
		} else {
			if len(conv.Response.Body) > 0 {
				h.Log.Debugf("body: %s", conv.Response.Body)
				_, err = s.Write(append([]byte(conv.Response.Body), crlf...))
				if err != nil {
					h.Log.Errorf("while writing: %s", err)
					break
				}
			}
			if conv.Response.TerminateConnection {
				h.Log.Info("connection terminated by user.")
				break
			}
			if conv.Response.Prompt != "" {
				_, _ = s.Write([]byte(conv.Response.Prompt))
			} else {
				_, _ = s.Write([]byte(h.DefaultPrompt))
			}
		}
	}
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
