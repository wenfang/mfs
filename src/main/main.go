package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"mfs"
	"net"
	"strconv"
	"strings"
)

var Cmd = map[string]func([]string, io.ReadWriter) error{
	"get":    getCmd,
	"put":    putCmd,
	"del":    delCmd,
	"update": updateCmd,
}
var img *mfs.Img

func getCmd(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("Get Args Number Error")
	}

	objId, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		log.Println(err)
		return errors.New("Get Args Parse Error")
	}

	objLen, err := img.GetObjLen(uint32(objId))
	if err != nil {
		log.Println(err)
		return err
	}
	io.WriteString(c, fmt.Sprintf("+S %d\r\n", objLen))

	if err = img.Get(uint32(objId), c); err != nil {
		log.Println(err)
	}
	return nil
}

func putCmd(args []string, c io.ReadWriter) error {
	if len(args) != 2 {
		return errors.New("Put Args Number Error")
	}

	objLen, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		log.Println(err)
		return errors.New("Put Args Parse Len Error")
	}

	objSize, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		log.Println(err)
		return errors.New("Put Args Parse Size Error")
	}

	id, err := img.Put(objLen, objSize, c)
	if err != nil {
		log.Println(err)
		return err
	}
	io.WriteString(c, fmt.Sprintf("+S %d\r\n", id))
	return nil
}

func delCmd(args []string, c io.ReadWriter) error {
  if len(args) != 1 {
    return errors.New("Del Args Number Error")
  }

	objId, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		log.Println(err)
		return errors.New("Del Args Parse Len Error")
	}

	if err = img.Del(uint32(objId)); err != nil {
    log.Println(err)
    return err
  }
	io.WriteString(c, "+S\r\n")
	return nil
}

func updateCmd(args []string, c io.ReadWriter) error {
	return nil
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

  if err = Cmd[fields[0]](fields[1:], c); err != nil {
		io.WriteString(c, fmt.Sprintf("+E %s\r\n", err.Error()))
	}
}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	var err error
	if img, err = mfs.NewImg("img"); err != nil {
		log.Fatal("Open img Error")
	}

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
