package mockhttp

import (
	"fmt"
	"github.com/thorsager/mockdev/util"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

type Configuration struct {
	Name              string   `yaml:"name"`
	BindAddr          string   `yaml:"bind-addr"`
	ConversationFiles []string `yaml:"conversation-files"`
	Conversations     []Conversation
	Logging           SessionLogging `yaml:"session-logging"`
}

type SessionLogging struct {
	LogReceived bool   `yaml:"log-received"`
	Location    string `yaml:"location"`
}

type Conversation struct {
	Name        string   `yaml:"name"`
	Order       int      `yaml:"match-order"`
	Request     Request  `yaml:"request"`
	Response    Response `yaml:"response"`
	AfterScript []string `yaml:"after-script"`
}

type Request struct {
	UrlMatcher     UrlMatcher `yaml:"url-matcher"`
	MethodMatcher  string     `yaml:"method-matcher"`
	HeaderMatchers []string   `yaml:"header-matchers,omitempty"`
	BodyMatcher    string     `yaml:"body-matcher,omitempty"`
}

type Response struct {
	StatusCode int      `yaml:"status-code"`
	Headers    []string `yaml:"headers"`
	Body       string   `yaml:"body"`
	BodyFile   string   `yaml:"body-file,omitempty"`
}

type UrlMatcher struct {
	Path  string `yaml:"path,omitempty"`
	Query string `yaml:"query,omitempty"`
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
