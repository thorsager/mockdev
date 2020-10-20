package configuration

import (
	"github.com/thorsager/mockdev/mockhttp"
	"github.com/thorsager/mockdev/mocksnmp"
)

type Config struct {
	Snmp []*mocksnmp.Configuration `yaml:"snmp"`
	Http []*mockhttp.Configuration `yaml:"http"`
}
