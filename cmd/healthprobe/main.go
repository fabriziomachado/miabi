// SPDX-FileCopyrightText: 2026 Jonas Kaninda
// SPDX-License-Identifier: AGPL-3.0-or-later

// Command healthprobe is the HTTP healthcheck Miabi copies into app containers,
// so an HTTP check works in images without a shell, curl or wget.
//
//	healthprobe <port> <path>
//
// It exits 0 when the app answers 2xx/3xx on loopback, 1 otherwise.
package main

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Docker kills the probe at the app's healthcheck timeout; this only bounds a
// probe run by hand.
const maxWait = 60 * time.Second

func main() {
	if len(os.Args) != 3 {
		say(os.Stderr, "usage: healthprobe <port> <path>\n")
		os.Exit(2)
	}
	port, path := os.Args[1], os.Args[2]
	os.Exit(probe([]string{"127.0.0.1", "::1"}, port, path, maxWait))
}

// probe GETs path on the first host that accepts a connection.
func probe(hosts []string, port, path string, timeout time.Duration) int {
	var conn net.Conn
	var err error
	for _, h := range hosts {
		if conn, err = net.DialTimeout("tcp", net.JoinHostPort(h, port), timeout); err == nil {
			break
		}
	}
	if err != nil {
		say(os.Stderr, "healthprobe: "+err.Error()+"\n")
		return 1
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	req := "GET " + path + " HTTP/1.1\r\nHost: localhost\r\nUser-Agent: miabi-healthprobe\r\nConnection: close\r\n\r\n"
	if _, err := conn.Write([]byte(req)); err != nil {
		say(os.Stderr, "healthprobe: "+err.Error()+"\n")
		return 1
	}
	line, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		say(os.Stderr, "healthprobe: no response: "+err.Error()+"\n")
		return 1
	}
	code := statusCode(line)
	say(os.Stdout, "GET "+path+" -> "+strings.TrimSpace(line)+"\n")
	if code >= 200 && code < 400 {
		return 0
	}
	return 1
}

// statusCode extracts the code from an HTTP status line, or 0.
func statusCode(line string) int {
	fields := strings.Fields(line)
	if len(fields) < 2 || !strings.HasPrefix(fields[0], "HTTP/") {
		return 0
	}
	n, _ := strconv.Atoi(fields[1])
	return n
}

func say(w *os.File, s string) { _, _ = w.WriteString(s) }
