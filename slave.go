package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ProxySlave struct {
	ConnectManager
	conn net.Conn

	proxyAddr    string
	remote2Local map[uint32]uint32
}

func (slave *ProxySlave) Start(addr string, port uint16, proxyAddr string) {
	slave.init()

	slave.remote2Local = make(map[uint32]uint32)

	slave.proxyAddr = proxyAddr

	var err error
	slave.conn, err = net.Dial("tcp", addr)
	if err != nil {
		log.Fatalln("ProxySlave:", err)
	}

	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, port)
	_, err = slave.conn.Write(buf)
	if err != nil {
		log.Fatalln("ProxySlave:", err)
	}

	slave.init()

	// go startHTTP()

	go slave.loop()
}

func (slave *ProxySlave) loop() {
	defer func() {

	}()

	conn := slave.conn
	var err error
	buf := make([]byte, 0xffff)
	uint16Buffer := make([]byte, 2)

	var messageLength uint16
	var messageType uint16

	for {
		if _, err = io.ReadFull(conn, uint16Buffer); err != nil {
			return
		}
		// log.Printf("[SLAVE] %x", uint16Buffer)
		messageLength = binary.BigEndian.Uint16(uint16Buffer)

		if _, err = io.ReadFull(conn, buf[:messageLength]); err != nil {
			return
		}
		messageType = binary.BigEndian.Uint16(buf[:2])

		switch messageType {
		case PingID:
			// err = SendPingMessage(conn)
			break
		case ConnectID:
			{
				id := binary.BigEndian.Uint32(buf[2:])
				log.Println("[REMOTE] connect:", id, *localAddr)

				localConn, err := net.Dial("tcp", *localAddr)
				if err != nil {
					if err = SendDisconnectMessage(conn, id); err != nil {
						return
					}
				}

				slave.remote2Local[id] = slave.put(localConn)
				go slave.transfor(localConn, id)
			}
			break
		case DisconnectID:
			{
				id := binary.BigEndian.Uint32(buf[2:])
				log.Println("[REMOTE] disconnect:", id)

				localID := slave.remote2Local[id]
				slave.delete(localID)
			}
			break
		case TransferID:
			{
				id := binary.BigEndian.Uint32(buf[2:])
				// log.Println("[REMOTE] transfer:", id, string(buf[6:messageLength]))

				localID := slave.remote2Local[id]
				if err = slave.write(localID, buf[6:messageLength]); err != nil {
					slave.delete(localID)
					if err = SendDisconnectMessage(conn, id); err != nil {
						return
					}
				}
			}
			break
		}

		if err != nil {
			conn.Close()
			return
		}
	}
}

func (slave *ProxySlave) write(id uint32, request []byte) error {
	localConn := slave.get(id)
	if localConn == nil {
		log.Println("localConn is nil")
		return fmt.Errorf("not found conn ", id)
	}

	var err error
	// l := len(request)
	// n := 0
	// sendBytes := 0

	// for n < l {
	// 	sendBytes, err = localConn.Write(request[n:])
	// 	if err != nil {
	// 		return err
	// 	}

	// 	n += sendBytes
	// 	log.Println("slave write", sendBytes)
	// }

	_, err = localConn.Write(request)
	if err != nil {
		return err
	}

	return nil
}

func (slave *ProxySlave) transfor(conn net.Conn, id uint32) {
	log.Println("transfor", id)

	buf := make([]byte, 0xffff-32)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		// log.Println("recv: %s", string(buf[:n]))

		if err = SendTransferMessage(slave.conn, id, buf[:n]); err != nil {
			return
		}
	}
}

type handle struct {
}

func (h *handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	remote, err := url.Parse("http://localhost:80")
	if err != nil {
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ServeHTTP(w, r)
}

func startHTTP() {
	h := new(handle)
	http.ListenAndServe(":6060", h)
}

func NewProxySlave() *ProxySlave {
	return new(ProxySlave)
}
