package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

var debug bool

func main() {
	var tag string
	flag.BoolVar(&debug, "debug", false, "enable debug output")
	flag.StringVar(&tag, "tag", "https", "tag of interest")
	// TODO: consider flag.BoolVar(&text, "text, false, "output text")
	// to output text haproxy config instead of JSON.
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Fprint(os.Stderr, "expected 2 args dtab string and services JSON")
		os.Exit(1)
	}
	c := do(flag.Arg(0), flag.Arg(1), tag)
	json.NewEncoder(os.Stdout).Encode(c)
}

func do(dtab, services, tag string) config {
	if dtab == "" || services == "" {
		return config{}
	}
	pd := parseDTab(dtab)
	h1e := filterHTTP1_1Enabled(pd)
	ps := parseServices(services)
	if debug {
		log.Printf("filter dtab: %v",h1e)
		log.Printf("parsed services: %v",ps)
	}
	return buildConfig(h1e, ps, tag)
}

// config represents haproxy config for frontend and backend
type config struct {
	CanaryServices []CanaryService `json:"canary_services,omitempty"`
}

// CanaryService has weights for canaries.
type CanaryService struct {
	Name            string `json:"name,omitempty"`
	NonCanaryWeight string `json:"non_canary_weight,omitempty"`
	CanaryWeight    string `json:"canary_weight,omitempty"`
}

func buildConfig(dtab map[string]string, services servicesByLen, tag string) (c config) {
	sort.Sort(services)
	for i := range services {
		tagFound := true
		if tag != "" {
			tagFound = false
			for j := range services[i].Tags {
				if services[i].Tags[j] == tag {
					tagFound = true
				}
			}
		}
		if tagFound {
			dst := dtab[services[i].Name]
			// e.g. /#/io.l5d.consulcanary/dc1/noncanary/frontend-service
			if strings.HasPrefix(dst, `/#/io.l5d.consulcanary/`) {
				after := dst[24:]
				parts := strings.Split(after, "/")
				// todo: consider log on inconsistent dc.
				// dc := parts[0]
				if parts[1] != "noncanary" {
					log.Print("unhandled dtab entry format: ", dst, "##", parts)
					continue
				}
				f := CanaryService{Name: services[i].Name,
					NonCanaryWeight:  "100"}
				c.CanaryServices = append(c.CanaryServices, f)
			} else if m := canaryRE.FindStringSubmatch(dst); len(m) > 0 {

				f := CanaryService{Name: services[i].Name,
					NonCanaryWeight:  m[2],
					CanaryWeight:     m[1],
				}
				c.CanaryServices = append(c.CanaryServices, f)
			} else {
				if dst != "" {
					m := canaryRE.FindStringSubmatch(dst)
					log.Printf("unhandled dtab entry format: %s match: %v", dst, m)
				}
			}
		}
	}
	return c
}

// canaryRE matches e.g. (10 * /#/io.l5d.consulcanary/aint1/canary/serv & 90 * /#/io.l5d.consulcanary/aint1/noncanary/serv) | /#/io.l5d.consulcanary/aint1/noncanary/serv
// (100 * /#/io.l5d.consulcanary/achm2/canary/hello-world-a & 0 * /#/io.l5d.consulcanary/achm2/noncanary/hello-world-a) | /#/io.l5d.consulcanary/achm2/noncanary/hello-world-a
var canaryRE = regexp.MustCompile(`\((\d+) \* /#/io.l5d.consulcanary/[a-z0-9-]+/canary/[a-z0-9-]+ & (\d+) \* /#/io.l5d.consulcanary/[a-z0-9]+/noncanary/[a-z0-9-]+\) | /#/io.l5d.consulcanary/[a-z0-9]+/noncanary/[a-z0-9-]+`)

type service struct {
	Name string   `json:"Name"`
	Tags []string `json:"Tags"`
}

func parseServices(services string) (s []service) {
	err := json.Unmarshal([]byte(services), &s)
	if err != nil {
		log.Printf("error parsing services: %v", err)
	}

	return s
}

type servicesByLen []service

func (a servicesByLen) Len() int      { return len(a) }
func (a servicesByLen) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less is actually greater because we want to sort largest to smallest by length of Name.
func (a servicesByLen) Less(i, j int) bool { return len(a[i].Name) > len(a[j].Name) }
