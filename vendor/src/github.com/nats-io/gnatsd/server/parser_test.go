// Copyright 2012-2015 Apcera Inc. All rights reserved.

package server

import (
	"bytes"
	"testing"
)

func dummyClient() *client {
	return &client{}
}

func dummyRouteClient() *client {
	return &client{typ: ROUTER}
}

func TestParsePing(t *testing.T) {
	c := dummyClient()
	if c.state != OP_START {
		t.Fatalf("Expected OP_START vs %d\n", c.state)
	}
	ping := []byte("PING\r\n")
	err := c.parse(ping[:1])
	if err != nil || c.state != OP_P {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(ping[1:2])
	if err != nil || c.state != OP_PI {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(ping[2:3])
	if err != nil || c.state != OP_PIN {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(ping[3:4])
	if err != nil || c.state != OP_PING {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(ping[4:5])
	if err != nil || c.state != OP_PING {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(ping[5:6])
	if err != nil || c.state != OP_START {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(ping)
	if err != nil || c.state != OP_START {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	// Should tolerate spaces
	ping = []byte("PING  \r")
	err = c.parse(ping)
	if err != nil || c.state != OP_PING {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	c.state = OP_START
	ping = []byte("PING  \r  \n")
	err = c.parse(ping)
	if err != nil || c.state != OP_START {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
}

func TestParsePong(t *testing.T) {
	c := dummyClient()
	if c.state != OP_START {
		t.Fatalf("Expected OP_START vs %d\n", c.state)
	}
	pong := []byte("PONG\r\n")
	err := c.parse(pong[:1])
	if err != nil || c.state != OP_P {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(pong[1:2])
	if err != nil || c.state != OP_PO {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(pong[2:3])
	if err != nil || c.state != OP_PON {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(pong[3:4])
	if err != nil || c.state != OP_PONG {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(pong[4:5])
	if err != nil || c.state != OP_PONG {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(pong[5:6])
	if err != nil || c.state != OP_START {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	err = c.parse(pong)
	if err != nil || c.state != OP_START {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	// Should tolerate spaces
	pong = []byte("PONG  \r")
	err = c.parse(pong)
	if err != nil || c.state != OP_PONG {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	c.state = OP_START
	pong = []byte("PONG  \r  \n")
	err = c.parse(pong)
	if err != nil || c.state != OP_START {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}

	// Should be adjusting c.pout, Pings Outstanding
	c.state = OP_START
	c.pout = 10
	pong = []byte("PONG\r\n")
	err = c.parse(pong)
	if err != nil || c.state != OP_START {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	if c.pout != 9 {
		t.Fatalf("Unexpected pout: %d vs %d\n", c.pout, 9)
	}
}

func TestParseConnect(t *testing.T) {
	c := dummyClient()
	connect := []byte("CONNECT {\"verbose\":false,\"pedantic\":true,\"ssl_required\":false}\r\n")
	err := c.parse(connect)
	if err != nil || c.state != OP_START {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	// Check saved state
	if c.as != 8 {
		t.Fatalf("ArgStart state incorrect: 8 vs %d\n", c.as)
	}
}

func TestParseSub(t *testing.T) {
	c := dummyClient()
	sub := []byte("SUB foo 1\r")
	err := c.parse(sub)
	if err != nil || c.state != SUB_ARG {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	// Check saved state
	if c.as != 4 {
		t.Fatalf("ArgStart state incorrect: 4 vs %d\n", c.as)
	}
	if c.drop != 1 {
		t.Fatalf("Drop state incorrect: 1 vs %d\n", c.as)
	}
	if !bytes.Equal(sub[c.as:], []byte("foo 1\r")) {
		t.Fatalf("Arg state incorrect: %s\n", sub[c.as:])
	}
}

func TestParsePub(t *testing.T) {
	c := dummyClient()

	pub := []byte("PUB foo 5\r\nhello\r")
	err := c.parse(pub)
	if err != nil || c.state != MSG_END {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	if !bytes.Equal(c.pa.subject, []byte("foo")) {
		t.Fatalf("Did not parse subject correctly: 'foo' vs '%s'\n", string(c.pa.subject))
	}
	if c.pa.reply != nil {
		t.Fatalf("Did not parse reply correctly: 'nil' vs '%s'\n", string(c.pa.reply))
	}
	if c.pa.size != 5 {
		t.Fatalf("Did not parse msg size correctly: 5 vs %d\n", c.pa.size)
	}

	// Clear snapshots
	c.argBuf, c.msgBuf, c.state = nil, nil, OP_START

	pub = []byte("PUB foo.bar INBOX.22 11\r\nhello world\r")
	err = c.parse(pub)
	if err != nil || c.state != MSG_END {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	if !bytes.Equal(c.pa.subject, []byte("foo.bar")) {
		t.Fatalf("Did not parse subject correctly: 'foo' vs '%s'\n", string(c.pa.subject))
	}
	if !bytes.Equal(c.pa.reply, []byte("INBOX.22")) {
		t.Fatalf("Did not parse reply correctly: 'INBOX.22' vs '%s'\n", string(c.pa.reply))
	}
	if c.pa.size != 11 {
		t.Fatalf("Did not parse msg size correctly: 11 vs %d\n", c.pa.size)
	}
}

func testPubArg(c *client, t *testing.T) {
	if !bytes.Equal(c.pa.subject, []byte("foo")) {
		t.Fatalf("Mismatched subject: '%s'\n", c.pa.subject)
	}
	if !bytes.Equal(c.pa.szb, []byte("22")) {
		t.Fatalf("Bad size buf: '%s'\n", c.pa.szb)
	}
	if c.pa.size != 22 {
		t.Fatalf("Bad size: %d\n", c.pa.size)
	}
}

func TestParsePubArg(t *testing.T) {
	c := dummyClient()
	if err := c.processPub([]byte("foo 22")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	testPubArg(c, t)
	if err := c.processPub([]byte(" foo 22")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	testPubArg(c, t)
	if err := c.processPub([]byte(" foo 22 ")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	testPubArg(c, t)
	if err := c.processPub([]byte("foo   22")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	if err := c.processPub([]byte("foo   22\r")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	testPubArg(c, t)
}

func TestParsePubBadSize(t *testing.T) {
	c := dummyClient()
	// Setup localized max payload
	c.mpay = 32768
	if err := c.processPub([]byte("foo 2222222222222222\r")); err == nil {
		t.Fatalf("Expected parse error for size too large")
	}
}

func TestParseMsg(t *testing.T) {
	c := dummyRouteClient()

	pub := []byte("MSG foo RSID:1:2 5\r\nhello\r")
	err := c.parse(pub)
	if err != nil || c.state != MSG_END {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	if !bytes.Equal(c.pa.subject, []byte("foo")) {
		t.Fatalf("Did not parse subject correctly: 'foo' vs '%s'\n", c.pa.subject)
	}
	if c.pa.reply != nil {
		t.Fatalf("Did not parse reply correctly: 'nil' vs '%s'\n", c.pa.reply)
	}
	if c.pa.size != 5 {
		t.Fatalf("Did not parse msg size correctly: 5 vs %d\n", c.pa.size)
	}
	if !bytes.Equal(c.pa.sid, []byte("RSID:1:2")) {
		t.Fatalf("Did not parse sid correctly: 'RSID:1:2' vs '%s'\n", c.pa.sid)
	}

	// Clear snapshots
	c.argBuf, c.msgBuf, c.state = nil, nil, OP_START

	pub = []byte("MSG foo.bar RSID:1:2 INBOX.22 11\r\nhello world\r")
	err = c.parse(pub)
	if err != nil || c.state != MSG_END {
		t.Fatalf("Unexpected: %d : %v\n", c.state, err)
	}
	if !bytes.Equal(c.pa.subject, []byte("foo.bar")) {
		t.Fatalf("Did not parse subject correctly: 'foo' vs '%s'\n", c.pa.subject)
	}
	if !bytes.Equal(c.pa.reply, []byte("INBOX.22")) {
		t.Fatalf("Did not parse reply correctly: 'INBOX.22' vs '%s'\n", c.pa.reply)
	}
	if c.pa.size != 11 {
		t.Fatalf("Did not parse msg size correctly: 11 vs %d\n", c.pa.size)
	}
}

func testMsgArg(c *client, t *testing.T) {
	if !bytes.Equal(c.pa.subject, []byte("foobar")) {
		t.Fatalf("Mismatched subject: '%s'\n", c.pa.subject)
	}
	if !bytes.Equal(c.pa.szb, []byte("22")) {
		t.Fatalf("Bad size buf: '%s'\n", c.pa.szb)
	}
	if c.pa.size != 22 {
		t.Fatalf("Bad size: %d\n", c.pa.size)
	}
	if !bytes.Equal(c.pa.sid, []byte("RSID:22:1")) {
		t.Fatalf("Bad sid: '%s'\n", c.pa.sid)
	}
}

func TestParseMsgArg(t *testing.T) {
	c := dummyClient()
	if err := c.processMsgArgs([]byte("foobar RSID:22:1 22")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	testMsgArg(c, t)
	if err := c.processMsgArgs([]byte(" foobar RSID:22:1 22")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	testMsgArg(c, t)
	if err := c.processMsgArgs([]byte(" foobar   RSID:22:1 22 ")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	testMsgArg(c, t)
	if err := c.processMsgArgs([]byte("foobar   RSID:22:1  \t22")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	if err := c.processMsgArgs([]byte("foobar\t\tRSID:22:1\t22\r")); err != nil {
		t.Fatalf("Unexpected parse error: %v\n", err)
	}
	testMsgArg(c, t)
}

func TestParseMsgSpace(t *testing.T) {
	c := dummyRouteClient()

	// Ivan bug he found
	if err := c.parse([]byte("MSG \r\n")); err == nil {
		t.Fatalf("Expected parse error for MSG <SPC>")
	}

	c = dummyClient()

	// Anything with an M from a client should parse error
	if err := c.parse([]byte("M")); err == nil {
		t.Fatalf("Expected parse error for M* from a client")
	}
}

func TestShouldFail(t *testing.T) {
	c := dummyClient()

	if err := c.parse([]byte(" PING")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("CONNECT \r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("POO")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("PUB foo\r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("PUB \r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("PUB foo bar       \r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("SUB\r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("SUB \r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("SUB foo\r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("SUB foo bar baz 22\r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("PUB foo 2\r\nok \r\n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
	c.state = OP_START
	if err := c.parse([]byte("PUB foo 2\r\nok\r \n")); err == nil {
		t.Fatal("Should have received a parse error")
	}
}

func TestProtoSnippet(t *testing.T) {
	sample := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	tests := []struct {
		input    int
		expected string
	}{
		{0, `"abcdefghijklmnopqrstuvwxyzABCDEF"`},
		{1, `"bcdefghijklmnopqrstuvwxyzABCDEFG"`},
		{2, `"cdefghijklmnopqrstuvwxyzABCDEFGH"`},
		{3, `"defghijklmnopqrstuvwxyzABCDEFGHI"`},
		{4, `"efghijklmnopqrstuvwxyzABCDEFGHIJ"`},
		{5, `"fghijklmnopqrstuvwxyzABCDEFGHIJK"`},
		{6, `"ghijklmnopqrstuvwxyzABCDEFGHIJKL"`},
		{7, `"hijklmnopqrstuvwxyzABCDEFGHIJKLM"`},
		{8, `"ijklmnopqrstuvwxyzABCDEFGHIJKLMN"`},
		{9, `"jklmnopqrstuvwxyzABCDEFGHIJKLMNO"`},
		{10, `"klmnopqrstuvwxyzABCDEFGHIJKLMNOP"`},
		{11, `"lmnopqrstuvwxyzABCDEFGHIJKLMNOPQ"`},
		{12, `"mnopqrstuvwxyzABCDEFGHIJKLMNOPQR"`},
		{13, `"nopqrstuvwxyzABCDEFGHIJKLMNOPQRS"`},
		{14, `"opqrstuvwxyzABCDEFGHIJKLMNOPQRST"`},
		{15, `"pqrstuvwxyzABCDEFGHIJKLMNOPQRSTU"`},
		{16, `"qrstuvwxyzABCDEFGHIJKLMNOPQRSTUV"`},
		{17, `"rstuvwxyzABCDEFGHIJKLMNOPQRSTUVW"`},
		{18, `"stuvwxyzABCDEFGHIJKLMNOPQRSTUVWX"`},
		{19, `"tuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"`},
		{20, `"uvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"`},
		{21, `"vwxyzABCDEFGHIJKLMNOPQRSTUVWXY"`},
		{22, `"wxyzABCDEFGHIJKLMNOPQRSTUVWXY"`},
		{23, `"xyzABCDEFGHIJKLMNOPQRSTUVWXY"`},
		{24, `"yzABCDEFGHIJKLMNOPQRSTUVWXY"`},
		{25, `"zABCDEFGHIJKLMNOPQRSTUVWXY"`},
		{26, `"ABCDEFGHIJKLMNOPQRSTUVWXY"`},
		{27, `"BCDEFGHIJKLMNOPQRSTUVWXY"`},
		{28, `"CDEFGHIJKLMNOPQRSTUVWXY"`},
		{29, `"DEFGHIJKLMNOPQRSTUVWXY"`},
		{30, `"EFGHIJKLMNOPQRSTUVWXY"`},
		{31, `"FGHIJKLMNOPQRSTUVWXY"`},
		{32, `"GHIJKLMNOPQRSTUVWXY"`},
		{33, `"HIJKLMNOPQRSTUVWXY"`},
		{34, `"IJKLMNOPQRSTUVWXY"`},
		{35, `"JKLMNOPQRSTUVWXY"`},
		{36, `"KLMNOPQRSTUVWXY"`},
		{37, `"LMNOPQRSTUVWXY"`},
		{38, `"MNOPQRSTUVWXY"`},
		{39, `"NOPQRSTUVWXY"`},
		{40, `"OPQRSTUVWXY"`},
		{41, `"PQRSTUVWXY"`},
		{42, `"QRSTUVWXY"`},
		{43, `"RSTUVWXY"`},
		{44, `"STUVWXY"`},
		{45, `"TUVWXY"`},
		{46, `"UVWXY"`},
		{47, `"VWXY"`},
		{48, `"WXY"`},
		{49, `"XY"`},
		{50, `"Y"`},
		{51, `""`},
		{52, `""`},
		{53, `""`},
		{54, `""`},
	}

	for _, tt := range tests {
		got := protoSnippet(tt.input, sample)
		if tt.expected != got {
			t.Errorf("Expected protocol snippet to be %s when start=%d but got %s\n", tt.expected, tt.input, got)
		}
	}
}
