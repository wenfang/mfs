package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"mfs"
	"net"
	"strconv"
	"strings"
)

const (
	idxSizeBit = 4
	datFlagBit = 8
)

type cFun func(uint64, io.ReadWriter)

var cmdFun map[string]cFun
var img *mfs.Img

func getCmd(objId uint64, c io.ReadWriter) {
  entry := img.GetObjIdxEntry(uint32(objId))
  if entry == nil {
    io.WriteString(c, fmt.Sprintf("+E No Obj Found\r\n"))
    return
  }

  io.WriteString(c, fmt.Sprintf("+S %d\r\n", entry.ObjLen))
	img.GetObj(entry, c)
	return
}

func putCmd(objLen uint64, c io.ReadWriter) {
	id := img.PutObj(objLen, c)
  io.WriteString(c, fmt.Sprintf("+S %d\r\n", id))
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

	cmdFun[fields[0]](arg, c)
}

func init() {
	cmdFun = map[string]cFun{
		"get": getCmd,
		"put": putCmd,
	}

	if img = mfs.NewImg("img"); img == nil {
		log.Fatal("Open img Error")
	}
	img.LoadSuper()
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
