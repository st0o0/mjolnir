package nut

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Querier interface {
	ListVars(upsName string) (map[string]string, error)
}

type UPS struct {
	Name        string
	Description string
}

type NUTError struct {
	Code string
}

func (e *NUTError) Error() string {
	return fmt.Sprintf("NUT error: %s", e.Code)
}

type Client struct {
	addr    string
	mu      sync.Mutex
	conn    net.Conn
	scanner *bufio.Scanner
}

func Dial(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) connect() error {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}
	c.conn = conn
	c.scanner = bufio.NewScanner(conn)
	return nil
}

func (c *Client) ensureConnected() error {
	if c.conn == nil {
		return c.connect()
	}
	return nil
}

func (c *Client) send(cmd string) error {
	_, err := fmt.Fprintf(c.conn, "%s\n", cmd)
	return err
}

func (c *Client) readLine() (string, error) {
	if c.scanner.Scan() {
		return c.scanner.Text(), nil
	}
	if err := c.scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("connection closed")
}

func (c *Client) disconnect() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
		c.scanner = nil
	}
}

func (c *Client) doWithRetry(fn func() error) error {
	err := fn()
	if err == nil {
		return nil
	}
	c.disconnect()
	if connErr := c.connect(); connErr != nil {
		return connErr
	}
	return fn()
}

func (c *Client) ListUPS() ([]UPS, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result []UPS
	err := c.doWithRetry(func() error {
		if err := c.ensureConnected(); err != nil {
			return err
		}
		if err := c.send("LIST UPS"); err != nil {
			return err
		}

		lines, err := c.readList("UPS")
		if err != nil {
			return err
		}

		result = nil
		for _, line := range lines {
			name, desc, ok := parseUPSLine(line)
			if !ok {
				continue
			}
			result = append(result, UPS{Name: name, Description: desc})
		}
		return nil
	})
	return result, err
}

func (c *Client) ListVars(upsName string) (map[string]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result map[string]string
	err := c.doWithRetry(func() error {
		if err := c.ensureConnected(); err != nil {
			return err
		}
		if err := c.send(fmt.Sprintf("LIST VAR %s", upsName)); err != nil {
			return err
		}

		lines, err := c.readList(fmt.Sprintf("VAR %s", upsName))
		if err != nil {
			return err
		}

		result = make(map[string]string, len(lines))
		for _, line := range lines {
			_, key, value, ok := parseVarLine(line)
			if !ok {
				continue
			}
			result[key] = value
		}
		return nil
	})
	return result, err
}

func (c *Client) GetVar(upsName, varName string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var result string
	err := c.doWithRetry(func() error {
		if err := c.ensureConnected(); err != nil {
			return err
		}
		if err := c.send(fmt.Sprintf("GET VAR %s %s", upsName, varName)); err != nil {
			return err
		}

		line, err := c.readLine()
		if err != nil {
			return err
		}
		if err := checkError(line); err != nil {
			return err
		}

		_, _, value, ok := parseVarLine(line)
		if !ok {
			return fmt.Errorf("unexpected response: %s", line)
		}
		result = value
		return nil
	})
	return result, err
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	_ = c.send("LOGOUT")
	_, _ = c.readLine()
	c.disconnect()
	return nil
}

func (c *Client) readList(listType string) ([]string, error) {
	beginPrefix := fmt.Sprintf("BEGIN LIST %s", listType)
	endPrefix := fmt.Sprintf("END LIST %s", listType)

	line, err := c.readLine()
	if err != nil {
		return nil, err
	}
	if err := checkError(line); err != nil {
		return nil, err
	}
	if line != beginPrefix {
		return nil, fmt.Errorf("expected %q, got %q", beginPrefix, line)
	}

	var lines []string
	for {
		line, err = c.readLine()
		if err != nil {
			return nil, err
		}
		if line == endPrefix {
			return lines, nil
		}
		lines = append(lines, line)
	}
}

func checkError(line string) error {
	if strings.HasPrefix(line, "ERR ") {
		return &NUTError{Code: strings.TrimPrefix(line, "ERR ")}
	}
	return nil
}

// parseVarLine parses "VAR <ups> <name> "<value>"" into (ups, name, value, ok).
func parseVarLine(line string) (string, string, string, bool) {
	if !strings.HasPrefix(line, "VAR ") {
		return "", "", "", false
	}
	rest := line[4:]

	upsEnd := strings.IndexByte(rest, ' ')
	if upsEnd < 0 {
		return "", "", "", false
	}
	ups := rest[:upsEnd]
	rest = rest[upsEnd+1:]

	nameEnd := strings.IndexByte(rest, ' ')
	if nameEnd < 0 {
		return "", "", "", false
	}
	name := rest[:nameEnd]
	value := rest[nameEnd+1:]

	value = strings.TrimPrefix(value, "\"")
	value = strings.TrimSuffix(value, "\"")

	return ups, name, value, true
}

// parseUPSLine parses "UPS <name> "<description>"" into (name, description, ok).
func parseUPSLine(line string) (string, string, bool) {
	if !strings.HasPrefix(line, "UPS ") {
		return "", "", false
	}
	rest := line[4:]

	nameEnd := strings.IndexByte(rest, ' ')
	if nameEnd < 0 {
		return rest, "", true
	}
	name := rest[:nameEnd]
	desc := rest[nameEnd+1:]

	desc = strings.TrimPrefix(desc, "\"")
	desc = strings.TrimSuffix(desc, "\"")

	return name, desc, true
}
