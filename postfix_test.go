package main

import (
	"regexp"
	"testing"
)

type tPostfixLogLine struct {
	line      string
	prefixLen int
}

func TestPostfixLogLine(t *testing.T) {
	lines := []tPostfixLogLine{
		{"Jan  1 01:05:02 mailserver postfix/local[17781]:", 35},
		{"Feb 10 01:05:02 mailserver postfix/local[17781]:", 35},
		{"Mar 21 01:05:02 mailserver postfix/local[17781]:", 35},
		{"Apr 31 01:05:02 mailserver postfix/local[17781]:", 35},
		{"May 12 00:00:00 mailserver postfix/local[17781]:", 35},
		{"Jun 13 01:11:11 mailserver postfix/local[17781]:", 35},
		{"Jul 14 02:22:22 mailserver postfix/local[17781]:", 35},
		{"Aug 15 01:33:33 mailserver postfix/local[17781]:", 35},
		{"Sep 16 01:44:44 mailserver postfix/local[17781]:", 35},
		{"Oct 17 01:55:55 mailserver postfix/local[17781]:", 35},
		{"Nov 18 01:05:06 mailserver postfix/local[17781]:", 35},
		{"Dec 19 01:05:07 mailserver postfix-instalnce_name/local[17781]:", 50},
	}

	re, err := regexp.Compile(postfixLogLine)
	if err != nil {
		t.Error("rePostfixLogLine regexp compile error:", err)
	} else {
		for _, l := range lines {
			lMatch := re.FindStringSubmatch(l.line)
			if lMatch == nil {
				t.Errorf("postfix log line %q does not match", l)
			}
			prefixLen := len(lMatch[0])
			if prefixLen != l.prefixLen {
				t.Errorf("found of incorrect log line prefix length (%d) in %q", prefixLen, l.line)
			}
		}
	}
}

func TestForwardedLine(t *testing.T) {
	aLine := "local[17781]: 9093C182F98: to=<x@example.com>, relay=local, delay=0.04, delays=0.03/0.01/0/0, dsn=2.0.0, status=sent (forwarded as 96643182F99)"

	re, err := regexp.Compile(forwardedLine)
	if err != nil {
		t.Error("reForwardedLine regexp compile error:", err)
	} else if !re.MatchString(aLine) {
		t.Error("forwarded line does not match")
	}
}
