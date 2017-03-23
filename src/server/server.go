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

type connAndId struct {
	Socket net.TCPConn
	Id     int
}

type idAndMsg struct {
	Id  int
	Msg T.ClientMsg
}
type connAndMsg struct {
	Socket *net.TCPConn
	Msg    T.ClientMsg
}

func main() {
	println("Starting Server:\n--------------------------\n\n")

	msgInCh := make(chan connAndMsg)
	msgOutCh := make(chan T.ServerMsg)
	connUserMap := make(map[*net.TCPConn]string)
	go netHandler(msgInCh, msgOutCh, &connUserMap)

	for {
		select {
		case rx := <-msgInCh:
			rxSocket := rx.Socket
			rxMsg := rx.Msg

			switch {
			case rxMsg.Request == "login":
				if connUserMap[rxSocket] == "" {
					connUserMap[rxSocket] = rxMsg.Content
					println("Logging in user", rxMsg.Content)
				} else {
					println("Already logged in")
				}
			case rxMsg.Request == "logout":
				connUserMap[rxSocket] = ""
			case rxMsg.Request == "msg":
				println("MAIN: Send msg: \n", rxMsg.Content, "------------------------\n")

			}
		}
	}

}

func netHandler(msgInCh chan<- connAndMsg, msgOutCh <-chan T.ServerMsg, connUserMapPtr *map[*net.TCPConn]string) {
	println("Enter host and port to allow connections...")
	reader := bufio.NewReader(os.Stdin)
	print("HOST: ")
	HOST, _ := reader.ReadString('\n')
	HOST = HOST[:len(HOST)-1]
	print("PORT: ")
	PORT, _ := reader.ReadString('\n')
	PORT = PORT[:len(PORT)-1]

	idConnMap := make(map[int]*net.TCPConn)
	connArr := make([]*net.TCPConn, 8)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", HOST+":"+PORT)
	listner, _ := net.ListenTCP("tcp", tcpAddr)

	newConnCh := make(chan connAndId)
	go getTcpConnections(newConnCh, listner)

	var selectReceiver []reflect.SelectCase
	rxCh := make(chan idAndMsg)
	go parseReceivers(&selectReceiver, rxCh)
	time.Sleep(50 * time.Millisecond) // let parser init before allow mod to selecReceiver

	for {
		select {
		case newConn := <-newConnCh:
			idConnMap[newConn.Id] = &newConn.Socket
			(*connUserMapPtr)[&newConn.Socket] = ""
			connArr = append(connArr, &newConn.Socket)
			newChan := make(chan []byte)
			go receiver(&newConn.Socket, newConn.Id, newChan)
			selectReceiver = append(selectReceiver,
				reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(newChan)},
			)

		case rx := <-rxCh:
			println("netHandler received request: ", rx.Msg.Request)
			println("content: ", rx.Msg.Content)
			println("from id: ", rx.Id, "  | using socket: ", idConnMap[rx.Id])

			msgInCh <- connAndMsg{idConnMap[rx.Id], rx.Msg}

		case tx := <-msgOutCh:
			println("outgoing;", tx.Response)
		}
	}
}

func parseReceivers(selectReceiverPtr *[]reflect.SelectCase, rxCh chan idAndMsg) {
	tickCh := time.Tick(5 * time.Second)
	*selectReceiverPtr = append(*selectReceiverPtr, reflect.SelectCase{Dir: reflect.SelectRecv,
		Chan: reflect.ValueOf(tickCh)})
	var rxIdAndMsg idAndMsg
	for {
		println("parseRecvs: len of current arr is: ", len(*selectReceiverPtr), "\n")
		index, rxBytes, _ := reflect.Select(*selectReceiverPtr)
		//rif reflect.TypeOf(rxBytes) == time.Ticker {
		if index == 0 {
			continue
		}
		json.Unmarshal(rxBytes.Bytes(), &rxIdAndMsg)
		println("parser got msg from id: ", rxIdAndMsg.Id)
		rxCh <- rxIdAndMsg
		time.Sleep(30 * time.Millisecond)
	}
}

func getTcpConnections(newConnCh chan<- connAndId, listner *net.TCPListener) {
	for id := 0; true; id++ {
		conn, _ := listner.AcceptTCP()
		println("getTcpConnections(..): I got a new connection!!")
		println("I'm assiging id: ", id, "\n")
		newConnCh <- connAndId{*conn, id}
	}
}

func receiver(socket *net.TCPConn, id int, rxCh chan<- []byte) {
	println("Starting a new receiver")
	println("I am using socket: ", socket, " | id: ", id)
	var b [4096]byte
	var clMsg T.ClientMsg
	for {
		n, err := socket.Read(b[:])
		if err != nil {
			println("A receiver encountered an error and is quitting")
			return
		}
		println("A receiver Read a msg, applying my socket:", socket)
		json.Unmarshal(b[:n], &clMsg)
		rxBytes, _ := json.Marshal(idAndMsg{id, clMsg})
		rxCh <- rxBytes

	}
}
