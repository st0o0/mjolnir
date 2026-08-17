package nut

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
)

func startMockNUT(t *testing.T, handler func(conn net.Conn)) (string, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer conn.Close()
				handler(conn)
			}()
		}
	}()

	return ln.Addr().String(), func() {
		ln.Close()
		wg.Wait()
	}
}

func defaultHandler(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		cmd := scanner.Text()
		switch {
		case cmd == "LIST UPS":
			fmt.Fprintln(conn, "BEGIN LIST UPS")
			fmt.Fprintln(conn, `UPS ecoflow "EcoFlow Delta"`)
			fmt.Fprintln(conn, `UPS apc "APC BackUPS"`)
			fmt.Fprintln(conn, "END LIST UPS")

		case strings.HasPrefix(cmd, "LIST VAR "):
			upsName := strings.TrimPrefix(cmd, "LIST VAR ")
			if upsName == "nonexistent" {
				fmt.Fprintln(conn, "ERR UNKNOWN-UPS")
				continue
			}
			fmt.Fprintf(conn, "BEGIN LIST VAR %s\n", upsName)
			fmt.Fprintf(conn, "VAR %s battery.charge \"87\"\n", upsName)
			fmt.Fprintf(conn, "VAR %s ups.status \"OL\"\n", upsName)
			fmt.Fprintf(conn, "VAR %s ups.load \"23\"\n", upsName)
			fmt.Fprintf(conn, "END LIST VAR %s\n", upsName)

		case strings.HasPrefix(cmd, "GET VAR "):
			parts := strings.SplitN(strings.TrimPrefix(cmd, "GET VAR "), " ", 2)
			if len(parts) != 2 {
				fmt.Fprintln(conn, "ERR INVALID-ARGUMENT")
				continue
			}
			ups, varName := parts[0], parts[1]
			if varName == "nonexistent.var" {
				fmt.Fprintln(conn, "ERR VAR-NOT-SUPPORTED")
				continue
			}
			fmt.Fprintf(conn, "VAR %s %s \"42\"\n", ups, varName)

		case cmd == "LOGOUT":
			fmt.Fprintln(conn, "OK Goodbye")
			return

		default:
			fmt.Fprintln(conn, "ERR UNKNOWN-COMMAND")
		}
	}
}

func TestListUPS(t *testing.T) {
	addr, stop := startMockNUT(t, defaultHandler)
	defer stop()

	c := Dial(addr)
	defer c.Close()

	ups, err := c.ListUPS()
	if err != nil {
		t.Fatal(err)
	}

	if len(ups) != 2 {
		t.Fatalf("expected 2 UPS units, got %d", len(ups))
	}
	if ups[0].Name != "ecoflow" || ups[0].Description != "EcoFlow Delta" {
		t.Fatalf("unexpected first UPS: %+v", ups[0])
	}
	if ups[1].Name != "apc" || ups[1].Description != "APC BackUPS" {
		t.Fatalf("unexpected second UPS: %+v", ups[1])
	}
}

func TestListVars(t *testing.T) {
	addr, stop := startMockNUT(t, defaultHandler)
	defer stop()

	c := Dial(addr)
	defer c.Close()

	vars, err := c.ListVars("ecoflow")
	if err != nil {
		t.Fatal(err)
	}

	if vars["battery.charge"] != "87" {
		t.Fatalf("expected battery.charge=87, got %s", vars["battery.charge"])
	}
	if vars["ups.status"] != "OL" {
		t.Fatalf("expected ups.status=OL, got %s", vars["ups.status"])
	}
	if vars["ups.load"] != "23" {
		t.Fatalf("expected ups.load=23, got %s", vars["ups.load"])
	}
	if len(vars) != 3 {
		t.Fatalf("expected 3 vars, got %d", len(vars))
	}
}

func TestListVarsUnknownUPS(t *testing.T) {
	addr, stop := startMockNUT(t, defaultHandler)
	defer stop()

	c := Dial(addr)
	defer c.Close()

	_, err := c.ListVars("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown UPS")
	}

	var nutErr *NUTError
	if !errors.As(err, &nutErr) {
		t.Fatalf("expected NUTError, got %T: %v", err, err)
	}
	if nutErr.Code != "UNKNOWN-UPS" {
		t.Fatalf("expected UNKNOWN-UPS, got %s", nutErr.Code)
	}
}

func TestGetVar(t *testing.T) {
	addr, stop := startMockNUT(t, defaultHandler)
	defer stop()

	c := Dial(addr)
	defer c.Close()

	val, err := c.GetVar("ecoflow", "battery.charge")
	if err != nil {
		t.Fatal(err)
	}
	if val != "42" {
		t.Fatalf("expected 42, got %s", val)
	}
}

func TestGetVarUnknown(t *testing.T) {
	addr, stop := startMockNUT(t, defaultHandler)
	defer stop()

	c := Dial(addr)
	defer c.Close()

	_, err := c.GetVar("ecoflow", "nonexistent.var")
	if err == nil {
		t.Fatal("expected error for unknown var")
	}

	var nutErr *NUTError
	if !errors.As(err, &nutErr) {
		t.Fatalf("expected NUTError, got %T: %v", err, err)
	}
	if nutErr.Code != "VAR-NOT-SUPPORTED" {
		t.Fatalf("expected VAR-NOT-SUPPORTED, got %s", nutErr.Code)
	}
}

