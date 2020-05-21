package main

import (
	"bytes"
	"fmt"
)

type StandardStdoutHandler struct {

}

func (handler StandardStdoutHandler) Handle(sequence RequestSequence) error  {

	fmt.Println(sequence)

	// 发送输出内容
	htmlHead := "Content-type: text/html\r\n\r\n";  //响应头
	htmlBody := "HELLO";  // 把请求文件路径作为响应体返回

	bytes := bytes.NewBufferString(htmlHead)
	bytes.WriteString(htmlBody)

	l, error := WriteFcgStdout(sequence.conn, bytes.Bytes(), sequence.requestId)
	if error != nil {
		panic(error)
	}
	fmt.Printf("write stdout %d\n", l)

	// 发送输出结束消息
	_, error = WriteFcgStdout(sequence.conn, []byte{}, sequence.requestId)
	if error != nil {
		panic(error)
	}

	// 发送请求结束消息
	_, error = WriteFcgiEndRequest(sequence.conn, sequence.requestId)
	if error != nil {
		panic(error)
	}

	sequence.conn.Close()

	return nil
}