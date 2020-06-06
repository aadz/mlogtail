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
	logLines := []tPostfixLogLine{
		{"Jan  1 01:05:02 mailserver postfix/local[17781]:", 35},
		{"Feb 10 01:05:02 mailserver postfix/local[17781]:", 35},
		{"Mar 21 01:05:02 mail-server postfix/local[17781]:", 36},
		{"Apr 31 01:05:02 mail_server postfix/local[17781]:", 36},
		{"May 12 00:00:00 mailserver postfix/local[17781]:", 35},
		{"Jun 13 01:11:11 -mailserver- postfix/local[17781]:", 37},
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
		for _, l := range logLines {
			lMatch := re.FindStringSubmatch(l.line)
			if lMatch == nil {
				t.Errorf("postfixLogLine does not match to %q", l)
			}
			prefixLen := len(lMatch[0])
			if prefixLen != l.prefixLen {
				t.Errorf("incorrect log line prefix length in %q, wanted %d got %d", l.line, l.prefixLen, prefixLen)
			}
		}
	}
}

func TestReceivedLine(t *testing.T) {
	logLines := []string{
		"smtpd[15500]: AD59432D65: client=mail1.example.com[123.123.123.123]",
		"smtps[15500]: AD59432D65: client=mail1.example.com[123.123.123.123]",
		"submission/smtpd[14518]: AD59432D65: client=mail.client.int[172.30.0.1], sasl_method=PLAIN, sasl_username=mail@iclient.int",
		"submission/smtps[14518]: AD59432D65: client=mail.client.int[172.30.0.1], sasl_method=PLAIN, sasl_username=mail@iclient.int",
		"pickup[13652]: AD59432D65: uid=65534 from=<nobody>",
	}
	queueIdWanted := "AD59432D65"

	re, err := regexp.Compile(receivedLine)
	if err != nil {
		t.Error("receivedLine regexp compile error:", err)
	} else {
		for _, l := range logLines {
			if sMatch := re.FindStringSubmatch(l); sMatch == nil {
				t.Errorf("receivedLine does not match to %q", l)
			} else if len(sMatch) != 2 || sMatch[1] != queueIdWanted {
				t.Errorf("receivedLine - cannot find queue ID in %q", l)
			}
		}
	}
}

func TestQueueActiveLine(t *testing.T) {
	aLine := "qmgr[29052]: 0A2D132D5F: from=<x@example.com>, size=1000, nrcpt=1 (queue active)"
	queueIdWanted := "0A2D132D5F"
	msgSizeWanted := "1000"

	re, err := regexp.Compile(queueActiveLine)
	if err != nil {
		t.Error("reQueueActiveLine regexp compile error:", err)
	} else if sMatch := re.FindStringSubmatch(aLine); sMatch == nil {
		t.Errorf("queueActiveLine does not match to %q", aLine)
	} else if !re.MatchString(aLine) {
		t.Error("queueActiveLine does not match")
	} else {
		if sMatch[1] != queueIdWanted {
			t.Errorf("queueActiveLine - cannot find queue ID in %q, wanted %q, got %q",
				aLine, queueIdWanted, sMatch[1])
		}
		if sMatch[2] != msgSizeWanted {
			t.Errorf("queueActiveLine - message size wanted %q, got %q", msgSizeWanted, sMatch[2])
		}
	}
}

func TestQueueRemoveLine(t *testing.T) {
	logLines := []string{
		"qmgr[4753]: 2B69A469711: removed",
		"postsuper[4753]: 2B69A469711: removed",
	}
	queueIdWanted := "2B69A469711"

	re, err := regexp.Compile(queueRemoveLine)
	if err != nil {
		t.Error("reQueueRemoveLine regexp compile error:", err)
	} else {
		for _, l := range logLines {
			sMatch := re.FindStringSubmatch(l)
			if sMatch == nil || len(sMatch) != 2 {
				t.Errorf("queueRemoveLine does not match to %q", l)
			} else if sMatch[1] != queueIdWanted {
				t.Errorf("queueRemoveLine - cannot find queue ID in %q, wanted %q, got %q",
					l, queueIdWanted, sMatch[1])
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
		t.Errorf("reForwardedLine does not match to %q", aLine)
	}
}

func TestDeliveredLine(t *testing.T) {
	aLine := "[30345]: 923745823B: to=<user@example.com>, relay=mail.exampleoutlook.com[123.123.123.123]:25, delay=1.5, delays=0.03/0.02/0.19/1.2, dsn=2.6.0, status=sent"
	queueIdWanted := "923745823B"

	re, err := regexp.Compile(deliveredLine)
	if err != nil {
		t.Error("reDeliveredLine regexp compile error:", err)
	} else {
		sMatch := re.FindStringSubmatch(aLine)
		if sMatch == nil || len(sMatch) != 2 {
			t.Errorf("deliveredLine does not match to %q", aLine)
		} else if sMatch[1] != queueIdWanted {
			t.Errorf("deliveredLine - cannot find queue ID in %q, wanted %q, got %q",
				aLine, queueIdWanted, sMatch[1])
		}
	}
}

/*
func TestForwardedLine(t *testing.T) {
	re, err := regexp.Compile(forwardedLine)
	if err != nil {
		t.Error("reForwardedLine regexp compile error:", err)
	} else if !re.MatchString(aLine) {
		t.Error("forwardedLine does not match")
	}
}

func TestForwardedLine(t *testing.T) {
	re, err := regexp.Compile(forwardedLine)
	if err != nil {
		t.Error("reForwardedLine regexp compile error:", err)
	} else if !re.MatchString(aLine) {
		t.Error("forwardedLine does not match")
	}
}
*/
