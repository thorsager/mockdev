package mockssh

import (
	"fmt"
	"github.com/thorsager/mockdev/util"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Configuration struct {
	Name              string                 `yaml:"name"`
	BindAddr          string                 `yaml:"bind-addr"`
	ConversationFiles []string               `yaml:"conversation-files"`
	Conversations     []Conversation         `yaml:"conversations"`
	HostKeyPEM        []string               `yaml:"host-keys"`
	HostKeyFiles      []string               `yaml:"host-key-files"`
	Users             map[string]Credentials `yaml:"users"`
	DefaultPrompt     string                 `yaml:"default-prompt"`
	Motd              string                 `yaml:"motd"`
	Logging           SessionLogging         `yaml:"session-logging"`
}

type SessionLogging struct {
	LogReceived bool   `yaml:"log-received"`
	LogSent     bool   `yaml:"log-sent"`
	Location    string `yaml:"location"`
}

type Credentials struct {
	Password      string `yaml:"password"`
	AuthorizedKey string `yaml:"authorized-key"`
}

type Conversation struct {
	Name           string   `yaml:"name"`
	Order          int      `yaml:"match-order"`
	RequestMatcher string   `yaml:"request-matcher"`
	Response       Response `yaml:"response"`
}

type Response struct {
	Body                string `yaml:"body"`
	BodyFile            string `yaml:"body-file,omitempty"`
	Prompt              string `yaml:"prompt"`
	TerminateConnection bool   `yaml:"terminate-connection"`
}

func DecodeConversationFile(filename string) ([]Conversation, error) {
	cff, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to load file '%s':, %w", filename, err)
	}
	defer func() { _ = cff.Close() }()
	var vl []Conversation
	err = yaml.NewDecoder(cff).Decode(&vl)
	if err != nil {
		return nil, fmt.Errorf("unable to decode conversation in file '%s': %w", filename, err)
	}
	for i := 0; i < len(vl); i++ {
		if vl[i].Response.BodyFile != "" {
			vl[i].Response.BodyFile = util.MakeFileAbsolute(filepath.Dir(filename), vl[i].Response.BodyFile)
		}
	}
	return vl, nil
}
