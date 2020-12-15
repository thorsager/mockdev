package queryexp

import (
	"fmt"
	"github.com/thorsager/mockdev/keyvalueexp"
	"net/url"
	"strings"
)

type QueryExpr struct {
	*keyvalueexp.KeyValueExpr
}

// MatchString, will match against the query part of an URL, return true
// if all parameters are matched, if not false is returned
func (q *QueryExpr) MatchString(s string) bool {
	u := url.URL{RawQuery: s}
	return q.MatchUrl(u)
}

// MatchUrl will match the URL.Query() against the QueryExpr, and return
// true if all parameters are matched, if not false is returned
func (q *QueryExpr) MatchUrl(u url.URL) bool {
	return q.MatchQuery(u.Query())
}

// MatchQuery will match the URL.Values against the QueryExpr, and return
// true if all parameters are matched, if not false is returned.
func (q *QueryExpr) MatchQuery(v url.Values) bool {
	m := make(map[string]string)
	for k, v := range v {
		m[k] = v[0]
	}
	return q.MatchMap(m)
}

// Compile, will create a QueryExpr from a string in URL query format, but with
// the twist that all parameter will be treated as a RegularExpression.
// ex.
// 'foo=^[a-z]{2}$&bar=^.^$'
func Compile(s string) (*QueryExpr, error) {
	rawMatchers, err := pseudoQueryToMap(s)
	if err != nil {
		return nil, err
	}
	kve, err := keyvalueexp.Compile(rawMatchers)
	if err != nil {
		return nil, err
	}
	return &QueryExpr{kve}, nil
}

// MustCompile, does the same thing as Compile, only it will panic if an error
// occurs during compilation of expression.
func MustCompile(s string) *QueryExpr {
	q, err := Compile(s)
	if err != nil {
		panic(err)
	}
	return q
}

func pseudoQueryToMap(s string) (map[string]string, error) {
	m := make(map[string]string)
	kvs := strings.Split(s, "&")
	for _, kv := range kvs {
		tuple := strings.Split(kv, "=")
		if len(tuple) != 2 {
			return nil, fmt.Errorf("unable to parse '%s' (invlaid '%s')", s, kv)
		}
		m[tuple[0]] = tuple[1]
	}
	return m, nil
}
