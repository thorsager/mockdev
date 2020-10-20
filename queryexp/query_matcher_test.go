package queryexp

import (
	queryexp "github.com/thorsager/mockdev/headerexp"
	"github.com/thorsager/mockdev/keyvalueexp"
	"reflect"
	//"regexp"

	//"regexp"
	"testing"
)

func TestMustCompile(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want *QueryExpr
	}{
		{"basic", args{"foo=.*&bar=.*"}, &QueryExpr{keyvalueexp.MustCompile(map[string]string{"foo": ".*", "bar": ".*"})}},
		{"number", args{"foo=\\d+&bar=.*"}, &QueryExpr{keyvalueexp.MustCompile(map[string]string{"foo": "\\d+", "bar": ".*"})}},
		{"start-end", args{"foo=^\\d+$&bar=.*"}, &QueryExpr{keyvalueexp.MustCompile(map[string]string{"foo": "^\\d+$", "bar": ".*"})}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustCompile(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustCompile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryExpr_MatchString(t *testing.T) {
	type fields struct {
		expr string
	}
	type args struct {
		s string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"all", fields{"foo=.*"}, args{"foo=bar"}, true},
		{"none", fields{"foo=.*"}, args{""}, false},
		{"numbers", fields{"foo=^\\d+$"}, args{"foo=123"}, true},
		{"numbers_fail", fields{"foo=^\\d+$"}, args{"foo=abc"}, false},
		{"single-number", fields{"foo=^\\d$"}, args{"foo=3"}, true},
		{"single-number_fail", fields{"foo=^\\d$"}, args{"foo=13"}, false},
		{"multi", fields{"f=^\\d$&b=^\\w$"}, args{"f=1&b=d"}, true},
		{"multi_fail", fields{"f=^\\d$&b=^\\w$"}, args{"f=d&b=1"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := queryexp.MustCompile(tt.fields.expr)
			if got := q.MatchString(tt.args.s); got != tt.want {
				t.Errorf("MatchString() = %v, want %v", got, tt.want)
			}
		})
	}
}
