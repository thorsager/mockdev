package mocksnmp

//import "github.com/slayercat/gosnmp"
import (
	"fmt"
	"github.com/slayercat/GoSNMPServer"
	"github.com/slayercat/gosnmp"
	"github.com/thorsager/mockdev/logging"
	"github.com/thorsager/mockdev/snmpsup"
	"io/ioutil"
	"net"
	"strings"
)

func parseConfigOIDs(config *Configuration) ([]*GoSNMPServer.PDUValueControlItem, error) {
	var controlItems []*GoSNMPServer.PDUValueControlItem
	if cs, err := parseWalkFiles(config.SnapshotFiles); err == nil {
		controlItems = append(controlItems, cs...)
	} else {
		return nil, err
	}

	for _, str := range config.OIDs {
		ci, err := ControlItemFromString(str)
		if err != nil {
			return nil, err
		}
		controlItems = append(controlItems, ci)
	}

	return uniqueLast(controlItems), nil
}

func uniqueLast(items []*GoSNMPServer.PDUValueControlItem) []*GoSNMPServer.PDUValueControlItem {
	buf := make(map[string]*GoSNMPServer.PDUValueControlItem)
	for _, item := range items {
		buf[item.OID] = item
	}
	var finalBuf []*GoSNMPServer.PDUValueControlItem
	for _, v := range buf {
		finalBuf = append(finalBuf, v)
	}
	return finalBuf
}

func NewServer(config *Configuration, logger logging.Logger) (*GoSNMPServer.SNMPServer, error) {
	cis, err := parseConfigOIDs(config)
	if err != nil {
		return nil, err
	}
	master := GoSNMPServer.MasterAgent{
		Logger: logger,
		SecurityConfig: GoSNMPServer.SecurityConfig{
			AuthoritativeEngineBoots: 1,
		},
		SubAgents: []*GoSNMPServer.SubAgent{
			{
				CommunityIDs: []string{config.ReadCommunity, config.WriteCommunity},
				OIDs:         cis,
			},
		},
	}
	return GoSNMPServer.NewSNMPServer(master), nil
}

func parseWalkFiles(files []string) ([]*GoSNMPServer.PDUValueControlItem, error) {
	var oids []*GoSNMPServer.PDUValueControlItem
	for _, file := range files {
		os, err := parseSnapshotFile(file)
		if err != nil {
			return nil, err
		}
		oids = append(oids, os...)
	}
	return oids, nil
}

func parseSnapshotFile(file string) ([]*GoSNMPServer.PDUValueControlItem, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var oids []*GoSNMPServer.PDUValueControlItem
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
			continue
		}
		ci, err := ControlItemFromString(line)
		if err != nil {
			return nil, err
		}
		oids = append(oids, ci)
	}
	return oids, nil
}

func ControlItemFromString(snapshotString string) (*GoSNMPServer.PDUValueControlItem, error) {
	nPDU, err := snmpsup.ParseNeutralPDU(snapshotString)
	if err != nil {
		return nil, err
	}
	return valueControlItem(nPDU)
}

func valueControlItem(npdu *snmpsup.NeutralPDU) (*GoSNMPServer.PDUValueControlItem, error) {
	return &GoSNMPServer.PDUValueControlItem{
		OID:  npdu.Oid,
		Type: gosnmp.Asn1BER(npdu.Asn1BER),
		OnGet: func() (value interface{}, err error) {
			return wrappedValue(npdu)
		},
	}, nil
}

func wrappedValue(npdu *snmpsup.NeutralPDU) (interface{}, error) {
	switch t := gosnmp.Asn1BER(npdu.Asn1BER); t {
	case gosnmp.IPAddress:
		ip := net.ParseIP(npdu.Value.(string))
		return GoSNMPServer.Asn1IPAddressWrap(ip), nil
	case gosnmp.OctetString:
		return GoSNMPServer.Asn1OctetStringWrap(npdu.Value.(string)), nil
	case gosnmp.Integer:
		return GoSNMPServer.Asn1IntegerWrap(npdu.Value.(int)), nil
	case gosnmp.Gauge32:
		return GoSNMPServer.Asn1Gauge32Wrap(npdu.Value.(uint)), nil
	case gosnmp.Counter32:
		return GoSNMPServer.Asn1Counter32Wrap(npdu.Value.(uint)), nil
	case gosnmp.Counter64:
		return GoSNMPServer.Asn1Counter64Wrap(npdu.Value.(uint64)), nil
	case gosnmp.ObjectIdentifier:
		return GoSNMPServer.Asn1ObjectIdentifierWrap(npdu.Value.(string)), nil
	case gosnmp.TimeTicks:
		return GoSNMPServer.Asn1TimeTicksWrap(npdu.Value.(uint32)), nil
	default:
		return nil, fmt.Errorf("unable to wrap value of type %x", t)
	}
}
