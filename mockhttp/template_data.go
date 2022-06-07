package mockhttp

import (
	"net"
	"os"
	"strings"
)

const cfg = "cfg"
const env = "env"
const run = "run"
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

func createRuntimeData() (templateData, error) {
	td := templateData{}
	ipv4, ipv6, err := listLocalAddresses()
	if err != nil {
		return td, err
	}
	td["ipv4"] = ipv4
	td["ipv6"] = ipv6
	return td, nil
}

func listLocalAddresses() ([]string, []string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, err
	}

	var ipv4 []string
	var ipv6 []string
	for _, intf := range interfaces {
		if intf.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if intf.Flags&net.FlagLoopback == 1 {
			continue // interface is loop back, don't care
		}
		addrs, err := intf.Addrs()
		if err != nil {
			return nil, nil, err
		}
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() == nil {
					ipv6 = append(ipv6, ipnet.IP.String())
				} else {
					ipv4 = append(ipv4, ipnet.IP.String())
				}
			}
		}
	}
	return ipv4, ipv6, nil
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
