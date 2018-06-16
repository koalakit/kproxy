package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"net"
)

type TCPServer struct {
	ConnectManager

	listener  net.Listener
	slaveConn net.Conn
}

func (server *TCPServer) Start(addr string) {
	server.init()

	var err error
	server.listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("ProxyMaster:", err)
	}

	log.Println("[TCPServer]", addr)

	go server.loop()
	go server.slaveHandle()
}

func (server *TCPServer) loop() {
	defer func() {
		server.listener.Close()
		server.deleteAll()
		server.slaveConn.Close()
	}()

	for {
		conn, err := server.listener.Accept()
		if err != nil {
			continue
		}

		log.Println("client:", conn.RemoteAddr(), conn.LocalAddr())

		// 转发消息
		if server.connCount >= server.maxConn {
			conn.Close()
			continue
		}

		id := server.put(conn)

		err = SendConnectMessage(server.slaveConn, id)
		if err != nil {
			return
		}

		go server.clientHandle(conn, id)
	}
}

func (server *TCPServer) clientHandle(conn net.Conn, id uint32) {
	defer func() {
		server.delete(id)
		SendDisconnectMessage(server.slaveConn, id)
	}()

	var err error
	buf := make([]byte, 0xffff-16)
	var n int

	for {
		n, err = conn.Read(buf)
		if err != nil {
			return
		}

		if n <= 0 {
			return
		}

		err = SendTransferMessage(server.slaveConn, id, buf[:n])
		if err != nil {
			return
		}
	}
}

func (server *TCPServer) slaveHandle() {
	defer func() {
		server.Stop()
	}()

	buf := make([]byte, 0xffff)
	head := make([]byte, 2)
	var messageLength uint16
	var messageType uint16

	var err error

	conn := server.slaveConn
	reader := bufio.NewReaderSize(conn, 0xffff)

	for {
		// 读取消息长度
		_, err = io.ReadFull(reader, head)
		if err != nil {
			return
		}
		messageLength = binary.BigEndian.Uint16(head)

		if _, err = io.ReadFull(reader, buf[:messageLength]); err != nil {
			return
		}
		messageType = binary.BigEndian.Uint16(buf[:2])

		log.Println("slave:", messageLength, messageType)

		switch messageType {
		case TransferID: // transfer
			{
				id := binary.BigEndian.Uint32(buf[2:])
				// log.Println("[REMOTE] transfer:", id, string(buf[6:messageLength]))

				if localConn := server.get(id); localConn != nil {
					if _, err = localConn.Write(buf[6:messageLength]); err != nil {
						return
					}
				}
			}
			break
		}
	}
}

func (server *TCPServer) Stop() {
	server.listener.Close()
	server.slaveConn.Close()
}

func NewTCPServer() *TCPServer {
	return new(TCPServer)
}
