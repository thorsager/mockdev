package mockhttp

import (
	"bytes"
	"fmt"
	"github.com/thorsager/mockdev/headerexp"
	"github.com/thorsager/mockdev/logging"
	"github.com/thorsager/mockdev/queryexp"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
)

type ConversationsHandler struct {
	sync.Mutex
	Conversations      []Conversation
	Log                logging.Logger
	SessionLogLocation string
	SessionLogReceived bool
	sessionCounter     int
}

func (h *ConversationsHandler) nextSession() int {
	h.Lock()
	defer h.Unlock()
	h.sessionCounter = h.sessionCounter + 1
	return h.sessionCounter
}

func (h *ConversationsHandler) sessionFilename(sesId int) string {
	return path.Join(h.SessionLogLocation, fmt.Sprintf("sess_%.4d.log", sesId))
}

func (h *ConversationsHandler) initSessionLog(sesId int) error {
	if h.SessionLogReceived {
		filename := h.sessionFilename(sesId)
		h.Log.Debugf("logging session to: %s", filename)
		err := os.MkdirAll(h.SessionLogLocation, 0770)
		if err != nil {
			return err
		}
		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = f.WriteString(fmt.Sprintf("# session %d, %s \n", sesId, time.Now()))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *ConversationsHandler) logRequestBody(sesId int, reader io.Reader) error {
	if !h.SessionLogReceived {
		return nil
	}
	f, err := os.OpenFile(h.sessionFilename(sesId), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = io.Copy(f, reader)
	if err != nil {
		return err
	}
	return nil
}

func (h *ConversationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sid := h.nextSession()
	err := h.initSessionLog(sid)
	if err != nil {
		h.Log.Errorf("sessionLogInit: %v", err)
		http.Error(w, "sessionLogInit: "+err.Error(), 418)
		return
	}

	// make body-persistent
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	h.Log.Tracef("Request %s %s", r.Method, r.URL)
	h.Log.Tracef("%s", bodyBytes)

	candidates := h.filterByMethod(r, h.Conversations)
	if len(candidates) < 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("Found no Method Matches")
		return
	}
	candidates = h.filterByUrl(r, candidates)
	if len(candidates) < 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("No matches after url-filter")
		return
	}
	candidates = h.filterByHeader(r, candidates)
	if len(candidates) < 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("No matches after header-filter")
		return
	}
	candidates = h.filterByBody(r, candidates)
	if len(candidates) > 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("Multiple matches after body-filter")
		return
	}
	if len(candidates) != 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("No matches after body-filter")
		return
	}
	_ = h.logRequestBody(sid, bytes.NewBuffer(bodyBytes))
	_ = h.serveResponse(w, r, candidates[0])
}

func (h *ConversationsHandler) serveResponse(w http.ResponseWriter, r *http.Request, conversation Conversation) error {
	for _, s := range conversation.Response.Headers {
		addHeaderFromString(w, s)
	}
	bodyBuffer := &bytes.Buffer{}
	if conversation.Response.BodyFile != "" {
		f, err := os.Open(conversation.Response.BodyFile)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = io.Copy(bodyBuffer, f)
		if err != nil {
			return err
		}
	} else {
		bodyBuffer.Write([]byte(conversation.Response.Body))
	}

	groups := make(map[string]string)

	if conversation.Request.UrlMatcher.Path != "" {
		m := regexp.MustCompile(conversation.Request.UrlMatcher.Path)
		matches := m.FindStringSubmatch(r.URL.Path)
		for i, j := range matches {
			h.Log.Debugf("p%d=%s", i, j)
			groups[fmt.Sprintf("p%d", i)] = j
		}
	}

	if conversation.Request.UrlMatcher.Query != "" {
		m := regexp.MustCompile(conversation.Request.UrlMatcher.Query)
		matches := m.FindStringSubmatch(r.URL.Path)
		for i, j := range matches {
			groups[fmt.Sprintf("q%d", i)] = j
		}
	}

	if conversation.Request.BodyMatcher != "" {
		m := regexp.MustCompile(conversation.Request.BodyMatcher)
		matches := m.FindStringSubmatch(r.URL.Path)
		for i, j := range matches {
			groups[fmt.Sprintf("b%d", i)] = j
		}
	}

	// TODO: Figure out how match-groups could be implemented on Header Matchers.

	tmpl, err := template.New("body").Parse(bodyBuffer.String())
	if err != nil {
		return err
	}

	executedBuffer := &bytes.Buffer{}
	err = tmpl.Execute(executedBuffer, groups)
	if err != nil {
		return err
	}

	w.Header().Add("X-Powered-By", "mockdev")
	w.Header().Add("size", fmt.Sprintf("%d", executedBuffer.Len()))
	w.WriteHeader(conversation.Response.StatusCode)
	_, _ = w.Write(executedBuffer.Bytes())

	h.Log.Infof("Served response from conversation: '%s' (%s)", conversation.Name, r.URL)
	return nil
}

func addHeaderFromString(w http.ResponseWriter, s string) {
	t := strings.SplitN(s, ":", 2)
	w.Header().Add(t[0], t[1])
}

func (h *ConversationsHandler) filterByHeader(r *http.Request, candidates []Conversation) []Conversation {
	var filtered []Conversation
	for _, c := range candidates {
		if len(c.Request.HeaderMatchers) == 0 {
			filtered = append(filtered, c)
			continue
		}
		headers := headerexp.MustCompile(c.Request.HeaderMatchers...)
		if headers.MatchHeader(r.Header) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (h *ConversationsHandler) filterByMethod(r *http.Request, candidates []Conversation) []Conversation {
	var filtered []Conversation
	for _, c := range candidates {
		if c.Request.MethodMatcher == "" {
			filtered = append(filtered, c)
			continue
		}
		method := regexp.MustCompile(c.Request.MethodMatcher)
		if method.MatchString(r.Method) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (h *ConversationsHandler) filterByBody(r *http.Request, candidates []Conversation) []Conversation {
	var filtered []Conversation
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return filtered
	}
	for _, c := range candidates {
		if c.Request.BodyMatcher == "" {
			filtered = append(filtered, c)
			continue
		}
		method := regexp.MustCompile(c.Request.BodyMatcher)
		if method.Match(body) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

func (h *ConversationsHandler) filterByUrl(r *http.Request, candidates []Conversation) []Conversation {
	var filtered []Conversation
	for _, c := range candidates {
		if c.Request.UrlMatcher == (UrlMatcher{}) {
			filtered = append(filtered, c)
			continue
		}

		pathMatch := true
		if c.Request.UrlMatcher.Path != "" {
			urlPath := regexp.MustCompile(c.Request.UrlMatcher.Path)
			pathMatch = urlPath.MatchString(r.URL.Path)
		}

		queryMatch := true
		if c.Request.UrlMatcher.Query != "" {
			urlQuery := queryexp.MustCompile(c.Request.UrlMatcher.Query)
			queryMatch = urlQuery.MatchQuery(r.URL.Query())
		}

		if pathMatch && queryMatch {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
