package main

import (
	"context"
	"fmt"
	"net"
)

type RequestSequence struct {
	requestId int
	conn net.Conn
	params map[string]string
	data []byte
	stdin []byte
}

func NewRequestSequence(conn net.Conn) RequestSequence {
	return RequestSequence{
		conn: conn,
		params: make(map[string]string),
	}
}

func (requestSequence RequestSequence) setRequestId(requestId int) RequestSequence {
	requestSequence.requestId = requestId

	return requestSequence
}

func (requestSequence RequestSequence) addParams(key string, value string) RequestSequence  {
	requestSequence.params[key] = value

	return requestSequence
}

func (requestSequence RequestSequence) setData(data []byte) RequestSequence  {
	requestSequence.data = make([]byte, len(data))
	requestSequence.data = data

	return requestSequence
}

func (requestSequence RequestSequence) setStdin(data []byte) RequestSequence  {
	requestSequence.stdin = make([]byte, len(data))
	requestSequence.stdin = data

	return requestSequence
}

func (requestSequence RequestSequence) String() string {
	var info string = fmt.Sprintf("---- request: %d ----\n", requestSequence.requestId)
	for key,value := range requestSequence.params {
		info += fmt.Sprintf("Param %s = %s \n", key, value)
	}
	info += fmt.Sprintf("Data %v\n", requestSequence.data)
	info += fmt.Sprintf("Stdin %v\n", requestSequence.stdin)

	return info
}

type FCGIStdoutHandler interface {
	Handle(sequence RequestSequence) error
}

type FCGIServer struct {
	listener net.Listener
	connChan chan net.Conn
	requestSequenceChan chan RequestSequence
	connHandlerCount int
	stdoutHandler FCGIStdoutHandler
	context context.Context
}

func NewFCGIServer(addr string, handler FCGIStdoutHandler) *FCGIServer {
	listener, error :=net.Listen("tcp", addr)
	if error != nil {
		panic(error)
	}

	return &FCGIServer{
		listener: listener,
		connChan: make(chan net.Conn),
		requestSequenceChan: make(chan RequestSequence),
		connHandlerCount: 5,
		stdoutHandler: handler,
		context: context.Background(),
	}
}

func (fcgiServer *FCGIServer) Serve() error {
	// 获取连接的requestSequence信息
	for workerIndex := 0; workerIndex < fcgiServer.connHandlerCount; workerIndex++ {
		go func() {
			for  {
				select {
				case conn := <- fcgiServer.GetConnConsumer():
					var fcgiHeader FCGI_Header
					var requestId int
					requestSequence := NewRequestSequence(conn)
					for {
						fcgiHeader,_ = ReadFcgiHeader(conn);
						var contentLength  = (int(fcgiHeader.contentLengthB1) << 8) + int(fcgiHeader.contentLengthB0)
						requestId = int((fcgiHeader.requestIdB1 << 8) + fcgiHeader.requestIdB0)

						if fcgiHeader.typ == FCGI_STDIN && contentLength == 0 {
							break
						}

						switch fcgiHeader.typ {
						case FCGI_BEGIN_REQUEST:
							beginRequestBody,_ := ReadFcgiBeginRequestBody(conn)
							fmt.Printf("BEGIN_REQUEST: role=%d,flags=%d \n",(int(beginRequestBody.roleB1)<<8)+int(beginRequestBody.roleB0), beginRequestBody.flags)
						case FCGI_PARAMS:
							for contentLength > 0 {
								var nameLength int
								var c = make([]byte, 1);
								_, error := conn.Read(c)
								if error != nil {
									panic(error)
								}
								contentLength -= 1;
								if (c[0] & 0x80) != 0 {
									var c3 = make([]byte, 3);
									_, error := conn.Read(c3)
									if error != nil {
										panic(error)
									}
									nameLength = (int(c[0]) << 24) + (int(c3[0]) << 16) + (int(c3[1]) << 8) + int(c3[2])
									contentLength -= 3
								} else {
									nameLength = int(c[0])
								}
								var valueLength int
								c = make([]byte, 1);
								_, error = conn.Read(c)
								if error != nil {
									panic(error)
								}
								contentLength -= 1
								if (c[0] & 0x80) != 0 {
									var c3 = make([]byte, 3);
									_, error := conn.Read(c3)
									if error != nil {
										panic(error)
									}
									valueLength = (int(c[0]) << 24) + (int(c3[0]) << 16) + (int(c3[1]) << 8) + int(c3[2])
									contentLength -= 3
								} else {
									valueLength = int(c[0])
								}

								var paramsNameBit = make([]byte, nameLength)
								_, error = conn.Read(paramsNameBit)
								if error != nil {
									panic(error)
								}
								contentLength -= nameLength

								var paramsValueBit = make([]byte, valueLength)
								_, error = conn.Read(paramsValueBit)
								if error != nil {
									panic(error)
								}
								contentLength -= valueLength
								requestSequence = requestSequence.addParams(string(paramsNameBit), string(paramsValueBit))
							}

							seekFcgiBodyPadding(conn, int(fcgiHeader.paddingLength))
						case FCGI_DATA:
							data, error := ReadFcgiData(conn, contentLength)
							if error != nil {
								panic(error)
							}
							requestSequence = requestSequence.setData(data)
							seekFcgiBodyPadding(conn, int(fcgiHeader.paddingLength))
						case FCGI_STDIN:
							data,error := ReadFcgiStdin(conn, contentLength)
							if error != nil {
								panic(error)
							}
							requestSequence = requestSequence.setStdin(data)
							seekFcgiBodyPadding(conn, int(fcgiHeader.paddingLength))
						case FCGI_ABORT_REQUEST:
							break
						default:

						}
					}
					// 无副作用值拷贝
					fcgiServer.AddRequestSequence(requestSequence.setRequestId(requestId))

				case <- fcgiServer.context.Done():
					break
				}

			}
		}()
	}

	// 处理handler worker可再抽象
	for workerIndex := 0; workerIndex < fcgiServer.connHandlerCount; workerIndex++ {
		go func() {
			for {
				select {
				case requestSequence := <-fcgiServer.GetRequestSequence():
					error := fcgiServer.stdoutHandler.Handle(requestSequence)
					if error != nil {
						panic(error)
					}
				case <-fcgiServer.context.Done():
					break
				}
			}
		}()
	}

	// 处理连接
	for {
		conn, error := fcgiServer.listener.Accept()
		if error != nil {
			panic(error)
		}
		fcgiServer.AddConnConsumer(conn)
	}
}

func (fcgiServer *FCGIServer) AddConnConsumer(conn net.Conn)  {
	fcgiServer.connChan <- conn
}

func (fcgiServer *FCGIServer) GetConnConsumer() <-chan net.Conn  {
	return fcgiServer.connChan
}

func (fcgiServer *FCGIServer) AddRequestSequence(sequence RequestSequence)  {
	fcgiServer.requestSequenceChan <- sequence
}

func (fcgiServer *FCGIServer) GetRequestSequence() <-chan RequestSequence{
	return fcgiServer.requestSequenceChan
}
