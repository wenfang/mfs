package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	idxSizeBit = 4
	datFlagBit = 8
)

type WriteData struct {
	c      net.Conn
	datLen int64
	fin    chan string
}

type cFun func(net.Conn, uint64) string

var site int64
var wchan chan WriteData
var cmdFun map[string]cFun

func getCmd(c net.Conn, id uint64) string {
	ri, err := os.Open("idx")
	if err != nil {
		log.Println(err)
		return "+E Open idx Error\r\n"
	}
	defer ri.Close()

	buf := make([]byte, 16)
	idxOff := int64((id & 0x00000000FFFFFFFF) << idxSizeBit)
	if _, err = ri.ReadAt(buf, idxOff); err != nil {
		log.Println(err, idxOff)
		return "+E Read idx Error\r\n"
	}

	datOff, _ := binary.Varint(buf[0:8])
	datLen, _ := binary.Varint(buf[8:])
	datLen >>= datFlagBit

	rd, err := os.Open("data")
	if err != nil {
		log.Println(err)
		return "+E Open Data Error\r\n"
	}
	defer rd.Close()
	rd.Seek(datOff, 0)

	io.WriteString(c, fmt.Sprintf("+S %d\r\n", datLen))
	if _, err = io.CopyN(c, rd, datLen); err != nil {
		log.Println(err)
	}

	return ""
}

func putCmd(c net.Conn, size uint64) string {
	fin := make(chan string)
	wchan <- WriteData{c, int64(size), fin}
	return <-fin
}

func delCmd(c net.Conn, id uint64) string {
	return ""
}

func putHandle(wi, wd *os.File) {
	var idxOff, datOff int64
	var err error

	buf := make([]byte, 16)

	for v := range wchan {
		if idxOff, err = wi.Seek(0, 1); err != nil {
			v.fin <- "+E Idx Offset Read Error"
			continue
		}

		if datOff, err = wd.Seek(0, 1); err != nil {
			v.fin <- "+E Data Offset Read Error"
			continue
		}
		// write dat
		if _, err = io.CopyN(wd, v.c, v.datLen); err != nil {
			v.fin <- "+E Data Write Error"
			continue
		}
		// write idx
		binary.PutVarint(buf, datOff)
		binary.PutVarint(buf[8:], v.datLen<<datFlagBit)
		if _, err := wi.Write(buf); err != nil {
			v.fin <- "+E Idx Write Error"
			continue
		}

		v.fin <- fmt.Sprintf("+S %d\r\n", (site<<32)+(idxOff>>idxSizeBit))
	}
}

func mainHandle(c net.Conn) {
	defer c.Close()

	line, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		log.Println(err)
		io.WriteString(c, "+E Read Client Error\r\n")
		return
	}
	fields := strings.Fields(line)

	if _, ok := cmdFun[fields[0]]; !ok {
		io.WriteString(c, "+E Command Not Found\r\n")
		return
	}
	if len(fields) != 2 {
		io.WriteString(c, "+E Command Arg Error\r\n")
		return
	}
	arg, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		io.WriteString(c, "+E Command Arg Parse Error\r\n")
		return
	}

	io.WriteString(c, cmdFun[fields[0]](c, arg))
}

func init() {
	site = 1
	wchan = make(chan WriteData, 512)
	cmdFun = map[string]cFun{
		"get": getCmd,
		"put": putCmd,
		"del": delCmd,
	}

	var err error
	var wi, wd *os.File
	if wi, err = os.OpenFile("idx", os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		log.Fatal(err)
	}
	if _, err = wi.Seek(0, 2); err != nil {
		log.Fatal(err)
	}

	if wd, err = os.OpenFile("data", os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		log.Fatal(err)
	}
	if _, err = wd.Seek(0, 2); err != nil {
		log.Fatal(err)
	}

	go putHandle(wi, wd)
}

func main() {
	l, err := net.Listen("tcp4", ":7879")
	if err != nil {
		log.Fatal(err)
	}

	for {
		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go mainHandle(c)
	}
}
