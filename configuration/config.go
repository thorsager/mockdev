package configuration

import (
	"github.com/thorsager/mockdev/util"
	"gopkg.in/yaml.v2"
	"os"
	"path"
)

func Read(filename string) (*Config, error) {
	// TODO: add deep validation of arguments, that the server may fail EARLY
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	config := &Config{}
	err = yaml.NewDecoder(file).Decode(config)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(config.Snmp); i++ {
		config.Snmp[i].SnapshotFiles = util.MakeFilesAbsolute(path.Dir(filename), config.Snmp[i].SnapshotFiles)
	}
	for i := 0; i < len(config.Http); i++ {
		config.Http[i].ConversationFiles = util.MakeFilesAbsolute(path.Dir(filename), config.Http[i].ConversationFiles)
	}
	for i := 0; i < len(config.Ssh); i++ {
		config.Ssh[i].ConversationFiles = util.MakeFilesAbsolute(path.Dir(filename), config.Ssh[i].ConversationFiles)
	}
	return config, nil
}
