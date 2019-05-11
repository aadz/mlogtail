package main

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type MsgStatusCountersType struct {
	mutex       *sync.Mutex
	counters    map[string]uint64
	bytesDlvMap map[string]uint64
	bytesRcvMap map[string]bool
}

const (
	PostfixLogLine  = `^[JAMDFONS][aeucop][nrbcglptvy] [1-3 ]\d [0-2]\d:[0-5]\d:[0-5]\d \S+ postfix/[a-z]+\[\d+\]: `
	ReceivedLine    = `/(?:(?:smtpd|pickup)\[\d+\]: ([\dA-F]+): (?:client|uid|sender)=)`
	QueueActiveLine = `/qmgr\[\d+\]: ([\dA-F]+): .* size=(\d+)[, ].+queue active`
	QueueRemoveLine = `/(?:qmgr|postsuper)\[\d+\]: ([\dA-F]+): removed`
	DeliveredLine   = `\[\d+\]: ([\dA-F]+): .+ status=sent`
	ForwardedLine   = `forwarded as `
	DeferredLine    = `\[\d+\]: ([\dA-F]+): .+ status=deferred`
	BouncedLine     = `\[\d+\]: ([\dA-F]+): .+ status=bounced`
	RejectLine      = `/(?:smtpd|cleanup)\[\d+\]: .*?\breject: `
	HoldLine        = ` NOQUEUE: hold: `
	DiscardLine     = `: NOQUEUE: discard: `
)

var (
	needMx           bool // Need mutex
	PostfixStatusArr = [10]string{"bytes-received", "bytes-delivered",
		"received", "delivered", "forwarded", "deferred",
		"bounced", "rejected", "held", "discarded"}
	rePostfixLogLine  = regexp.MustCompile(PostfixLogLine)
	reReceivedLine    = regexp.MustCompile(ReceivedLine)
	reQueueActiveLine = regexp.MustCompile(QueueActiveLine)
	reQueueRemoveLine = regexp.MustCompile(QueueRemoveLine)
	reDeliveredLine   = regexp.MustCompile(DeliveredLine)
	reForwardedLine   = regexp.MustCompile(ForwardedLine)
	reDeferredLine    = regexp.MustCompile(DeferredLine)
	reBouncedLine     = regexp.MustCompile(BouncedLine)
	reRejectLine      = regexp.MustCompile(RejectLine)
	reHoldLine        = regexp.MustCompile(HoldLine)
	reDiscardLine     = regexp.MustCompile(DiscardLine)
	msgStatusCounters MsgStatusCountersType
)

func IsPostfixLine(s string) bool {
	return rePostfixLogLine.MatchString(s)
}

func PostfixCmgHandle(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Cannot accept a connection: %s\n", err)
		} else {
			go postfixCmdProcess(conn)
		}
	}
}

func PostfuxLineParse(s string) {
	var statusKey string
	if sMatch := reReceivedLine.FindStringSubmatch(s); sMatch != nil { // received
		statusKey = "received"
		msgStatusCounters.lock()
		msgStatusCounters.bytesRcvMap[sMatch[1]] = true
		msgStatusCounters.unlock()
	} else if sMatch := reQueueActiveLine.FindStringSubmatch(s); sMatch != nil { // queue active
		sz, err := strconv.Atoi(sMatch[2])
		if err != nil {
			fmt.Printf("Cannot convert to a number: %s\n", err)
		}
		msgStatusCounters.lock()
		msgStatusCounters.bytesDlvMap[sMatch[1]] = uint64(sz)
		if msgStatusCounters.bytesRcvMap[sMatch[1]] {
			msgStatusCounters.counters["bytes-received"] += uint64(sz)
			delete(msgStatusCounters.bytesRcvMap, sMatch[1])
		}
		msgStatusCounters.unlock()
	} else if sMatch := reQueueRemoveLine.FindStringSubmatch(s); sMatch != nil { // removed
		msgStatusCounters.lock()
		delete(msgStatusCounters.bytesDlvMap, sMatch[1])
		msgStatusCounters.unlock()
	} else if reForwardedLine.MatchString(s) { // forwarded
		statusKey = "forwarded"
	} else if sMatch := reDeliveredLine.FindStringSubmatch(s); sMatch != nil { // sent
		statusKey = "delivered"
		msgStatusCounters.lock()
		msgStatusCounters.counters["bytes-delivered"] += msgStatusCounters.bytesDlvMap[sMatch[1]]
		msgStatusCounters.unlock()
	} else if reRejectLine.MatchString(s) { // rejected
		statusKey = "rejected"
	} else if reDeferredLine.MatchString(s) { // deffered
		statusKey = "deferred"
	} else if reBouncedLine.MatchString(s) { // bounced
		statusKey = "bounced"
	} else if reHoldLine.MatchString(s) { // held
		statusKey = "held"
	} else if reDiscardLine.MatchString(s) { // discarded
		statusKey = "discarded"
	}
	if len(statusKey) != 0 {
		msgStatusCounters.lock()
		msgStatusCounters.counters[statusKey]++
		msgStatusCounters.unlock()
	}
}

func PostfixParserInit(cfg *Config) {
	msgStatusCounters.reset()
	if cfg.cmd == "tail" {
		needMx = true
		msgStatusCounters.mutex = new(sync.Mutex)
	}
}

// PostfixStats is called in fail reaging mode (not while tailing)
// so we do not use locks here.
func PostfixStats() string {
	return msgStatusCounters.String()
}

func (c *MsgStatusCountersType) lock() {
	// we do not perform locking here if we are reading file
	if needMx {
		c.mutex.Lock()
	}
}

func (c *MsgStatusCountersType) reset() {
	c.counters = make(map[string]uint64, 8)
	c.bytesRcvMap = make(map[string]bool)
	c.bytesDlvMap = make(map[string]uint64)
}

func (c *MsgStatusCountersType) String() string {
	var res string
	for _, s := range PostfixStatusArr {
		res += fmt.Sprintf("%-16s%d\n", s, c.counters[s])
	}
	return res
}

func (c *MsgStatusCountersType) unlock() {
	// we do not perform locking if we are reading file
	if needMx {
		c.mutex.Unlock()
	}
}

func postfixCmdProcess(conn net.Conn) {
	buf := make([]byte, 32)
	cnt, err := conn.Read(buf)
	if err != nil {
		// conn.Close()
		return
	}
	cmd := strings.TrimSpace(string(buf[:cnt]))

	var resp string
	if cmd == "stats" {
		msgStatusCounters.lock()
		resp = msgStatusCounters.String()
		msgStatusCounters.unlock()
	} else if cmd == "stats_reset" {
		msgStatusCounters.lock()
		resp = msgStatusCounters.String()
		msgStatusCounters.reset()
		msgStatusCounters.unlock()
	} else if cmd == "reset" {
		msgStatusCounters.lock()
		msgStatusCounters.reset()
		msgStatusCounters.unlock()
	} else {
		msgStatusCounters.lock()
		resp = fmt.Sprintf("%d\n", msgStatusCounters.counters[cmd])
		msgStatusCounters.unlock()
	}

	conn.Write([]byte(resp))
	conn.Close()
}