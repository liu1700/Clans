package netWorking

import (
	"Clans/server/flats"
	"Clans/server/log"
	"bytes"
	"encoding/binary"
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

// packet sending procedure
func (buf *Buffer) Send(sess *Session, data []byte) {
	// in case of empty packet
	if data == nil {
		return
	}

	// encryption
	// (NOT_ENCRYPTED) -> KEYEXCG -> ENCRYPT
	if sess.Flag&SESS_ENCRYPT != 0 { // encryption is enabled
		sess.Encoder.XORKeyStream(data, data)
	} else if sess.Flag&SESS_KEYEXCG != 0 { // key is exchanged, encryption is not yet enabled
		sess.Flag &^= SESS_KEYEXCG
		sess.Flag |= SESS_ENCRYPT
	}

	// queue the data for sending
	select {
	case buf.pending <- data:
	default: // packet will be dropped if txqueuelen exceeds
		log.Logger().Warn("pending full")
	}
	return
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

func (buf *Buffer) SendFrame(data []byte) bool {
	// 写入packetid 供前端解析
	d := new(bytes.Buffer)
	binary.Write(d, binary.BigEndian, byte(flats.PacketIdGame))
	// 写入数据长度
	binary.Write(d, binary.BigEndian, uint16(len(data)))
	// binary.Write(d, binary.BigEndian, data)
	d.Write(data)

	n, err := buf.conn.Write(d.Bytes())
	if err != nil {
		log.Logger().Warnf("Error send frame, bytes: %v reason: %v", n, err)
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
