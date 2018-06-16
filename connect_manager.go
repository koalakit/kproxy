package main

import "net"

type ConnectManager struct {
	conns     []net.Conn
	maxConn   uint32
	connCount uint32
	idIndex   uint32
}

func (cm *ConnectManager) init() {
	cm.maxConn = 1024
	cm.connCount = 0
	cm.idIndex = 0
	cm.conns = make([]net.Conn, cm.maxConn)
}

func (server *ConnectManager) put(conn net.Conn) uint32 {
	var slot uint32

	for {
		server.idIndex++

		slot = server.idIndex % server.maxConn
		if server.conns[slot] == nil {
			break
		}
	}

	server.conns[slot] = conn

	return server.idIndex
}

func (server *ConnectManager) get(id uint32) net.Conn {
	return server.conns[id%server.maxConn]
}

func (server *ConnectManager) delete(id uint32) {
	slot := id % server.maxConn
	conn := server.conns[slot]
	server.conns[slot] = nil

	if conn != nil {
		conn.Close()
	}
}

func (server *ConnectManager) deleteAll() {
	for _, conn := range server.conns {
		if conn != nil {
			conn.Close()
		}
	}

	server.conns = nil
}
