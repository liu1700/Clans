package services

import "time"

const (
	GAME_SERVICE = "game"
	CHAT_SERVICE = "chat"
)

type Config struct {
	Listen                        string
	ReadDeadline                  time.Duration
	Sockbuf                       int
	Udp_sockbuf                   int
	Txqueuelen                    int
	Dscp                          int
	Sndwnd                        int
	Rcvwnd                        int
	Mtu                           int
	Nodelay, Interval, Resend, Nc int
	RpmLimit                      int
}
