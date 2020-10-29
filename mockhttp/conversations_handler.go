package mockhttp

import (
	"fmt"
	"github.com/thorsager/mockdev/headerexp"
	"github.com/thorsager/mockdev/logging"
	"github.com/thorsager/mockdev/queryexp"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type ConversationsHandler struct {
	Conversations []Conversation
	Log           logging.Logger
}

func (h ConversationsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	candidates := filterByMethod(r, h.Conversations)
	if len(candidates) < 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("Found no Method Matches")
		return
	}
	candidates = filterByUrl(r, candidates)
	if len(candidates) < 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("No matches after url-filter")
		return
	}
	candidates = filterByHeader(r, candidates)
	if len(candidates) < 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("No matches after header-filter")
		return
	}
	candidates = filterByBody(r, candidates)
	if len(candidates) != 1 {
		http.Error(w, "I'm not a teapot", 418)
		h.Log.Warnf("No matches after body-filter")
		return
	}
	_ = h.serveResponse(w, r, candidates[0])
}

func (h *ConversationsHandler) serveResponse(w http.ResponseWriter, r *http.Request, conversation Conversation) error {
	for _, s := range conversation.Response.Headers {
		addHeaderFromString(w, s)
	}
	if conversation.Response.BodyFile != "" {
		fi, err := os.Stat(conversation.Response.BodyFile)
		if err != nil {
			return err
		}
		w.Header().Add("X-JOE", "dalton")
		w.Header().Add("size", fmt.Sprintf("%d", fi.Size()))
		w.WriteHeader(conversation.Response.StatusCode)
		f, err := os.Open(conversation.Response.BodyFile)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		_, err = io.Copy(w, f)
		if err != nil {
			return err
		}
	} else {
		w.WriteHeader(conversation.Response.StatusCode)
		_, _ = w.Write([]byte(conversation.Response.Body))
	}
	h.Log.Infof("Served response from conversation: '%s' (%s)", conversation.Name, r.URL)
	return nil
}

func addHeaderFromString(w http.ResponseWriter, s string) {
	t := strings.SplitN(s, ":", 2)
	w.Header().Add(t[0], t[1])
}

func filterByHeader(r *http.Request, candidates []Conversation) []Conversation {
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

func filterByMethod(r *http.Request, candidates []Conversation) []Conversation {
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

func filterByBody(r *http.Request, candidates []Conversation) []Conversation {
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

func filterByUrl(r *http.Request, candidates []Conversation) []Conversation {
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
