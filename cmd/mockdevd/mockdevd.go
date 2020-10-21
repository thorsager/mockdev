package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/thorsager/mockdev/configuration"
	"github.com/thorsager/mockdev/mockhttp"
	"github.com/thorsager/mockdev/mocksnmp"
	"github.com/thorsager/mockdev/mockssh"
	"net/http"
	"os"
)

var Version = "*unset*"

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	var configFile string
	flag.StringVar(&configFile, "c", "config.yaml", "configuration file")

	flag.Parse()

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.Infof("reading config from: %s", configFile)
	config, err := configuration.Read(configFile)
	if err != nil {
		logger.Fatalf("while reading config: %v", err)
	}
	logger.Infof("mockdevd v%s starting", Version)
	for _, c := range config.Snmp {
		entry := logger.WithField("type", "snmp")
		go startSnmpService(c, entry)
	}

	for _, c := range config.Http {
		entry := logger.WithField("type", "http")
		go startHttpService(c, entry)
	}

	for _, c := range config.Ssh {
		entry := logger.WithField("type", "ssh")
		go startSshService(c, entry)
	}

	// this could be done a lot nicer...
	select {}
}

func startSshService(config *mockssh.Configuration, logger *logrus.Entry) {
	logger.Infof("Server %s listening on %s", config.Name, config.BindAddr)
	s, err := mockssh.NewServer(config, logger)
	if err != nil {
		logger.Error(err)
	}
	err = s.ListenAndServe()
	if err != nil {
		logger.Error(err)
	}
}

func startHttpService(config *mockhttp.Configuration, logger *logrus.Entry) {
	conversations := config.Conversations
	for _, cf := range config.ConversationFiles {
		// this is where we load the reset of the conversations
		con, err := mockhttp.DecodeConversationFile(cf)
		if err != nil {
			logger.Error(err)
		} else {
			conversations = append(conversations, *con)
		}
	}
	logger.Infof("Server %s listening on %s", config.Name, config.BindAddr)
	err := http.ListenAndServe(config.BindAddr, mockhttp.ConversationsHandler{Conversations: conversations, Log: logger})
	if err != nil {
		logger.Error(err)
	}
}

func startSnmpService(config *mocksnmp.Configuration, logger *logrus.Entry) {
	server, err := mocksnmp.NewServer(config, logger.WithField("name", config.Name))
	if err != nil {
		logger.Fatalf("while creating server: %v", err)
	}
	err = server.ListenUDP("udp", config.BindAddr)
	if err != nil {
		logger.Fatalf("while setting up socket: %v", err)
	}
	logger.Infof("snmp service '%s' listening on %s (ro=%s,rw=%s)", config.Name, config.BindAddr, config.ReadCommunity, config.WriteCommunity)
	err = server.ServeForever()
	if err != nil {
		logger.Fatalf("while serving: %v", err)
	}
}
