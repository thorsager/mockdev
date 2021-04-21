package headerexp

import (
	"fmt"
	"github.com/thorsager/mockdev/keyvalueexp"
	"net/http"
	"net/textproto"
	"strings"
)

type HeaderExpr struct {
	*keyvalueexp.KeyValueExpr
}

// MatchString, will match against the query part of an URL, return true
// if all parameters are matched, if not false is returned
func (q *HeaderExpr) MatchString(headerStrings ...string) bool {
	headers := http.Header{}
	for _, s := range headerStrings {
		tuple := strings.SplitN(s, ":", 2)
		if len(tuple) != 2 {
			return false
		}
		headers.Add(strings.TrimSpace(tuple[0]), strings.TrimSpace(tuple[1]))
	}
	return q.MatchHeader(headers)
}

// MatchHeader will match the http.Header structure against the HeaderExpr,
// and return true all headers present in the request matches a HeaderExpr if one is
// found for the specific header. If no HeaderExpr is found for a passed header it
// is ignored thus will *not* fail the match.
func (q *HeaderExpr) MatchHeader(h http.Header) bool {
	m := make(map[string]string)
	for k, v := range h {
		m[k] = v[0]
	}
	return q.MatchIfPresentMap(m)
}

// Compile, will create a HeaderExpr from a string in URL query format, but with
// the twist that all parameter will be treated as a RegularExpression.
// ex.
// m,err := Compile(["Content-Type: ^application/.*$","Accept: ^text/.*$"])
// please note that all header names are passed through textproto.CanonicalMIMEHeaderKey
// to account for header-name transformations.
func Compile(headerStrings ...string) (*HeaderExpr, error) {
	rawMatchers := make(map[string]string)
	for _, s := range headerStrings {
		m, err := headerStringToMap(s)
		if err != nil {
			return nil, err
		}
		rawMatchers, err = stringMapJoin(rawMatchers, m, false)
		if err != nil {
			return nil, err
		}
	}
	kve, err := keyvalueexp.Compile(rawMatchers)
	if err != nil {
		return nil, err
	}
	return &HeaderExpr{kve}, nil
}

// MustCompile, does the same thing as Compile, only it will panic if an error
// occurs during compilation of expression.
func MustCompile(headerStrings ...string) *HeaderExpr {
	q, err := Compile(headerStrings...)
	if err != nil {
		panic(err)
	}
	return q
}

func headerStringToMap(s string) (map[string]string, error) {
	m := make(map[string]string)
	tuple := strings.SplitN(s, ":", 2)
	if len(tuple) != 2 {
		return nil, fmt.Errorf("unable to parse '%s'", s)
	}
	m[textproto.CanonicalMIMEHeaderKey(strings.TrimSpace(tuple[0]))] = strings.TrimSpace(tuple[1])
	return m, nil
}

func stringMapJoin(m1, m2 map[string]string, overwrite bool) (map[string]string, error) {
	joined := make(map[string]string)
	for k, v := range m1 {
		joined[k] = v
	}
	for k, v := range m2 {
		if overwrite {
			v1, ok := joined[k]
			if ok {
				return nil, fmt.Errorf("key collision: '%s=>%s' vs. '%s=%s", k, v, k, v1)
			}
		}
		joined[k] = v
	}
	return joined, nil
}
