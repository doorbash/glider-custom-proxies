package doh

import (
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nadoo/glider/pkg/log"
	"github.com/nadoo/glider/proxy"
)

func init() {
	proxy.RegisterDialer("doh", NewDohDialer)
}

type Doh struct {
	dialer proxy.Dialer
	addr   string
	// path   string
	client *http.Client
}

func NewDoh(s string, d proxy.Dialer) (*Doh, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.F("[doh] parse err: %s", err)
		return nil, err
	}

	query := u.Query()

	timeout, _ := strconv.ParseInt(query.Get("timeout"), 10, 0)

	p := &Doh{
		dialer: d,
		addr:   u.Host,
		// path:   u.Path,
		client: &http.Client{
			Transport: &http.Transport{
				Dial: func(network string, addr string) (net.Conn, error) {
					rc, err := d.Dial("tcp", addr)
					if err != nil {
						return nil, err
					}

					return rc, nil
				},
			},
			Timeout: time.Duration(timeout) * time.Second,
		},
	}

	return p, nil
}

func NewDohDialer(s string, d proxy.Dialer) (proxy.Dialer, error) {
	return NewDoh(s, d)
}

func (s *Doh) Addr() string {
	if s.addr == "" {
		return s.dialer.Addr()
	}
	return s.addr
}

func (s *Doh) Dial(network, addr string) (c net.Conn, err error) {
	return s.dialer.Dial(network, addr)
}

func (s *Doh) DialUDP(network, addr string) (pc net.PacketConn, err error) {
	if strings.HasSuffix(addr, ":53") {
		return &DohPacketConn{
			d:  s,
			ch: make(chan []byte, 1),
		}, nil
	}
	return s.dialer.DialUDP(network, addr)
}
