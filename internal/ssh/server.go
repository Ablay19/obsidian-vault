package ssh

import (
	"fmt"
	"io"
	"net"
	"net/http" // Add this import
	"os"
	"strconv" // Add this import

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

func StartServer() {
	InitDB() // Initialize the SSH user database

	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			userEntry := logrus.WithFields(logrus.Fields{
				"user":   conn.User(),
				"method": "password",
				"remote": conn.RemoteAddr(),
			})
			var user User
			if err := DB.Where("username = ?", conn.User()).First(&user).Error; err != nil {
				userEntry.Errorf("Authentication failed: user not found - %v", err)
				return nil, fmt.Errorf("user not found")
			}
			// TODO: Implement proper password hashing and comparison
			if user.Password != string(password) {
				userEntry.Errorf("Authentication failed: password mismatch")
				return nil, fmt.Errorf("password mismatch")
			}
			userEntry.Info("Authentication successful")
			return nil, nil
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			userEntry := logrus.WithFields(logrus.Fields{
				"user":        conn.User(),
				"method":      "publickey",
				"fingerprint": ssh.FingerprintSHA256(key),
				"remote":      conn.RemoteAddr(),
			})
			var user User
			if err := DB.Where("username = ?", conn.User()).First(&user).Error; err != nil {
				userEntry.Errorf("Authentication failed: user not found - %v", err)
				return nil, fmt.Errorf("user not found")
			}
			// TODO: Implement proper public key comparison
			if user.PublicKey != string(ssh.MarshalAuthorizedKey(key)) {
				userEntry.Errorf("Authentication failed: public key mismatch")
				return nil, fmt.Errorf("public key mismatch")
			}
			userEntry.Info("Authentication successful")
			return nil, nil
		},
	}

	// You can generate a key with 'ssh-keygen -t rsa'
	privateBytes, err := os.ReadFile("id_rsa")
	if err != nil {
		logrus.Fatalf("Failed to load private key (./id_rsa): %v", err)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		logrus.Fatalf("Failed to parse private key: %v", err)
	}

	config.AddHostKey(private)

	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		logrus.Fatalf("Failed to listen on 2222: %v", err)
	}

	go StartAPI()

	logrus.Info("Listening on 2222...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Errorf("Failed to accept incoming connection: %v", err)
			continue
		}

		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
		if err != nil {
			logrus.Errorf("Failed to handshake: %v", err)
			continue
		}
		logrus.Infof("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())

		// Discard all channels and requests
		go ssh.DiscardRequests(reqs)
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		logrus.Errorf("could not accept channel: %v", err)
		return
	}

	switch newChannel.ChannelType() {
	case "session":
		go handleSession(channel, requests)
	case "direct-tcpip":
		go handleDirectTCPIP(channel, newChannel.ExtraData())
	case "forwarded-tcpip":
		// This is for reverse tunneling; not implemented yet
		newChannel.Reject(ssh.Prohibited, "forwarded-tcpip not supported yet")
		channel.Close()
	default:
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", newChannel.ChannelType()))
		channel.Close()
	}
}

func handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()
	for req := range requests {
		logrus.Infof("session request type: %s", req.Type)
		switch req.Type {
		case "shell":
			req.Reply(true, nil)
			// For a simple proxy, we don't provide a shell. Close the channel.
			return
		case "pty-req":
			req.Reply(true, nil)
		case "env":
			req.Reply(true, nil)
		default:
			req.Reply(false, nil)
		}
	}
}

func handleDirectTCPIP(channel ssh.Channel, extraData []byte) {
	defer channel.Close()

	var payload struct {
		DestAddr string
		DestPort uint32
		SrcAddr  string
		SrcPort  uint32
	}
	if err := ssh.Unmarshal(extraData, &payload); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Errorf("failed to unmarshal direct-tcpip payload")
		return
	}

	dest := net.JoinHostPort(payload.DestAddr, strconv.Itoa(int(payload.DestPort)))
	connEntry := logrus.WithFields(logrus.Fields{
		"destination_address": payload.DestAddr,
		"destination_port":    payload.DestPort,
		"source_address":      payload.SrcAddr,
		"source_port":         payload.SrcPort,
	})

	connEntry.Info("Attempting direct-tcpip connection")
	conn, err := net.Dial("tcp", dest)
	if err != nil {
		connEntry.Errorf("failed to dial destination: %v", err)
		return
	}
	defer conn.Close() // Defer closing the connection, handles both read and write ends

	connEntry.Info("Direct-tcpip connection established")

	// Copy data from SSH channel to destination connection
	go func() {
		_, err := io.Copy(channel, conn)
		if err != nil {
			connEntry.Errorf("Error copying from remote to channel: %v", err)
		}
	}()

	// Copy data from destination connection to SSH channel
	_, err = io.Copy(conn, channel)
	if err != nil {
		connEntry.Errorf("Error copying from channel to remote: %v", err)
	}
}

// StartAPI initializes and starts the REST API for SSH user management.
func StartAPI() {
	router := mux.NewRouter()
	apiHandler := UserAPIHandler{}
	apiHandler.RegisterRoutes(router)

	logrus.Infof("SSH Management API listening on :8081")
	if err := http.ListenAndServe(":8081", router); err != nil {
		logrus.Fatalf("SSH Management API failed to start: %v", err)
	}
}

