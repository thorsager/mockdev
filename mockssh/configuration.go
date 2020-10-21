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
}

type Credentials struct {
	Password      string `yaml:"password"`
	AuthorizedKey string `yaml:"authorized-key"`
}

type Conversation struct {
	Name           string   `yaml:"name"`
	RequestMatcher string   `yaml:"request-matcher"`
	Response       Response `yaml:"response"`
}

type Response struct {
	Body                string `yaml:"body"`
	BodyFile            string `yaml:"body-file,omitempty"`
	Prompt              string `yaml:"prompt"`
	TerminateConnection bool   `yaml:"terminate-connection"`
}

func DecodeConversationFile(filename string) (*Conversation, error) {
	cff, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to load file '%s':, %w", filename, err)
	}
	defer func() { _ = cff.Close() }()
	var v Conversation
	err = yaml.NewDecoder(cff).Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("unable to decode conversation in file '%s': %w", filename, err)
	}
	if v.Response.BodyFile != "" {
		v.Response.BodyFile = util.MakeFileAbsolute(filepath.Dir(filename), v.Response.BodyFile)
	}
	return &v, nil
}
