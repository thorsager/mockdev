package keyvalueexp

import (
	"fmt"
	"regexp"
)

type matcherMap map[string]*regexp.Regexp

// KeyValueExpr, is a type that can be used to match a map[string]string against
// a map[string]*regexp.Regexp.
// Also the KeyValueExpr, can be created form a map[string]string containing a key
// and the string representation of a regexp. for this the Compile, or the MustCompile
// methods can be used in the same way as regexp.Compile or regexp.MustCompile
type KeyValueExpr struct {
	paramMatchers matcherMap
}

// MatchMap will compare all values, against the matching regexp.Regexp in the map
// regexp.Regexp and value, are "mapped" on the map key value, all keys in the passed value
// map are checked. If a key from the passed map, does not have a corresponding key
// and regexp.Regexp the will return false, also if a regexp.Regexp is found but
// does not match MatchMap will return false. Only if all keys are found and all
// values match will it return true.
func (ke *KeyValueExpr) MatchMap(m map[string]string) bool {
	if len(m) == 0 && len(ke.paramMatchers) != 0 {
		return false // no params but we have matchers
	}
	for k, v := range m {
		matcher, ok := ke.paramMatchers[k]
		if !ok {
			return false // no matcher found == unknown param
		}
		if !matcher.MatchString(v) {
			return false // value does not match
		}
	}
	return true
}

func (ke *KeyValueExpr) MatchIfPresentMap(m map[string]string) bool {
	if len(m) == 0 && len(ke.paramMatchers) != 0 {
		return false // no params but we have matchers
	}
	for k, matcher := range ke.paramMatchers {
		v, ok := m[k]
		if !ok {
			continue
		}
		if !matcher.MatchString(v) {
			return false // value does not match
		}
	}
	return true
}

// MustCompile this performs the same function as Compile, but it will panic if
// unable to successfully Compile.
func MustCompile(keyValue map[string]string) *KeyValueExpr {
	e, err := Compile(keyValue)
	if err != nil {
		panic(err)
	}
	return e
}

// Compile, compiles a map[string]string into a *KeyValueExpr, this is done by
// using regexp.Compile on all values in the passed map. If unable to do regexp.Compile
// on any of the map values, an error is returned, and the pointer returned will be nil
// ex.
// kve,err := Compile(map[string]string{"match_any":".*", "start_with_a":"^a.*"})
func Compile(keyValue map[string]string) (*KeyValueExpr, error) {
	matchers := make(matcherMap)
	for k, v := range keyValue {
		rxp, err := regexp.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("[%s]=%s: %v", k, v, err)
		}
		matchers[k] = rxp
	}
	return &KeyValueExpr{matchers}, nil
}
