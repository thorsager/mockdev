package main

import "strings"

type headerList []string

func (l *headerList) Set(s string) error {
	*l = append(*l, s)
	return nil
}
func (l *headerList) String() string {
	return strings.Join(*l, ", ")
}
