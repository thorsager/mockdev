package main

import (
	"flag"
	"fmt"
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/thorsager/mockdev/mockhttp"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var Version = "*unset*"

type headerList []string

func (l *headerList) Set(s string) error {
	*l = append(*l, s)
	return nil
}
func (l *headerList) String() string {
	return strings.Join(*l, ", ")
}

func main() {
	flag.Usage = func() {
		bin := filepath.Base(os.Args[0])
		_, _ = fmt.Fprintf(os.Stderr, "%s Version %s\n", bin, Version)
		_, _ = fmt.Fprintln(os.Stderr, "Usage:")
		_, _ = fmt.Fprintf(os.Stderr, "  %s [options] <url>\n", bin)
		_, _ = fmt.Fprintln(os.Stderr, "  Options:")
		_, _ = fmt.Fprintln(os.Stderr, "    -o <file>        Name of output file (default: '-', STDOUT)")
		_, _ = fmt.Fprintln(os.Stderr, "    -v               Verbose, print out progress")
		_, _ = fmt.Fprintln(os.Stderr, "    -f               Overwrite output-file, if exists")
		_, _ = fmt.Fprintln(os.Stderr, "    -X <method>      HTTP method to use for request (default: GET)")
		_, _ = fmt.Fprintln(os.Stderr, "    -H <header>      HTTP header(s)")
		_, _ = fmt.Fprintln(os.Stderr, "    -d <data>        Request content")
		_, _ = fmt.Fprintln(os.Stderr, "  Arguments:")
		_, _ = fmt.Fprintln(os.Stderr, "    url              Url to dump")
	}
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "Verbose, print out progress")

	var overwrite bool
	flag.BoolVar(&overwrite, "f", false, "Overwrite output-file, if exists")

	var output string
	flag.StringVar(&output, "o", "-", "Name of output file (default: '-', STDOUT)")

	var method string
	flag.StringVar(&method, "X", "GET", "HTTP method to use for request (default: GET)")

	var data string
	flag.StringVar(&data, "d", "", "Request content")

	var reqHeaders headerList
	flag.Var(&reqHeaders, "H", "HTTP method to use for request (default: GET)")

	flag.Parse()

	if len(flag.Args()) < 1 {
		flag.Usage()
		os.Exit(1)
	}
	url := flag.Arg(0)

	var requestBody *strings.Reader = nil
	if data != "" {
		requestBody = strings.NewReader(data)
	}
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "while creating http-request: %s", err)
		os.Exit(2)
	}
	for _, rawHeader := range reqHeaders {
		t := strings.SplitN(rawHeader, ":", 2)
		if len(t) != 2 {
			_, _ = fmt.Fprintf(os.Stderr, "invalid header: '%s'", rawHeader)
			os.Exit(2)
		}
		req.Header.Add(t[0], t[1])
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "while doing http-request: %s", err)
		os.Exit(2)
	}
	defer func() { _ = resp.Body.Close() }()

	conv := mockhttp.Conversation{}
	conv.Name = namesgenerator.GetRandomName(0)
	conv.Request = mockhttp.Request{
		UrlMatcher:     mockhttp.UrlMatcher{Path: req.URL.Path, Query: req.URL.Query().Encode()},
		MethodMatcher:  req.Method,
		HeaderMatchers: reqHeaders,
		BodyMatcher:    data,
	}

	var headers []string
	for k, hvs := range resp.Header {
		headers = append(headers, fmt.Sprintf("%s: %s", k, hvs[0]))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "while reading response: %s", err)
		os.Exit(2)
	}
	fmt.Printf("body: %s\n", body)

	conv.Response = mockhttp.Response{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       string(body),
		BodyFile:   "",
	}

	outputWriter := os.Stdout
	if output != "-" {
		file, err := os.Create(output)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to create file '%s': %s", output, err)
			os.Exit(2)
		}
		defer func() { _ = file.Close() }()
		outputWriter = file
	}
	err = yaml.NewEncoder(outputWriter).Encode(&conv)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "while encoding yaml: %s", err)
		os.Exit(2)
	}
}
