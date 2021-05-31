package mockssh

import (
	"bufio"
	"fmt"
	"github.com/gliderlabs/ssh"
	"github.com/sirupsen/logrus"
	"github.com/thorsager/mockdev/logging"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	sync.Mutex
	Conversations      []Conversation
	Log                logging.Logger
	Users              map[string]Credentials
	DefaultPrompt      string
	MOTD               string
	SessionLogLocation string
	SessionLogReceived bool
	SessionLogSent     bool
	sessionCounter     int
}

var crlf = []byte{'\r', '\n'}

const keyEnter = '\r'
const newLine = '\n'

func (h *Handler) nextSession() int {
	h.Lock()
	defer h.Unlock()
	h.sessionCounter = h.sessionCounter + 1
	return h.sessionCounter
}

func (h *Handler) sessionFilename(sesId int) string {
	return path.Join(h.SessionLogLocation, fmt.Sprintf("sess_%d.log", sesId))
}

func (h *Handler) initSessionLog(sesId int) error {
	if h.SessionLogSent || h.SessionLogReceived {
		filename := h.sessionFilename(sesId)
		h.Log.Debugf("logging session to: %s", filename)
		err := os.MkdirAll(h.SessionLogLocation, 0770)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = f.WriteString(fmt.Sprintf("# session %d, %s \n", sesId, time.Now()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) sLog(isSend bool, sesId int, line string) error {
	if !h.SessionLogReceived && !h.SessionLogSent {
		return nil
	}
	if (isSend && !h.SessionLogSent) || (!isSend && !h.SessionLogReceived) {
		return nil
	}
	isDual := h.SessionLogSent && h.SessionLogReceived

	f, err := os.OpenFile(h.sessionFilename(sesId), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	if isDual {
		if isSend {
			line = "s: " + line
		} else {
			line = "r: " + line
		}
	}
	_, err = f.WriteString(line + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) write(s ssh.Session, id int, buf []byte) (int, error) {
	i, err := s.Write(buf)
	if err != nil {
		return i, err
	}
	err = h.sLog(true, id, string(buf))
	if err != nil {
		return i, err
	}
	return i, nil
}

func (h *Handler) handle(s ssh.Session) {
	sessionId := h.nextSession()
	err := h.initSessionLog(sessionId)
	if err != nil {
		h.Log.Errorf("sessionLogInit: %v", err)
		return
	}
	_, err = h.write(s, sessionId, append([]byte(h.MOTD), crlf...))
	if err != nil {
		h.Log.Errorf("writeMotd: %v", err)
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

		err = h.sLog(false, sessionId, string(line))
		if err != nil {
			h.Log.Errorf("while writing session log: %s", err)
			break
		}

		conv := h.findConversation(string(line))

		if conv == nil {
			h.Log.Warn("no conv, teapot?")
			_, err = h.write(s, sessionId, []byte("i'm no a teapot\n"))
			if err != nil {
				h.Log.Errorf("while writing: %s", err)
				break
			}
			_, _ = s.Write([]byte(h.DefaultPrompt))
		} else {
			if len(conv.Response.Body) > 0 {
				h.Log.Tracef("body: %s", conv.Response.Body)

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
			h.Log.Debugf("matched conv: %s", conv.Name)
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
			logrus.Warnf("unable to parse authorized key: %s", err)
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
