package main

import (
	"flag"
	"fmt"
	"github.com/gosnmp/gosnmp"
	"github.com/thorsager/mockdev/snmpsup"
	"github.com/thorsager/mockdev/spinner"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var Version = "*unset*"

func main() {
	flag.Usage = func() {
		bin := filepath.Base(os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "%s Version %s\n", bin, Version)
		_, _ = fmt.Fprintln(os.Stderr, "Usage:")
		_, _ = fmt.Fprintf(os.Stderr, "  %s [-c <community>] [-n] host[:port] [oid]\n", bin)
		_, _ = fmt.Fprintln(os.Stderr, "  Options:")
		_, _ = fmt.Fprintln(os.Stderr, "    -c <community>   Community string for walking device")
		_, _ = fmt.Fprintln(os.Stderr, "    -n               No bulk-walk (use plain walk)")
		_, _ = fmt.Fprintln(os.Stderr, "    -v               Verbose, print out progress")
		_, _ = fmt.Fprintln(os.Stderr, "    -o <file>        Name of output file (default: '-', STDOUT)")
		_, _ = fmt.Fprintln(os.Stderr, "    -f               Overwrite output-file, if exists")
		_, _ = fmt.Fprintln(os.Stderr, "  Arguments:")
		_, _ = fmt.Fprintln(os.Stderr, "    host[:port]      name/ip, and optional port of host to be walked")
		_, _ = fmt.Fprintln(os.Stderr, "    oid              starting oid")

	}
	var community string
	flag.StringVar(&community, "c", "public", "Community string for walking device")
	var output string
	flag.StringVar(&output, "o", "-", "Name of output file (default: '-', STDOUT)")
	var noBulk bool
	flag.BoolVar(&noBulk, "n", false, "No bulk-walk (use plain walk)")
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Verbose, print out progress")
	var overwrite bool
	flag.BoolVar(&overwrite, "f", false, "Overwrite output-file, if exists")
	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	target := flag.Arg(0)

	var oid string
	if len(flag.Args()) > 1 {
		oid = flag.Arg(1)
	}

	var port uint16 = 161
	if s := strings.SplitN(target, ":", 2); len(s) > 1 {
		target = s[0]
		x, err := strconv.ParseUint(s[1], 10, 16)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "invalid port number: %s", s[1])
		}
		port = uint16(x)
	}

	gosnmp.Default.Target = target
	gosnmp.Default.Port = port
	gosnmp.Default.Community = community
	gosnmp.Default.Timeout = time.Second * 10

	var err error
	var out = os.Stdout
	if output != "-" {
		if _, err = os.Stat(output); err == nil && overwrite == false {
			fmt.Printf("File '%s' already exists (use -f to overwrite)\n", output)
			os.Exit(1)
		}
		_, _ = fmt.Fprintf(os.Stderr, "Writing snapshot to '%s'\n", output)
		out, err = os.Create(output)
		if err != nil {
			fmt.Printf("Conect error: %s", err)
			os.Exit(1)
		}
		defer func() { _ = out.Close() }()
	}

	if verbose {
		_, _ = fmt.Fprintf(os.Stderr, "Creating snapshot of '%s'@%s:%d\n", community, target, port)
	}
	err = gosnmp.Default.Connect()
	if err != nil {
		fmt.Printf("Conect error: %s", err)
		os.Exit(1)
	}
	defer func() { _ = gosnmp.Default.Conn.Close() }()

	if verbose {
		_, _ = fmt.Fprintln(os.Stderr, "Got connection..")
	}

	s := &spinner.Spinner{}
	pduc := &snmpsup.PduCollector{Writer: out}
	if verbose {
		pduc.OnCollect = func(pdu snmpsup.NeutralPDU) bool {
			s.Increment()
			return true
		}
	}
	if noBulk {
		err = gosnmp.Default.Walk(oid, collectAsNeutralPdu(pduc))
	} else {
		err = gosnmp.Default.BulkWalk(oid, collectAsNeutralPdu(pduc))
	}
	if err != nil {
		fmt.Printf("Walk error: %s", err)
		os.Exit(1)
	}
}

func collectAsNeutralPdu(collector *snmpsup.PduCollector) func(pdu gosnmp.SnmpPDU) error {
	return func(pdu gosnmp.SnmpPDU) error {
		return collector.Collect(snmpsup.NeutralPDU{Oid: pdu.Name, Asn1BER: byte(pdu.Type), Value: pdu.Value})
	}
}
