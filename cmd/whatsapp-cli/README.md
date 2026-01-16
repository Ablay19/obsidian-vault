# WhatsApp CLI Tool

A command-line interface for WhatsApp Web API using go-whatsapp.

## Features

- QR code login
- Send text messages
- Receive messages
- Session persistence

## Usage

```bash
# Login (scan QR code)
whatsapp-cli login

# Send message
whatsapp-cli send "1234567890@s.whatsapp.net" "Hello World"

# Receive messages (listen mode)
whatsapp-cli receive

# Logout
whatsapp-cli logout
```

## Installation

```bash
go build -o whatsapp-cli cmd/whatsapp-cli/main.go
```

## Integration

This CLI is integrated with the project's WhatsApp bot service and can be used alongside the TUI for additional management.