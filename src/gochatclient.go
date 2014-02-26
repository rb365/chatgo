package main

import "net"
import (
	"log"
	"math/rand"
	"strconv"
)

const port string = "localhost:16180"

func main() {

	for i:=0;i < 1000;i++ {
		conn, err := net.Dial("tcp", port)
		if(err != nil) {
			log.Fatalln("Error connecting to", port, err.Error())
		}
		conn.Write([]byte( strconv.Itoa(rand.Int())))

		go handleConnection(conn)
	}
	conn, err := net.Dial("tcp", port)
	if(err != nil) {
		log.Fatalln("Error connecting to", port, err.Error())
	}
	conn.Write([]byte( strconv.Itoa(rand.Int())))

	handleConnection(conn)
}

func isControl(c byte) bool {
	return !(c >= 32 && c != 127)
}

func handleConnection(con net.Conn) {
	data := make([]byte, 500)
	dataNoCtrlChar := make([]byte, 500)
	//address := con.RemoteAddr().String()
	for {
		n, err := con.Read(data)
		if err != nil {
			log.Print("Error", err.Error())
			continue
		}
		j := 0
		for i := 0; i < n; i++ {
			if !isControl(data[i]) {
				dataNoCtrlChar[j] = data[i]
				j++
			}
		}
		//log.Println(address, "data:", string(dataNoCtrlChar))
		con.Write([]byte("/join test" + strconv.Itoa(rand.Int()%10)))

	}
}
