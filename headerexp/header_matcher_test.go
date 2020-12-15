package headerexp

import (
	"github.com/thorsager/mockdev/keyvalueexp"
	"reflect"
	"testing"
)

func TestMustCompile(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want *HeaderExpr
	}{
		{"basic", args{"Content-Type:.*"}, &HeaderExpr{keyvalueexp.MustCompile(map[string]string{"Content-Type": ".*"})}},
		{"with-space", args{"Content-Type: ^.*$"}, &HeaderExpr{keyvalueexp.MustCompile(map[string]string{"Content-Type": "^.*$"})}},
		{"app", args{"Content-Type: ^application/.*$"}, &HeaderExpr{keyvalueexp.MustCompile(map[string]string{"Content-Type": "^application/.*$"})}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustCompile(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustCompile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func a(arg ...string) []string {
	return arg
}

func TestHeaderExpr_MatchString(t *testing.T) {
	type fields struct {
		expr []string
	}
	type args struct {
		headerStrings []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"all", fields{a("C-T:.*")}, args{a("C-T:bar")}, true},
		{"all_space", fields{a("C-T:.*")}, args{a("C-T: bar")}, true},
		{"none", fields{a("hdr:.*")}, args{a("")}, false},
		{"numbers", fields{a("foo:^\\d+$")}, args{a("Foo: 123")}, true},
		{"numbers_fail", fields{a("foo: ^\\d+$")}, args{a("Foo: abc")}, false},
		{"single-number", fields{a("foo:^\\d$")}, args{a("foo: 3")}, true},
		{"single-number_fail", fields{a("foo:^\\d$")}, args{a("Foo:13")}, false},
		{"multi", fields{a("f:^\\d$", "d:^\\w+$")}, args{a("F: 3", "d:X")}, true},
		{"multi_fail", fields{a("f:^\\d$", "d:^\\w+$")}, args{a("F: b", "d:X")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := MustCompile(tt.fields.expr...)
			if got := q.MatchString(tt.args.headerStrings...); got != tt.want {
				t.Errorf("MatchString() = %v, want %v", got, tt.want)
			}
		})
	}
}
