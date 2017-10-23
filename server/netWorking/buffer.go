package netWorking

import (
	"Clans/server/log"
	"net"
)

// PIPELINE #3: buffer
// controls the packet sending for the client
type Buffer struct {
	ctrl    chan struct{} // receive exit signal
	pending chan []byte   // pending packets
	conn    net.Conn      // connection
	// cache   []byte        // for combined syscall write
}

// packet sending goroutine
func (buf *Buffer) Start() {
	for {
		select {
		case data := <-buf.pending:
			buf.RawSend(data)
		case <-buf.ctrl: // receive session end signal
			return
		}
	}
}

// raw packet encapsulation and put it online
func (buf *Buffer) RawSend(data []byte) bool {
	// // combine output to reduce syscall.write
	// sz := len(data)
	// binary.BigEndian.PutUint16(buf.cache, uint16(sz))
	// copy(buf.cache[2:], data)

	// write data
	n, err := buf.conn.Write(data)
	if err != nil {
		log.Logger().Warnf("Error send reply data, bytes: %v reason: %v", n, err)
		return false
	}

	return true
}

// create a associated write buffer for a session
func NewBuffer(conn net.Conn, ctrl chan struct{}, txqueuelen int) *Buffer {
	buf := Buffer{conn: conn}
	buf.pending = make(chan []byte, txqueuelen)
	buf.ctrl = ctrl
	// buf.cache = make([]byte, netPackages.PACKET_LIMIT+2)
	return &buf
}
