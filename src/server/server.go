package main

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"reflect"
	"time"
	T "typeDef"
)

type connAndBytes struct {
	Socket *net.TCPConn
	Bytes  []byte
}

type recvConnAndMsg struct {
	Socket  *net.TCPConn
	RecvMsg T.ClientMsg
}

func main() {
	println("Starting Server:\n--------------------------\n\n")

	msgInCh := make(chan T.ClientMsg)
	msgOutCh := make(chan T.ServerMsg)
	go netHandler(msgInCh, msgOutCh)

	select {} // Holds program

}

func netHandler(msgInCh <-chan T.ClientMsg, msgOutCh chan<- T.ServerMsg) {
	println("Enter host and port to allow connections...")
	reader := bufio.NewReader(os.Stdin)
	print("HOST: ")
	HOST, _ := reader.ReadString('\n')
	HOST = HOST[:len(HOST)-1]
	print("PORT: ")
	PORT, _ := reader.ReadString('\n')
	PORT = PORT[:len(PORT)-1]

	connMap := make(map[*net.TCPConn]string)
	connArr := make([]*net.TCPConn, 8)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", HOST+":"+PORT)
	listner, _ := net.ListenTCP("tcp", tcpAddr)

	newConnCh := make(chan net.TCPConn)
	go getTcpConnections(newConnCh, listner)

	var selectReceiver []reflect.SelectCase
	rxCh := make(chan recvConnAndMsg)
	go parseReceivers(&selectReceiver, rxCh)

	for {
		select {
		case newConn := <-newConnCh:
			connMap[&newConn] = ""
			connArr = append(connArr, &newConn)
			newChan := make(chan connAndBytes)
			go receiver(&newConn, newChan)
			selectReceiver = append(selectReceiver,
				reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(newChan)},
			)

		case rx := <-rxCh:
			println("netHandler received request: ", rx.RecvMsg.Request)
			println("content: ", rx.RecvMsg.Content)
			println("from socket: ", rx.Socket)
		}

	}
}

func parseReceivers(selectReceiverPtr *[]reflect.SelectCase, rxCh chan recvConnAndMsg) {
	println("parseRecvs: len of init arr is: ", len(*selectReceiverPtr))
	for len(*selectReceiverPtr) == 0 {
		time.Sleep(time.Second)
		println("Empty arr in parseReceivers")
	}
	var connAndMsg recvConnAndMsg
	for {
		println("parseRecvs: len of current arr is: ", len(*selectReceiverPtr), "\n")
		_, rxConnAndBytes, _ := reflect.Select(*selectReceiverPtr)
		println("Parser succesfully received something")
		rxBytes := rx.Bytes()
		json.Unmarshal(rxBytes, &connAndMsg)
		println("parser got msg from socket: ", connAndMsg.Socket)
		rxCh <- connAndMsg
		println("Parser succesfully SENT something")
		time.Sleep(30 * time.Millisecond)
	}
}

func getTcpConnections(newConnCh chan<- net.TCPConn, listner *net.TCPListener) {
	for {
		conn, _ := listner.AcceptTCP()
		println("getTcpConnections(..): I got a new connection!!\n")
		newConnCh <- *conn
	}
}

func receiver(socket *net.TCPConn, rxCh chan<- connAndBytes) {
	println("Starting a new receiver")
	println("I am using socket: ", socket)
	var b [4096]byte
	//var clMsg T.ClientMsg
	for {
		n, err := socket.Read(b[:])
		if err != nil {
			println("A receiver encountered an error and is quitting")
			return
		}
		println("A receiver Read a msg, applying my socket:", socket)
		//json.Unmarshal(b[:n], &clMsg)
		//rxBytes, _ := json.Marshal(recvConnAndMsg{socket, clMsg})
		rxCh <- connAndBytes{socket, b[:n]}

	}
}
