package mockhttp

import (
	"os"
	"strings"
)

const cfg = "cfg"
const env = "env"
const envPrefix = "MOCKDEV_"
const currentTime = "currentTime"
const currentTimeGMT = "currentTime_GMT"

type templateData map[string]interface{}
type envData map[string]string

type templateConfigData struct {
	Address string
	Port    string
}

func createConfigData(addressPort string) templateConfigData {
	segs := strings.SplitN(addressPort, ":", 2)
	return templateConfigData{
		segs[0], segs[1],
	}
}

func createEnvData() envData {
	evd := make(envData)
	for _, tuple := range os.Environ() {
		segs := strings.SplitN(tuple, "=", 2)
		if strings.HasPrefix(segs[0], envPrefix) {
			evd[strings.TrimPrefix(segs[0], envPrefix)] = segs[1]
		}
	}
	return evd
}
