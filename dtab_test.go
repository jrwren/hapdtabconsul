package main

import (
	"reflect"
	"testing"
)

func Test_parseDTab(t *testing.T) {
	tests := []struct {
		name    string
		dtabStr string
		want    dtab
	}{
		{
			name:    "empty",
			dtabStr: "",
			want:    nil,
		}, {
			name:    "single",
			dtabStr: "/s => /s#/env",
			want:    dtab{entry{src: "/s", dst: "/s#/env"}},
		},
		{
			name: "example",
			dtabStr: `/zk##    => /$/com.twitter.serverset;
			/zk#     => /zk##/127.0.0.1:2181;
			/s#      => /zk#/service;
			/env     => /s#/local;
			/s       => /env;`,
			want: dtab{
				entry{src: "/s", dst: "/env"},
				entry{src: "/env", dst: "/s#/local"},
				entry{src: "/s#", dst: "/zk#/service"},
				entry{src: "/zk#", dst: "/zk##/127.0.0.1:2181"},
				entry{src: "/zk##", dst: "/$/com.twitter.serverset"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDTab(tt.dtabStr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDTab() = %v, want %v", got, tt.want)
			}
		})
	}
}
