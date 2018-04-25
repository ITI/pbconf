package main

/***********************************************************************
   Copyright 2018 Information Trust Institute

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
***********************************************************************/

import (
	"fmt"
	database "github.com/iti/pbconf/lib/pbdatabase"
	"net"

	ssh "golang.org/x/crypto/ssh"
)

func ssh_server(ServerConfig *ssh.ServerConfig, db database.AppDatabase) {
	log.Debug("listening")

	listener, err := net.Listen("tcp", cfg.Broker.Listen)
	if err != nil {
		panic("failed to listen for connection")
	}

	log.Debug("Accepting")
	for {

		// Accept an incomming tcp Connection
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Info("Failed to accept incoming connection (%s)", err)
			continue
		}

		//Turn tcpConn into an SSH Connection
		srvConn, srvChannels, srvRequests, err := ssh.NewServerConn(tcpConn, ServerConfig)
		if err != nil {
			log.Debug("Failed SSH handshake (%s)", err)
			continue
		}
		_ = srvConn
		log.Info(fmt.Sprintf("New SSH connection from %s (%s)", srvConn.RemoteAddr(), srvConn.ClientVersion()))

		// Discard all global out-of-band Requests
		go ssh.DiscardRequests(srvRequests)
		// Accept all channels
		go handleChannels(srvChannels, db)
	}
}

func handleChannels(chans <-chan ssh.NewChannel, db database.AppDatabase) {
	log.Debug("Handling channels")
	// Service the incoming SSH Channel go-channel in go routine
	for newChannel := range chans {
		go handleChannel(newChannel, db)
	}
}

func handleChannel(newChannel ssh.NewChannel, db database.AppDatabase) {
	log.Debug("Handling channel")
	// Since we're handling a shell, we expect a
	// channel type of "session". This also describes
	// "x11", "direct-tcpip" and "forwarded-tcpip"
	// channel types.
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	// Accept the SSH Sesion Channel - We should now have a fully set up and
	// working SSH connection with the client.
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Debug("Could not accept channel: " + err.Error())
		return
	}
	log.Debug("Accepted Channel")

	// Handle out-of-band SSH Channel requests
	go func() {
		log.Debug("Handling requests")
		for req := range requests {
			log.Debug(fmt.Sprintf("request type: %s", req.Type))
			// We only handle shell and exec channels
			switch req.Type {
			case "shell":
				// Print a menu
				log.Debug("shell request")
				connection.Write([]byte("You requested a shell\r\n"))
				shell(connection, db)
				connection.Close()
			case "exec":
				log.Debug("exec request")
				// No need for a menu, set up the client connnection
				// based on req.Payload
				// Note: first 4 bytes is the length of payload+padding
				name := string(req.Payload[4:])
				connection.Write([]byte(fmt.Sprintf("You requested a direct connection to %s\r\n", name)))
				direct(name, connection, db)
				connection.Close()
			}
		}
	}()
}
