package mockssh

import (
	"github.com/gliderlabs/ssh"
	"github.com/thorsager/mockdev/logging"
	"sort"
)

func NewServer(config *Configuration, logger logging.Logger) (*ssh.Server, error) {
	conversations := config.Conversations

	for _, cf := range config.ConversationFiles {
		// this is where we load the reset of the conversations
		con, err := DecodeConversationFile(cf)
		if err != nil {
			logger.Error(err)
		} else {
			conversations = append(conversations, con...)
		}
	}
	sort.Slice(conversations, func(i, j int) bool { return conversations[i].Order < conversations[j].Order })
	for _, c := range conversations {
		logger.Infof("loaded conversation[%d]: %s", c.Order, c.Name)
	}

	handler := Handler{Conversations: conversations,
		Log:                logger,
		Users:              config.Users,
		DefaultPrompt:      config.DefaultPrompt,
		MOTD:               config.Motd,
		SessionLogLocation: config.Logging.Location,
		SessionLogSent:     config.Logging.LogSent,
		SessionLogReceived: config.Logging.LogReceived,
	}
	s := &ssh.Server{
		Addr:             config.BindAddr,
		Handler:          handler.handle,
		PublicKeyHandler: handler.publicKeyHandler,
		PasswordHandler:  handler.passwordHandler,
	}
	for _, file := range config.HostKeyFiles {
		logger.Infof("loading key-file: %s", file)
		err := s.SetOption(ssh.HostKeyFile(file))
		if err != nil {
			logger.Warnf("while loading key-file (%s): %v", file, err)
		}
	}
	for _, key := range config.HostKeyPEM {
		logger.Info(key)
		err := s.SetOption(ssh.HostKeyPEM([]byte(key)))
		if err != nil {
			logger.Warnf("while loading key: %v", err)
		}
	}
	return s, nil
}
