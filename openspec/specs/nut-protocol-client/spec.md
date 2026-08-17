## Purpose

Provide a native Go client for the NUT (Network UPS Tools) protocol, enabling direct TCP communication with NUT servers for querying UPS status and variables without relying on the `upsc` command-line tool.

## Requirements

### Requirement: NUT protocol TCP connection
The system SHALL provide a Go client that connects to a NUT server via TCP on a configurable address (default `localhost:3493`). The client SHALL maintain a persistent connection and automatically reconnect on connection loss.

#### Scenario: Successful connection
- **WHEN** a NUT client is created with address `localhost:3493` and the NUT server is running
- **THEN** the client establishes a TCP connection and is ready to send commands

#### Scenario: Connection refused
- **WHEN** a NUT client attempts to connect and no NUT server is listening
- **THEN** the client returns a typed connection error

#### Scenario: Automatic reconnect after disconnect
- **WHEN** the TCP connection drops during operation
- **THEN** the client reconnects transparently on the next command invocation

---

### Requirement: LIST UPS command
The system SHALL support the `LIST UPS` NUT protocol command, returning all UPS units known to the server with their names and descriptions.

#### Scenario: Single UPS available
- **WHEN** the client sends `LIST UPS` and the server has one UPS named `ecoflow` with description `"EcoFlow Delta"`
- **THEN** the client returns a list containing one entry with name `ecoflow` and description `EcoFlow Delta`

#### Scenario: Multiple UPS units
- **WHEN** the client sends `LIST UPS` and the server has two UPS units
- **THEN** the client returns a list with both entries

#### Scenario: No UPS units
- **WHEN** the client sends `LIST UPS` and the server has no UPS configured
- **THEN** the client returns an empty list

---

### Requirement: LIST VAR command
The system SHALL support the `LIST VAR <upsname>` NUT protocol command, returning all variables for a UPS as key-value pairs.

#### Scenario: Query all variables
- **WHEN** the client sends `LIST VAR ecoflow` and the UPS has variables `battery.charge=87` and `ups.status=OL`
- **THEN** the client returns a map with keys `battery.charge` and `ups.status` and their respective values

#### Scenario: Unknown UPS name
- **WHEN** the client sends `LIST VAR nonexistent`
- **THEN** the client returns an error indicating the UPS was not found (NUT ERR response)

---

### Requirement: GET VAR command
The system SHALL support the `GET VAR <upsname> <varname>` NUT protocol command, returning a single variable value.

#### Scenario: Query single variable
- **WHEN** the client sends `GET VAR ecoflow battery.charge` and the value is `87`
- **THEN** the client returns the string `87`

#### Scenario: Unknown variable
- **WHEN** the client sends `GET VAR ecoflow nonexistent.var`
- **THEN** the client returns an error indicating the variable was not found

---

### Requirement: NUT protocol response parsing
The system SHALL parse NUT protocol responses according to the line-based text protocol format: `BEGIN`/`END` list blocks, `VAR <ups> <name> "<value>"` lines, `OK` success, and `ERR <code> <message>` errors.

#### Scenario: Parse VAR line with quoted value
- **WHEN** the server responds with `VAR ecoflow battery.charge "87"`
- **THEN** the parser extracts UPS name `ecoflow`, variable `battery.charge`, value `87` (quotes stripped)

#### Scenario: Parse ERR response
- **WHEN** the server responds with `ERR VAR-NOT-SUPPORTED`
- **THEN** the parser returns a typed error with error code `VAR-NOT-SUPPORTED`

#### Scenario: Parse multi-line LIST response
- **WHEN** the server responds with a `BEGIN LIST VAR` / `END LIST VAR` block containing multiple `VAR` lines
- **THEN** the parser collects all VAR lines and returns them as a complete result

---

### Requirement: Client graceful shutdown
The system SHALL support closing the client connection gracefully by sending `LOGOUT` before closing the TCP socket.

#### Scenario: Graceful close
- **WHEN** the client's `Close()` method is called
- **THEN** the client sends `LOGOUT`, reads the `OK Goodbye` response, and closes the TCP connection

#### Scenario: Close on already-disconnected client
- **WHEN** `Close()` is called but the connection is already broken
- **THEN** the client returns without error (idempotent close)
