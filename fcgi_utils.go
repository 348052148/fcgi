package main

import (
	"errors"
	"net"
	"unsafe"
)

// fastCgi类型
const FCGI_BEGIN_REQUEST = 1
const FCGI_ABORT_REQUEST = 2
const FCGI_END_REQUEST = 3
const FCGI_PARAMS = 4
const FCGI_STDIN = 5
const FCGI_STDOUT = 6
const FCGI_STDERR = 7
const FCGI_DATA = 8
// 管理记录类型
const FCGI_GET_VALUES = 9
const FCGI_GET_VALUES_RESULT = 10
const FCGI_UNKNOWN_TYPE = 11
const FCGI_MAXTYPE = FCGI_UNKNOWN_TYPE

// fastCgi角色
const FCGI_RESPONDER = 1
const FCGI_AUTHORIZER = 2
const FCGI_FILTER = 3

// 请求头信息
type FCGI_Header struct {
	version         byte
	typ             byte
	requestIdB1     byte
	requestIdB0     byte
	contentLengthB1 byte
	contentLengthB0 byte
	paddingLength   byte
	reserved        byte
}

// FCGI_BEGIN_REQUEST Body
type FCGI_BeginRequestBody struct {
	roleB1   byte
	roleB0   byte
	flags    byte
	reserved [5]byte
}

// FCGI_BEGIN_REQUEST Record
type FCGI_BeginRequestRecord struct {
	header FCGI_Header
	body   FCGI_BeginRequestBody
}

// FCGI_PARAMS  11
type FCGI_NameValuePair11 struct {
	nameLengthB0  byte
	valueLengthB0 byte
	nameData      []byte // 名称数据
	valueData     []byte // 值数据
}

// FCGI_PARAMS 14
type FCGI_NameValuePair14 struct {
	nameLengthB0  byte
	valueLengthB3 byte
	valueLengthB2 byte
	valueLengthB1 byte
	valueLengthB0 byte
	nameData      []byte // 名称数据
	valueData     []byte // 值数据
}

// FCGI_PARAMS 41
type FCGI_NameValuePair41 struct {
	nameLengthB3  byte
	nameLengthB2  byte
	nameLengthB1  byte
	nameLengthB0  byte
	valueLengthB0 byte
	nameData      []byte // 名称数据
	valueData     []byte // 值数据
}

// FCGI_PARAMS 44
type FCGI_NameValuePair44 struct {
	nameLengthB3  byte
	nameLengthB2  byte
	nameLengthB1  byte
	nameLengthB0  byte
	valueLengthB3 byte
	valueLengthB2 byte
	valueLengthB1 byte
	valueLengthB0 byte
	nameData      []byte // 名称数据
	valueData     []byte // 值数据
}

// Values for protocolStatus component of FCGI_EndRequestBody
const FCGI_REQUEST_COMPLETE = 0
const FCGI_CANT_MPX_CONN = 1
const FCGI_OVERLOADED = 2
const FCGI_UNKNOWN_ROLE = 3

//FCGI_END_REQUEST Body
type FCGI_EndRequestBody struct {
	appStatusB3    byte
	appStatusB2    byte
	appStatusB1    byte
	appStatusB0    byte
	protocolStatus byte
	reserved       [3]byte
}

//FCGI_END_REQUEST Record
type FCGI_EndRequestRecord struct {
	header FCGI_Header
	body   FCGI_EndRequestBody
}

/*
 * Variable names for FCGI_GET_VALUES / FCGI_GET_VALUES_RESULT records
 */
const FCGI_MAX_CONNS = "FCGI_MAX_CONNS"
const FCGI_MAX_REQS = "FCGI_MAX_REQS"
const FCGI_MPXS_CONNS = "FCGI_MPXS_CONNS"

//
type FCGI_UnknownTypeBody struct {
	typ      byte
	reserved [7]byte
}

type FCGI_UnknownTypeRecord struct {
	header FCGI_Header
	body   FCGI_UnknownTypeBody
}

// Struct Transform Bits
type SliceMock struct {
	addr uintptr
	len  int
	cap  int
}

func TransformBits(addr uintptr, length int) []byte {
	sliceMockTest := SliceMock{
		addr: addr,
		len:  length,
		cap:  length,
	}
	bits := *(*[]byte)(unsafe.Pointer(&sliceMockTest))

	return bits
}

// Actions

