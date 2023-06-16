package monitor

import (
	"net/http"
	"net/url"

	"github.com/fasthttp/websocket"
)

func GetConn(wsurl string, proxy ...string) (*websocket.Conn, error) {
	u, err := url.Parse(wsurl)
	if err != nil {
		return nil, err
	}

	headers := http.Header{
		"Sec-WebSocket-Protocol": {"uproxy-hub-v1"},
	}

	if len(proxy) == 0 || proxy[0] == "" {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), headers)
		if err != nil {
			return conn, err
		}
		return conn, nil
	} else {
		pU, err := url.Parse(proxy[0])
		if err != nil {
			return nil, err
		}
		dialer := &websocket.Dialer{
			Proxy: http.ProxyURL(pU),
		}
		conn, _, err := dialer.Dial(u.String(), headers)
		return conn, err
	}

}
