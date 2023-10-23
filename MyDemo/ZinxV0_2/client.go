// @program:     ainx
// @file:        client.go
// @author:      ma
// @create:      2023-10-23 11:09
// @description:

package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("client start")

	time.Sleep(1 * time.Second)
	// 1.直接链接远程服务器，得到一个conn链接

	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("client start err , exit")
		return
	}

	for {
		// 链接调用write 写数据
		_, err := conn.Write([]byte("hello zinx V0.1"))
		if err != nil {
			fmt.Println("write conn err", err)
			return
		}
		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read buf error")
			return
		}

		fmt.Printf("server call back: %s = %d \n", buf, cnt)

		// cpu阻塞

		time.Sleep(2 * time.Second)

	}

}
