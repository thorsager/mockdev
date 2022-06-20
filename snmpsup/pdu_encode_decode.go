package snmpsup

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var ParseError = errors.New("parse error")

type NeutralPDU struct {
	Oid     string
	Asn1BER byte
	Value   interface{}
}

func (p *NeutralPDU) String() string {
	var tp string
	var v string
	switch t := p.Value.(type) {
	case string:
		tp, v = encodeString(p.Value.(string))
	case []uint8:
		b := p.Value.([]uint8)
		s := string(b)
		tp, v = encodeString(s)
	default:
		v = fmt.Sprintf("%v", p.Value)
		tp = fmt.Sprintf("%T", t)
	}
	return fmt.Sprintf("%s/%d/%s/%s", p.Oid, p.Asn1BER, tp, v)
}
func encodeString(s string) (string, string) {
	if isAsciiPrintable(s) {
		return "string", s
	}
	return "hex-string", hex.EncodeToString([]byte(s))
}

func ParseNeutralPDU(s string) (*NeutralPDU, error) {
	parts := strings.SplitN(s, "/", 4)
	if len(parts) != 4 {
		return nil, fmt.Errorf("%w; invalid number of parts: '%s'", ParseError, s)
	}
	asn1BER, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w; invalid asn1BER: %s", ParseError, err.Error())
	}
	if asn1BER < 0x00 && asn1BER > 0xff {
		return nil, fmt.Errorf("%w; invalid asn1BER: value out of range %d", ParseError, asn1BER)
	}

	value, err := typedDecode(parts[2], parts[3])
	if err != nil {
		return nil, fmt.Errorf("%w; while decoding: %s", ParseError, err.Error())
	}

	n := &NeutralPDU{Oid: parts[0], Asn1BER: byte(asn1BER), Value: value}
	return n, nil
}

func typedDecode(t string, v string) (interface{}, error) {
	var decVal interface{}
	var err error
	switch t {
	case "uint":
		var x uint64
		x, err = strconv.ParseUint(v, 10, 32)
		decVal = uint(x)
	case "uint32":
		var x uint64
		x, err = strconv.ParseUint(v, 10, 32)
		decVal = uint32(x)
	case "uint64":
		decVal, err = strconv.ParseUint(v, 10, 64)
	case "int":
		decVal, err = strconv.Atoi(v)
	case "string":
		decVal, err = v, nil
	case "hex-string":
		decVal, err = hexDecodeToString(v)
	default:
		err = fmt.Errorf("unknow type: %s", t)
	}
	if err != nil {
		return nil, fmt.Errorf("%w; unable to decode: %s/%s", err, t, v)
	}
	return decVal, err
}

func hexDecodeToString(s string) (string, error) {
	buf, err := hex.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

type PduCollector struct {
	pdus      []NeutralPDU
	OnCollect func(pdu NeutralPDU) bool
	Writer    io.Writer
}

func (c *PduCollector) Collect(pdu NeutralPDU) error {
	var ok bool
	if c.OnCollect != nil {
		ok = c.OnCollect(pdu)
	}
	if ok {
		if c.Writer != nil {
			_, err := c.Writer.Write([]byte(pdu.String() + "\n"))
			if err != nil {
				return err
			}
		} else {
			c.pdus = append(c.pdus, pdu)
		}
	}
	return nil
}

func (c *PduCollector) AsSlice() []NeutralPDU {
	return c.pdus
}

func isAsciiPrintable(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 32 || s[i] > 126 {
			return false
		}
	}
	return true
}
