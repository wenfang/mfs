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

var Cmd = map[string]func(uint64, io.ReadWriter){
	"get": getCmd,
	"put": putCmd,
  "del": delCmd,
}
var img *mfs.Img

func getCmd(objId uint64, c io.ReadWriter) {
	obj := img.GetObj(uint32(objId))
	if obj == nil {
		io.WriteString(c, "+E Object Not Found\r\n")
		return
	}
	io.WriteString(c, fmt.Sprintf("+S %d\r\n", obj.ObjLen))

	img.Get(obj, c)
	return
}

func putCmd(objLen uint64, c io.ReadWriter) {
	id := img.Put(objLen, c)
  if id == 0 {
    log.Println("Object Put Error")
    io.WriteString(c, "+E Put Object Error")
    return
  }
	io.WriteString(c, fmt.Sprintf("+S %d\r\n", id))
}

func delCmd(objId uint64, c io.ReadWriter) {
  res := img.Del(uint32(objId))
  if res == 0 {
    io.WriteString(c, "+E Del Object Not Found in Idx\r\n")
    return
  } else if res == 1 {
    io.WriteString(c, "+E Del Object Not Found in Obj\r\n")
    return
  }
  io.WriteString(c, "+S\r\n")
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

	if _, ok := Cmd[fields[0]]; !ok {
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

	Cmd[fields[0]](arg, c)
}

func init() {
	if img = mfs.NewImg("img"); img == nil {
		log.Fatal("Open img Error")
	}
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
