package agent_server

import (
	"Clans/server/log"
	"Clans/server/netPackages"
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

// packet sending procedure
func (buf *Buffer) send(sess *Session, respId int, srcPack *netPackages.NetPackage, sendData []byte) {

	// encryption
	// (NOT_ENCRYPTED) -> KEYEXCG -> ENCRYPT
	if sess.Flag&SESS_ENCRYPT != 0 { // encryption is enabled
		sess.Encoder.XORKeyStream(sendData, sendData)
	} else if sess.Flag&SESS_KEYEXCG != 0 { // key is exchanged, encryption is not yet enabled
		sess.Flag &^= SESS_KEYEXCG
		sess.Flag |= SESS_ENCRYPT
	}

	// 更新数据包的数据内容
	srcPack.HandlerId = byte(respId)
	srcPack.Version = version
	srcPack.Data = sendData

	// queue the data for sending
	select {
	case buf.pending <- srcPack.Bytes():
	default: // packet will be dropped if txqueuelen exceeds
		log.Logger().Warnf("userid %d ip %s", sess.UserId, sess.IP)
	}
	return
}

// packet sending goroutine
func (buf *Buffer) Start() {
	for {
		select {
		case data := <-buf.pending:
			buf.rawSend(data)
		case <-buf.ctrl: // receive session end signal
			return
		}
	}
}

// raw packet encapsulation and put it online
func (buf *Buffer) rawSend(data []byte) bool {
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
