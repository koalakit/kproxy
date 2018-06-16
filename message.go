package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
)

const (
	// PingID ping消息
	PingID uint16 = iota
	// ConnectID 建立连接消息
	ConnectID
	// DisconnectID 断开连接消息
	DisconnectID
	// TransferID 转发消息
	TransferID
)

// SendPingMessage 发送ping消息
func SendPingMessage(conn net.Conn) error {
	var err error

	cache := make([]byte, 0, 4)
	buf := bytes.NewBuffer(cache)

	binary.Write(buf, binary.BigEndian, uint16(2))
	binary.Write(buf, binary.BigEndian, PingID)

	if _, err = conn.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

// SendConnectMessage 发送连接建立消息
func SendConnectMessage(conn net.Conn, id uint32) error {
	var err error
	cache := make([]byte, 0, 8)
	buf := bytes.NewBuffer(cache)
	binary.Write(buf, binary.BigEndian, uint16(6))
	binary.Write(buf, binary.BigEndian, ConnectID)
	binary.Write(buf, binary.BigEndian, id)

	// log.Printf("SendConnectMessage: %x", buf.Bytes())
	log.Println("SendConnectMessage", buf.Len())

	if _, err = conn.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

// SendDisconnectMessage 发送断开连接消息
func SendDisconnectMessage(conn net.Conn, id uint32) error {
	var err error
	cache := make([]byte, 0, 8)
	buf := bytes.NewBuffer(cache)

	binary.Write(buf, binary.BigEndian, uint16(6))
	binary.Write(buf, binary.BigEndian, DisconnectID)
	binary.Write(buf, binary.BigEndian, id)

	log.Println("SendDisconnectMessage", buf.Len())
	if _, err = conn.Write(buf.Bytes()); err != nil {
		return err
	}

	return nil
}

// SendTransferMessage 发送转发消息
func SendTransferMessage(conn net.Conn, id uint32, data []byte) error {
	var err error

	head := make([]byte, 0, 8)
	buf := bytes.NewBuffer(head)

	binary.Write(buf, binary.BigEndian, uint16(6+len(data)))
	binary.Write(buf, binary.BigEndian, TransferID)
	binary.Write(buf, binary.BigEndian, id)

	log.Println("SendTransferMessage", uint16(6+len(data)), id)
	if _, err = conn.Write(buf.Bytes()); err != nil {
		return err
	}

	var n int
	if n, err = conn.Write(data); err != nil {
		return err
	}

	if n != len(data) {
		log.Println("SendTransferMessage error:", n, len(data))
	}

	return nil
}
