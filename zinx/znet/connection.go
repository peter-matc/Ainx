// @program:     ainx
// @file:        connection.go
// @author:      ma
// @create:      2023-10-23 11:27
// @description:

package znet

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/peter-matc/Ainx/zinx/ziface"
)

// Connection
// @Description: 连接模块
type Connection struct {
	// 当前的Socket TCP套接字
	Conn *net.TCPConn
	// 链接的ID
	ConnID uint32
	// 当前的链接状态
	isClosed bool
	// 当前连接绑定的处理业务的API
	//HandleApi ziface.HandleFunc

	// 告知当前连接已经退出的channel

	ExitChan chan bool
	// 无缓冲管道 用于读写Goroutine之间的通信
	msgChan chan []byte
	// 该链接处理的方法Router
	//Router ziface.IRouter
	MsgHandler ziface.IMsgHandle
}

// StartReader
// @Description:
// @receiver c
func (c *Connection) StartReader() {
	fmt.Println("[Reader Goroutine is running...]")
	defer fmt.Println("connID = ", c.ConnID, " [Reader is exit],remote addr is ", c.GetRemoteAddr().String())
	defer c.Stop()
	// 当前的处理业务
	for {

		dp := NewDataPack()

		// 获取客户端Msg Head 二进制流 8个字节
		headData := make([]byte, dp.GetHeadLen())
		_, err := io.ReadFull(c.GetTCPConnection(), headData)
		if err != nil {
			fmt.Println("read msg head error", err)
			break
		}
		// 将数据发送给客户端
		// 拆包 得到msgID 和 msgDataLen 放到对象中
		msg, err := dp.Unpack(headData)

		if err != nil {
			fmt.Println("unpack error", err)
			break
		}
		// 根据dataLen 再次读取data
		var data []byte
		if msg.GetMsgLen() > 0 {
			data = make([]byte, msg.GetMsgLen())
			_, err := io.ReadFull(c.GetTCPConnection(), data)
			if err != nil {
				fmt.Println("read msg data error ", err)
				break
			}
		}
		msg.SetData(data)
		//// 调用当前连接绑定的HandleAPI
		//if err = c.HandleApi(c.Conn, buf, cnt); err != nil {
		//	fmt.Println("ConnID ", c.ConnID, " handle is error ", err)
		//	break
		//}

		// 得到当前conn数据的Request请求数据

		req := Request{
			conn: c,
			msg:  msg,
		}

		// 从路由中，找到注册绑定的Conn对应的router调用
		// 执行注册路由方法
		go c.MsgHandler.DoMsgHandler(&req)

	}

}

// StartWriter
// @Description: 写消息，专门发送给客户端的模块
// @receiver c
func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running...]")
	defer fmt.Println("[conn Writer exit]", c.GetRemoteAddr().String())
	// 不断的阻塞的等待channel的消息，写给客户端
	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send data error ", err)
				return
			}
		case <-c.ExitChan:
			// 代表Reader已经退出，此时Writer也要退出
			return
		}
	}
}

func (c *Connection) Start() {
	fmt.Println("Conn Start() ... ConnID = ", c.ConnID)
	// 启动从当前连接的读数据的业务
	go c.StartReader()

	// 启动从当前连接写数据的业务
	go c.StartWriter()
}

func (c *Connection) Stop() {

	fmt.Println("Conn Stop().. ConnID= ", c.ConnID)

	if c.isClosed == true {
		return
	}

	c.isClosed = true
	// 关闭socket 连接

	err := c.Conn.Close()
	if err != nil {
		return
	}

	c.ExitChan <- true

	close(c.msgChan)
	//回收资源
	close(c.ExitChan)

}

func (c *Connection) GetTCPConnection() *net.TCPConn {

	return c.Conn
}

func (c *Connection) GetConnID() uint32 {

	return c.ConnID
}

func (c *Connection) GetRemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// SendMsg
//
//	@Description: 封包先进行封包 再发送
//	@receiver c
//	@param msgId
//	@param data
//	@return error
func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("connection closed when send msg")
	}
	// 将data 进行封包
	dp := NewDataPack()
	// msgDataLen ｜ msgID ｜ data
	binaryMsg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("pack error msg")
	}
	// 将数据发送给客户端
	//_, err = c.Conn.Write(binaryMsg)
	//if err != nil {
	//	fmt.Println("Write msg id ", msgId, " error ", err)
	//	return errors.New("conn Write error")
	//}
	// 发送给channel
	c.msgChan <- binaryMsg

	return nil
}

// NewConnection
// @Description: 初始化连接模块的方法
// @param conn
// @param connID
// @param router
// @return *Connection
func NewConnection(conn *net.TCPConn, connID uint32, MsgHandle ziface.IMsgHandle) *Connection {
	c := &Connection{
		Conn:     conn,
		ConnID:   connID,
		isClosed: false, //开启状态
		//HandleApi: callback_api,
		ExitChan:   make(chan bool, 1),
		msgChan:    make(chan []byte),
		MsgHandler: MsgHandle,
	}
	return c
}
