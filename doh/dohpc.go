package doh

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

type DohPacketConn struct {
	d  *Doh
	ch chan []byte
}

func (c *DohPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	msg := <-c.ch
	copy(p, msg)
	return len(msg), nil, nil
}

func (f *DohPacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	enc := base64.RawURLEncoding.EncodeToString(p)
	url := fmt.Sprintf("https://%s%s?dns=%s", f.d.addr, f.d.path, enc)
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("could not create request: %s", err)
	}
	r.Header.Set("Content-Type", "application/dns-message")
	r.Header.Set("Accept", "application/dns-message")

	resp, err := f.d.client.Do(r)
	if err != nil {
		return 0, fmt.Errorf("could not perform request: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("wrong response from DOH server got %s", http.StatusText(resp.StatusCode))
	}

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("could not read message from response: %s", err)
	}

	f.ch <- msg

	return len(p), nil
}

func (f *DohPacketConn) Close() error {
	close(f.ch)
	return nil
}

func (f *DohPacketConn) SetDeadline(t time.Time) error {
	return nil
}

func (f *DohPacketConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (f *DohPacketConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func (f *DohPacketConn) LocalAddr() net.Addr {
	return nil
}
