package httpobfs

import (
	"errors"
	"net"
	"net/url"

	"github.com/nadoo/glider/pkg/log"
	"github.com/nadoo/glider/proxy"
)

func init() {
	proxy.RegisterDialer("httpobfs", NewHttpObfsDialer)
}

type HttpObfs struct {
	dialer proxy.Dialer
	addr   string
	path   string
	host   string
}

func NewHttpObfs(s string, d proxy.Dialer) (*HttpObfs, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.F("[httpobfs] parse err: %s", err)
		return nil, err
	}

	query := u.Query()

	addr := u.Host
	host := query.Get("host")

	p := &HttpObfs{
		dialer: d,
		addr:   addr,
		path:   u.Path,
		host:   host,
	}

	return p, nil
}

func NewHttpObfsDialer(s string, d proxy.Dialer) (proxy.Dialer, error) {
	return NewHttpObfs(s, d)
}

// Addr returns forwarder's address
func (s *HttpObfs) Addr() string {
	if s.addr == "" {
		return s.dialer.Addr()
	}
	return s.addr
}

// Dial connects to the address addr on the network net via the proxy.
func (s *HttpObfs) Dial(network, addr string) (net.Conn, error) {
	c, err := s.dialer.Dial("tcp", s.addr)
	if err != nil {
		log.F("[http_simple] dial to %s error: %s", s.addr, err)
		return nil, err
	}

	conn := NewHttpObfsConn(c, s.path, s.host)
	if conn.Conn == nil || conn.RemoteAddr() == nil {
		return nil, errors.New("[http_simple] nil connection")
	}

	return conn, err
}

func (s *HttpObfs) DialUDP(network, addr string) (net.PacketConn, error) {
	return nil, proxy.ErrNotSupported
}

func init() {
	proxy.AddUsage("httpobfs", `
httpobfs scheme:
  httpobfs://host:port?host=xxx
`)
}
