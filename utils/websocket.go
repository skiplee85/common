package utils

import (
	"errors"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketConn 封装链接
type WebSocketConn struct {
	wsConnect *websocket.Conn
	inChan    chan []byte
	outChan   chan message
	closeChan chan byte

	mutex    sync.Mutex // 对closeChan关闭上锁
	IsClosed bool       // 防止closeChan被关闭多次
}

type message struct {
	messageType int
	data        []byte
}

// InitWebSocketConn 初始化
func InitWebSocketConn(wsConn *websocket.Conn) (conn *WebSocketConn, err error) {
	conn = &WebSocketConn{
		wsConnect: wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan message, 1000),
		closeChan: make(chan byte, 1),
	}
	// 启动读协程
	go conn.readLoop()
	// 启动写协程
	go conn.writeLoop()
	return
}

// ReadMessage 读取信息
func (conn *WebSocketConn) ReadMessage() (data []byte, err error) {

	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closeed")
	}
	return
}

// WriteMessage 下发信息
func (conn *WebSocketConn) WriteMessage(mt int, data []byte) (err error) {

	select {
	case conn.outChan <- message{mt, data}:
	case <-conn.closeChan:
		err = errors.New("connection is closeed")
	}
	return
}

// Close 关闭
func (conn *WebSocketConn) Close() {
	// 线程安全，可多次调用
	conn.wsConnect.Close()
	// 利用标记，让closeChan只关闭一次
	conn.mutex.Lock()
	if !conn.IsClosed {
		close(conn.closeChan)
		conn.IsClosed = true
	}
	conn.mutex.Unlock()
}

// 内部实现
func (conn *WebSocketConn) readLoop() {
	var (
		data []byte
		err  error
	)
	for {
		if _, data, err = conn.wsConnect.ReadMessage(); err != nil {
			goto ERR
		}
		//阻塞在这里，等待inChan有空闲位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan: // closeChan 感知 conn断开
			goto ERR
		}

	}

ERR:
	conn.Close()
}

func (conn *WebSocketConn) writeLoop() {
	var (
		m   message
		err error
	)

	for {
		select {
		case m = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}
		if err = conn.wsConnect.WriteMessage(m.messageType, m.data); err != nil {
			goto ERR
		}
	}

ERR:
	conn.Close()

}
