package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_respBodyWrapper(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "should return PONG for PING command",
			args: args{args: []string{"PING"}},
			want: []byte("*1\r\n$4\r\nPING\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := respBodyWrapper(tt.args.args); !assert.Equal(t, tt.want, got) {
				t.Errorf("respBodyWrapper() = %v, want %v", got, tt.want)
			}

		})
	}
}
