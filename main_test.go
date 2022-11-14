package main

import (
	"reflect"
	"testing"
)

func Test_parseServices(t *testing.T) {
	tests := []struct {
		name     string
		services string
		want     []service
	}{
		{
			name:     "empty",
			services: "",
			want:     nil,
		}, {
			name:     "single",
			services: `[{"Name":"one","Tags":["hi"]}]`,
			want: []service{
				{Name: "one", Tags: []string{"hi"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseServices(tt.services); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseServices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_do(t *testing.T) {
	type args struct {
		dtab     string
		services string
		tag      string
	}
	tests := []struct {
		name string
		args args
		want config
	}{{
		name: "empty",
		want: config{},
	}, {
		name: "ads only test",
		args: args{
			dtab:     `/http/1.1 => /#/io.l5d.consulcanary/dc1 ; /http/1.1/always => /#/io.l5d.consulcanary/dc1/canary | /#/io.l5d.consulcanary/dc1/noncanary ; /http/1.1/enabled => (0 * /#/io.l5d.consulcanary/dc1/canary & 100 * /#/io.l5d.consulcanary/dc1/noncanary) | /#/io.l5d.consulcanary/dc1/noncanary ; /http/1.1/disabled => /#/io.l5d.consulcanary/dc1/noncanary ; /http/1.1/enabled/ads => /#/io.l5d.consulcanary/dc1/noncanary/ads `,
			services: `[{"Name":"ads","Tags":["ads","https","noncanary","rolling","v13242211076df65z27797336"]},{"Name":"ads-admin","Tags":["ads","admin","http","noncanary","rolling"]},{"Name":"ads-health","Tags":["ads","health","noncanary"]}]`,
			tag:      `https`,
		},
		want: config{
			CanaryServices: []CanaryService{
				{
					Name:             "ads",
					NonCanaryWeight:  "100",
					CanaryWeight:  "0",
				},
			},
		},
	}, {
		name: "no canary in dtab, canary in tags",
		args: args{
			dtab:     ` /http/1.1/enabled/a-data-service => /#/io.l5d.consulcanary/dc1/noncanary/a-data-service `,
			services: `[{"Name":"a-data-service","Tags":["a-data-service","canary","https"]}]`,
		},
		want: config{
			CanaryServices: []CanaryService{{Name: "a-data-service", NonCanaryWeight: "100", CanaryWeight: "0"}},
		},
	}, {
		name: "canary in dtab",
		args: args{
			dtab:     `/http/1.1/enabled/hello-world-a => (100 * /#/io.l5d.consulcanary/achm2/canary/hello-world-a & 0 * /#/io.l5d.consulcanary/achm2/noncanary/hello-world-a) | /#/io.l5d.consulcanary/achm2/noncanary/hello-world-a`,
			services: `[{"Name":"hello-world-a","Tags":["canary","https","rolling"]}]`,
		},
		want: config{
			CanaryServices: []CanaryService{{
				Name:             "hello-world-a",
				NonCanaryWeight:  "0",
				CanaryWeight:     "100",
			}},
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := do(tt.args.dtab, tt.args.services, tt.args.tag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("do() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanaryRE_Matches(t *testing.T) {
	one := `(100 * /#/io.l5d.consulcanary/achm2/canary/hello-world-a & 0 * /#/io.l5d.consulcanary/achm2/noncanary/hello-world-a) | /#/io.l5d.consulcanary/achm2/noncanary/hello-world-a`
	if !canaryRE.MatchString(one) {
		t.Errorf("canaryRE does not match expected match %v", one)
	}
	m := canaryRE.FindStringSubmatch(one)
	if m[1] != "100" {
		t.Errorf("canaryRE does not match expected 100 from %v matches: %v", one, m)
	}
	if m[2] != "0" {
		t.Errorf("canaryRE does not match expected 100 from %v matches: %v", one, m)
	}

}