// ReadFcgiHeader
func ReadFcgiHeader(conn net.Conn) (FCGI_Header, error) {
	var length uintptr = unsafe.Sizeof(FCGI_Header{})
	var bit = make([]byte, int(length))
	_, error := conn.Read(bit)
	if error != nil {
		return FCGI_Header{}, error
	}
	fcgiHeader := *(*FCGI_Header)(unsafe.Pointer(&bit[0]))

	return fcgiHeader, nil
}

// FCGI_BEGIN_REQUEST
func ReadFcgiBeginRequestBody(conn net.Conn) (FCGI_BeginRequestBody, error) {
	var length uintptr = unsafe.Sizeof(FCGI_BeginRequestBody{})
	var bit = make([]byte, int(length))
	_, error := conn.Read(bit)
	if error != nil {
		return FCGI_BeginRequestBody{}, error
	}

	fcgiHeader := *(*FCGI_BeginRequestBody)(unsafe.Pointer(&bit[0]))

	return fcgiHeader, nil
}

func ReadFcgiParams() (interface{}, error) {
	return nil, errors.New("")
}

func ReadFcgiParamsPair11() (FCGI_NameValuePair11, error) {

	return FCGI_NameValuePair11{}, errors.New("")
}

// FCGI_STDIN
func ReadFcgiStdin(conn net.Conn, length int) ([]byte, error) {
	var bitLength = 4096
	var error error
	var data []byte
	for length > 0 {
		var bitss = make([]byte, bitLength)
		if length > bitLength {
			_, error = conn.Read(bitss)
			length -= bitLength
		} else {
			_, error = conn.Read(bitss)
			length -= length
		}
		if error != nil {
			return nil, error
		}
		data = append(data, bitss...)
	}
	return data, nil
}

// FCGI_DATA
func ReadFcgiData(conn net.Conn, length int) ([]byte, error) {
	var bitLength = 4096
	var error error
	var data []byte
	for length > 0 {
		var bitss = make([]byte, bitLength)
		if length > bitLength {
			_, error = conn.Read(bitss)
			length -= bitLength
		} else {
			_, error = conn.Read(bitss)
			length -= length
		}
		if error != nil {
			return nil, error
		}
		data = append(data, bitss...)
	}
	return data, nil
}

func seekFcgiBodyPadding(conn net.Conn, paddingLength int) error {
	if paddingLength > 0 {
		var padding = make([]byte, paddingLength)
		_, error := conn.Read(padding)
		if error != nil {
			return error
		}
	}
	return nil
}

// FCGI_STDOUT
func WriteFcgStdout(conn net.Conn, body []byte, requestId int) (int, error) {
	length := len(body)
	header := FCGI_Header{
		version:         1,
		typ:             FCGI_STDOUT,
		requestIdB0:     byte(requestId & 0xff),
		requestIdB1:     byte(requestId>>8) & 0xff,
		contentLengthB0: byte(length & 0xff),
		contentLengthB1: byte((length >> 8) & 0xff),
	}
	if (length % 8) > 0 {
		header.paddingLength = byte(8 - (length % 8))
	} else {
		header.paddingLength = 0
	}

	var headerLength = int(unsafe.Sizeof(header))
	bits := TransformBits(uintptr(unsafe.Pointer(&header)), headerLength)
	_, error := conn.Write(bits)
	if error != nil {
		panic(error)
	}

	// 如果有数据
	if length > 0 {
		_, error = conn.Write(body)
		if error != nil {
			return 0, error
		}
	}

	// 如果需要填充数据
	if int(header.paddingLength) > 0 {
		var b = make([]byte, header.paddingLength)
		conn.Write(b)
	}

	return length + headerLength, error
}

// FCGI_END_REQUEST
func WriteFcgiEndRequest(conn net.Conn, requestId int) (int, error)  {
	endRequestRecord := FCGI_EndRequestRecord {
		header: FCGI_Header{
			version:         1,
			typ:             FCGI_END_REQUEST,
			requestIdB0:     byte(requestId & 0xff),
			requestIdB1:     byte(requestId>>8) & 0xff,
			contentLengthB0: 0,
			contentLengthB1: 8,
		},
		body: FCGI_EndRequestBody{
			protocolStatus:FCGI_REQUEST_COMPLETE,
		},
	}

	var endRequestBodyLength = int(unsafe.Sizeof(endRequestRecord))
	bits := TransformBits(uintptr(unsafe.Pointer(&endRequestRecord)), endRequestBodyLength)
	_, error := conn.Write(bits)
	if error != nil {
		return 0, error
	}

	return endRequestBodyLength, nil
}
