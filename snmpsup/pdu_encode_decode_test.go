package snmpsup

import (
	"reflect"
	"testing"
)

func TestParseNeutralPDU(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *NeutralPDU
		wantErr bool
	}{
		{"hex", args{s: ".1.3.6.1.2.1.55.1.5.1.8.9/4/hex-string/802aa85d1dd5"}, &NeutralPDU{Oid: ".1.3.6.1.2.1.55.1.5.1.8.9", Asn1BER: 0x4, Value: "\x80*\xa8]\x1d\xd5"}, false},
		{"ascii", args{s: ".1.3.6.1.2.1.55.1.5.1.8.9/4/string/hello world"}, &NeutralPDU{Oid: ".1.3.6.1.2.1.55.1.5.1.8.9", Asn1BER: 0x4, Value: "hello world"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNeutralPDU(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNeutralPDU() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNeutralPDU() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestNeutralPDU_String(t *testing.T) {
	type fields struct {
		Oid     string
		Asn1BER byte
		Value   interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"hex", fields{Oid: ".1.3.6.1.2.1.55.1.5.1.8.9", Asn1BER: 4, Value: "\x80*\xa8]\x1d\xd5"}, ".1.3.6.1.2.1.55.1.5.1.8.9/4/hex-string/802aa85d1dd5"},
		{"ascii", fields{Oid: ".1.3.6.1.2.1.55.1.5.1.8.9", Asn1BER: 4, Value: "hello world"}, ".1.3.6.1.2.1.55.1.5.1.8.9/4/string/hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &NeutralPDU{
				Oid:     tt.fields.Oid,
				Asn1BER: tt.fields.Asn1BER,
				Value:   tt.fields.Value,
			}
			if got := p.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