func TestCloseIdempotent(t *testing.T) {
	addr, stop := startMockNUT(t, defaultHandler)
	defer stop()

	c := Dial(addr)

	if err := c.Close(); err != nil {
		t.Fatalf("first close on unconnected client: %v", err)
	}

	// Force a connection
	_, _ = c.ListUPS()

	if err := c.Close(); err != nil {
		t.Fatalf("close on connected client: %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestReconnectAfterServerRestart(t *testing.T) {
	connCount := 0
	var mu sync.Mutex
	var serverConns []net.Conn
	var connsMu sync.Mutex

	handler := func(conn net.Conn) {
		mu.Lock()
		connCount++
		mu.Unlock()
		connsMu.Lock()
		serverConns = append(serverConns, conn)
		connsMu.Unlock()
		defaultHandler(conn)
	}

	ln1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln1.Addr().String()

	go func() {
		for {
			conn, err := ln1.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				handler(conn)
			}()
		}
	}()

	c := Dial(addr)

	_, err = c.ListUPS()
	if err != nil {
		t.Fatal(err)
	}

	// Kill server and all existing connections
	ln1.Close()
	connsMu.Lock()
	for _, conn := range serverConns {
		conn.Close()
	}
	serverConns = nil
	connsMu.Unlock()

	// Start a new server on the same address
	ln2, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer ln2.Close()

	go func() {
		for {
			conn, err := ln2.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				handler(conn)
			}()
		}
	}()

	ups, err := c.ListUPS()
	if err != nil {
		t.Fatalf("expected reconnect to succeed: %v", err)
	}
	if len(ups) != 2 {
		t.Fatalf("expected 2 UPS after reconnect, got %d", len(ups))
	}

	mu.Lock()
	if connCount < 2 {
		t.Fatalf("expected at least 2 connections (reconnect), got %d", connCount)
	}
	mu.Unlock()

	c.Close()
}

func TestListUPSEmpty(t *testing.T) {
	handler := func(conn net.Conn) {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			cmd := scanner.Text()
			if cmd == "LIST UPS" {
				fmt.Fprintln(conn, "BEGIN LIST UPS")
				fmt.Fprintln(conn, "END LIST UPS")
			} else if cmd == "LOGOUT" {
				fmt.Fprintln(conn, "OK Goodbye")
				return
			}
		}
	}

	addr, stop := startMockNUT(t, handler)
	defer stop()

	c := Dial(addr)
	defer c.Close()

	ups, err := c.ListUPS()
	if err != nil {
		t.Fatal(err)
	}
	if len(ups) != 0 {
		t.Fatalf("expected 0 UPS units, got %d", len(ups))
	}
}

func TestListVarsEmptyVars(t *testing.T) {
	handler := func(conn net.Conn) {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			cmd := scanner.Text()
			if strings.HasPrefix(cmd, "LIST VAR ") {
				ups := strings.TrimPrefix(cmd, "LIST VAR ")
				fmt.Fprintf(conn, "BEGIN LIST VAR %s\n", ups)
				fmt.Fprintf(conn, "END LIST VAR %s\n", ups)
			} else if cmd == "LOGOUT" {
				fmt.Fprintln(conn, "OK Goodbye")
				return
			}
		}
	}

	addr, stop := startMockNUT(t, handler)
	defer stop()

	c := Dial(addr)
	defer c.Close()

	vars, err := c.ListVars("ecoflow")
	if err != nil {
		t.Fatal(err)
	}
	if len(vars) != 0 {
		t.Fatalf("expected 0 vars, got %d", len(vars))
	}
}

func TestConnectionRefused(t *testing.T) {
	c := Dial("127.0.0.1:1")

	_, err := c.ListUPS()
	if err == nil {
		t.Fatal("expected connection error")
	}
}

func TestQuerierInterface(t *testing.T) {
	var _ Querier = (*Client)(nil)
}

func TestParseVarLine(t *testing.T) {
	tests := []struct {
		line    string
		ups     string
		key     string
		value   string
		wantOK  bool
	}{
		{`VAR ecoflow battery.charge "87"`, "ecoflow", "battery.charge", "87", true},
		{`VAR my-ups ups.status "OL CHRG"`, "my-ups", "ups.status", "OL CHRG", true},
		{`VAR ups1 empty.val ""`, "ups1", "empty.val", "", true},
		{`NOT A VAR LINE`, "", "", "", false},
		{`VAR`, "", "", "", false},
	}

	for _, tt := range tests {
		ups, key, value, ok := parseVarLine(tt.line)
		if ok != tt.wantOK {
			t.Errorf("parseVarLine(%q): ok=%v, want %v", tt.line, ok, tt.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if ups != tt.ups || key != tt.key || value != tt.value {
			t.Errorf("parseVarLine(%q) = (%q, %q, %q), want (%q, %q, %q)", tt.line, ups, key, value, tt.ups, tt.key, tt.value)
		}
	}
}

func TestParseUPSLine(t *testing.T) {
	tests := []struct {
		line   string
		name   string
		desc   string
		wantOK bool
	}{
		{`UPS ecoflow "EcoFlow Delta"`, "ecoflow", "EcoFlow Delta", true},
		{`UPS simple`, "simple", "", true},
		{`NOT UPS`, "", "", false},
	}

	for _, tt := range tests {
		name, desc, ok := parseUPSLine(tt.line)
		if ok != tt.wantOK {
			t.Errorf("parseUPSLine(%q): ok=%v, want %v", tt.line, ok, tt.wantOK)
			continue
		}
		if !ok {
			continue
		}
		if name != tt.name || desc != tt.desc {
			t.Errorf("parseUPSLine(%q) = (%q, %q), want (%q, %q)", tt.line, name, desc, tt.name, tt.desc)
		}
	}
}
