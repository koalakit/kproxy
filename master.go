package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
)

type ProxyMaster struct {
	listener net.Listener
}

func (pm *ProxyMaster) Start(addr string) {
	var err error
	pm.listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("ProxyMaster:", err)
	}

	log.Println("[PROXY]", addr)

	go pm.loop()
}

func (pm *ProxyMaster) loop() {
	buf := make([]byte, 2)

	for {
		conn, err := pm.listener.Accept()
		if err != nil {
			continue
		}

		log.Println("Slave:", conn.RemoteAddr(), conn.LocalAddr())

		// 读取端口
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			continue
		}
		port := binary.BigEndian.Uint16(buf)

		server := new(TCPServer)
		server.slaveConn = conn
		server.Start(fmt.Sprintf(":%d", port))
	}
}

func NewProxyMaster() *ProxyMaster {
	return new(ProxyMaster)
}
