package httpobfs

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/nadoo/glider/pkg/pool"
	"github.com/nadoo/glider/proxy"
)

var (
	bufSize          = proxy.TCPBufSize
	requestUserAgent = []string{
		"Mozilla/5.0 (Windows NT 6.3; WOW64; rv:40.0) Gecko/20100101 Firefox/40.0",
		"Mozilla/5.0 (Windows NT 6.3; WOW64; rv:40.0) Gecko/20100101 Firefox/44.0",
		"Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/535.11 (KHTML, like Gecko) Ubuntu/11.10 Chromium/27.0.1453.93 Chrome/27.0.1453.93 Safari/537.36",
		"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:35.0) Gecko/20100101 Firefox/35.0",
		"Mozilla/5.0 (compatible; WOW64; MSIE 10.0; Windows NT 6.2)",
		"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/533.20.25 (KHTML, like Gecko) Version/5.0.4 Safari/533.20.27",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.3; Trident/7.0; .NET4.0E; .NET4.0C)",
		"Mozilla/5.0 (Windows NT 6.3; Trident/7.0; rv:11.0) like Gecko",
		"Mozilla/5.0 (Linux; Android 4.4; Nexus 5 Build/BuildID) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/30.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (iPad; CPU OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 5_0 like Mac OS X) AppleWebKit/534.46 (KHTML, like Gecko) Version/5.1 Mobile/9A334 Safari/7534.48.3",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type HttpObfsConn struct {
	net.Conn
	rawTransSent     bool
	rawTransReceived bool
	userAgentIndex   int
	readBuf          []byte
	extraReadBuf     *bytes.Buffer

	Path string
	Host string
}

func (t *HttpObfsConn) Encode(data []byte) (encodedData []byte, err error) {
	if t.rawTransSent {
		return data, nil
	}

	httpBuf := fmt.Sprintf(
		"GET %s HTTP/1.1\r\nAccept-Encoding: gzip, deflate\r\nConnection: keep-alive\r\nHost: %s\r\nPragma: no-cache\r\nUser-Agent: %s\r\n\r\n",
		t.Path,
		t.Host,
		requestUserAgent[t.userAgentIndex],
	)

	encodedData = make([]byte, len(httpBuf)+len(data))
	copy(encodedData, []byte(httpBuf))
	copy(encodedData[len(httpBuf):], data)

	t.rawTransSent = true

	return
}

func (t *HttpObfsConn) Decode(data []byte) (decodedData []byte, err error) {
	if t.rawTransReceived {
		return data, nil
	}

	pos := bytes.Index(data, []byte("\r\n\r\n"))
	if pos > 0 {
		decodedData = make([]byte, len(data)-pos-4)
		copy(decodedData, data[pos+4:])
		t.rawTransReceived = true
	}

	return decodedData, nil
}

func NewHttpObfsConn(c net.Conn, path string, host string) *HttpObfsConn {
	return &HttpObfsConn{
		Conn:             c,
		readBuf:          pool.GetBuffer(bufSize),
		extraReadBuf:     pool.GetBytesBuffer(),
		rawTransSent:     false,
		rawTransReceived: false,
		userAgentIndex:   rand.Intn(len(requestUserAgent)),
		Path:             path,
		Host:             host,
	}
}

func (c *HttpObfsConn) Close() error {
	pool.PutBuffer(c.readBuf)
	pool.PutBytesBuffer(c.extraReadBuf)
	return c.Conn.Close()
}

func (c *HttpObfsConn) Read(b []byte) (n int, err error) {
	for {
		n, err = c.doRead(b)
		if b == nil || n != 0 || err != nil {
			return n, err
		}
	}
}

func (c *HttpObfsConn) doRead(b []byte) (n int, err error) {
	if c.extraReadBuf.Len() > 0 {
		return c.extraReadBuf.Read(b)
	}

	n, err = c.Conn.Read(c.readBuf)
	if n == 0 || err != nil {
		return n, err
	}

	decodedData, err := c.Decode(c.readBuf[:n])
	if err != nil {
		return 0, err
	}

	decodedDataLen := len(decodedData)
	if decodedDataLen == 0 {
		return 0, nil
	}

	decodedDataLength := len(decodedData)
	blength := len(b)

	if blength >= decodedDataLength {
		copy(b, decodedData)
		return decodedDataLength, nil
	}

	copy(b, decodedData[:blength])
	c.extraReadBuf.Write(decodedData[blength:])

	return blength, nil
}

func (c *HttpObfsConn) preWrite(b []byte) (outData []byte, err error) {
	if b == nil {
		b = make([]byte, 0)
	}

	return c.Encode(b)
}

func (c *HttpObfsConn) Write(b []byte) (n int, err error) {
	outData, err := c.preWrite(b)
	if err != nil {
		return 0, err
	}
	n, err = c.Conn.Write(outData)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}
