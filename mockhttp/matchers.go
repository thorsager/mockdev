package mockhttp

import (
	"bytes"
	"context"
	"github.com/thorsager/mockdev/headerexp"
	"github.com/thorsager/mockdev/queryexp"
	"io/ioutil"
	"net/http"
	"regexp"
)

func matchURL(ctx context.Context, r *http.Request, c Conversation) bool {
	score, _ := getConversationScores(ctx)
	if c.Request.UrlMatcher == (UrlMatcher{}) {
		return true // no matcher, (that is a win)
	}

	pathMatch := true
	if c.Request.UrlMatcher.Path != "" {
		urlPath := regexp.MustCompile(c.Request.UrlMatcher.Path)
		if pathMatch = urlPath.MatchString(r.URL.Path); pathMatch {
			score.inc(c.Name)
		}
	}

	queryMatch := true
	if c.Request.UrlMatcher.Query != "" {
		urlQuery := queryexp.MustCompile(c.Request.UrlMatcher.Query)
		if c.Request.UrlMatcher.QueryLooseMatch {
			queryMatch = urlQuery.ContainedInQuery(r.URL.Query())
		} else {
			queryMatch = urlQuery.MatchQuery(r.URL.Query())
		}
		if queryMatch {
			score.bump(c.Name, urlQuery.MatcherCount())
		}
	}

	return pathMatch && queryMatch
}

func matchBody(ctx context.Context, r *http.Request, c Conversation) bool {
	score, _ := getConversationScores(ctx)
	log, _ := getContextLogger(ctx)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("%v", err)
		return false
	}
	_ = r.Body.Close() //  must close
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if c.Request.BodyMatcher == "" {
		return true // no matcher, that is a win
	}
	method := regexp.MustCompile(c.Request.BodyMatcher)
	if method.Match(body) {
		score.inc(c.Name)
		return true
	}
	return false
}

func matchMethod(ctx context.Context, r *http.Request, c Conversation) bool {
	score, _ := getConversationScores(ctx)
	if c.Request.MethodMatcher == "" {
		return true // no matcher, that is a win
	}
	method := regexp.MustCompile(c.Request.MethodMatcher)
	if method.MatchString(r.Method) {
		score.inc(c.Name)
		return true
	}
	return false
}

func matchHeaders(ctx context.Context, r *http.Request, c Conversation) bool {
	score, _ := getConversationScores(ctx)
	if len(c.Request.HeaderMatchers) == 0 {
		return true // mo matchers, that is a win
	}
	headers := headerexp.MustCompile(c.Request.HeaderMatchers...)
	var doesMatch bool
	if c.Request.GetHeaderMatchType() == Contains {
		doesMatch = headers.ContainedInHeader(r.Header)
	} else {
		doesMatch = headers.MatchHeader(r.Header)
	}
	if doesMatch {
		score.inc(c.Name)
		return true
	}
	return false
}
