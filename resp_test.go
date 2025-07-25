package main

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestRespReader_parseCommand(t *testing.T) {
	type cas struct {
		name      string
		resp      *RespReader
		exceptCmd []string
	}

	cases := []cas{
		{
			name:      "should parse PING command",
			resp:      NewRespReader(strings.NewReader("*1\r\n$4\r\nPING\r\n")),
			exceptCmd: []string{"PING"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			args, err := c.resp.Args()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assert.Equal(t, c.exceptCmd, args)
		})
	}
}
