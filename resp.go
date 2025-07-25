package main

import (
	"bufio"
	"io"
	"log"
	"strconv"
)

// RespReader which represents the Redis Serialization Prtotocol (RESP) format.
type RespReader struct {
	reader *bufio.Reader
}

const (
	arrayPrefix      = '*'
	bulkStringPrefix = '$'
)

const (
	crlf = "\r\n"
)

func NewRespReader(r io.Reader) *RespReader {
	return &RespReader{
		reader: bufio.NewReader(r),
	}
}

func (r *RespReader) Args() (args []string, err error) {
	argsLength := 0
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 || line[0] != arrayPrefix {
		err = io.ErrUnexpectedEOF
		return
	}
	line = line[1:] // Remove the '*' prefix
	argsLength, err = strconv.Atoi(line)
	if err != nil {
		return
	}
	args = make([]string, 0, argsLength)
	for i := 0; i < argsLength; i++ {
		arg, rErr := r.readBulkString()
		if rErr != nil {
			err = rErr
			return
		}
		args = append(args, arg)
	}
	return
}

func (r *RespReader) readBulkString() (string, error) {
	line, err := r.readLine()
	if err != nil {
		return "", err
	}
	if len(line) == 0 || line[0] != bulkStringPrefix {
		return "", io.ErrUnexpectedEOF
	}
	line = line[1:] // Remove the '$' prefix
	length, err := strconv.Atoi(line)
	if err != nil {
		return "", err
	}
	if length < 0 {
		return "", nil // Null bulk string
	}
	bulkData := make([]byte, length)
	for i := 0; i < length; i++ {
		bulkData[i], err = r.reader.ReadByte()
		if err != nil {
			return "", err
		}
	}
	// Read the CRLF at the end of the bulk string
	crlf, err := r.readLine()
	if err != nil || crlf != "" {
		return "", io.ErrUnexpectedEOF
	}
	return string(bulkData), nil
}

func (r *RespReader) readLine() (line string, err error) {
	b, isPrefix, err := r.reader.ReadLine()
	if err != nil {
		return
	}
	if isPrefix {
		log.Println("Line is too long, not supported yet")
		return "", err
	}
	line = string(b)

	return line, nil
}
