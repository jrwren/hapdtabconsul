package main

import (
	"log"
	"strings"
)

type dtab []entry

type entry struct {
	src, dst string
}

// parseDTab parses the dtab string into a dtab.
// https://gutefrage.github.io/the-finagle-docs/discovery/resolving-dtabs.html
func parseDTab(dtabStr string) dtab {
	var dtab dtab
	currentDtab := strings.ReplaceAll(dtabStr, "\n", "")
	entries := strings.Split(currentDtab, ";")
	// for i := reverse range entries {
	for i := len(entries) - 1; i >= 0; i-- {
		if strings.TrimSpace(entries[i]) == "" {
			continue
		}
		l, r, ok := strings.Cut(entries[i], " => ")
		if !ok {
			log.Print("error when reading dtab entry", entries[i])
		}
		l = strings.TrimSpace(l)
		r = strings.TrimSpace(r)
		dtab = append(dtab, entry{l, r})
	}
	return dtab
}

// filterHTTP1_1Enabled map only /http/1.1/enabled/ src entries
func filterHTTP1_1Enabled(dtab dtab) map[string]string {
	m := make(map[string]string)
	for i := range dtab {
		if strings.HasPrefix(dtab[i].src, `/http/1.1/enabled/`) {
			host := dtab[i].src[18:]
			if strings.Contains(host, "/") {
				log.Print("unexpected extra component after host in ",
					dtab[i].src)
			}
			m[host] = dtab[i].dst
			// Log if the last token does not match.
			j := strings.LastIndex(dtab[i].dst, "/")
			if j >= len(dtab[i].dst) {
				log.Print("unexpected dtab entry: ", dtab[i].src, " -> ", dtab[i].dst)
			}
			if host != dtab[i].dst[j+1:] {
				log.Print("unexpected dtab entry: ", dtab[i].src, " -> ", dtab[i].dst, "##", host, "!=", dtab[i].dst[j:])
			}
		}
	}
	return m
}
