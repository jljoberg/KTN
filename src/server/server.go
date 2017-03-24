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
type connAndServerMsg struct {
	Socket *net.TCPConn
	Msg    T.ServerMsg
}
type connAndHistory struct {
	Socket  *net.TCPConn
	History T.HistoryMsg
}

func main() {
	var history [][]byte

	msgInCh := make(chan connAndMsg)
	msgOutCh := make(chan connAndServerMsg)
	historyOutCh := make(chan connAndHistory)
	connUserMap := make(map[*net.TCPConn]string)
	go netHandler(msgInCh, msgOutCh, historyOutCh, &connUserMap)

	for {
		select {
		case rx := <-msgInCh:
			rxSocket := rx.Socket
			rxMsg := rx.Msg

			switch {
			case rxMsg.Request == "login":
				if connUserMap[rxSocket] == "" && userNameValid(rxMsg.Content) && nameAvailible(rxMsg.Content, connUserMap) {
					connUserMap[rxSocket] = rxMsg.Content
					println("Logging in user", rxMsg.Content)
					msgOutCh <- connAndServerMsg{rxSocket,
						T.ServerMsg{time.Now().String(), "server", "info", "Login Successful"}}
					time.Sleep(30 * time.Millisecond)
					historyOutCh <- connAndHistory{rx.Socket,
						T.HistoryMsg{time.Now().String(), "server", "history", history}}
				} else if connUserMap[rxSocket] != "" {
					msgOutCh <- connAndServerMsg{rxSocket,
						T.ServerMsg{time.Now().String(), "server", "error", "Already logged in"}}
				} else if !nameAvailible(rxMsg.Content, connUserMap) {
					msgOutCh <- connAndServerMsg{rxSocket,
						T.ServerMsg{time.Now().String(), "server", "error", "Username taken"}}
				} else {
					msgOutCh <- connAndServerMsg{rxSocket,
						T.ServerMsg{time.Now().String(), "server", "error", "Username not valid"}}
				}

			case rxMsg.Request == "logout":
				connUserMap[rxSocket] = ""
				msgOutCh <- connAndServerMsg{rxSocket,
					T.ServerMsg{time.Now().String(), connUserMap[rxSocket], "info", "Logout Successful"}}
			case rxMsg.Request == "msg":
				if connUserMap[rxSocket] == "" {
					msgOutCh <- connAndServerMsg{rxSocket,
						T.ServerMsg{time.Now().String(), "server", "error", "You are not logged in"}}
				} else {
					srvMsg := T.ServerMsg{time.Now().String(), connUserMap[rxSocket], "msg", rxMsg.Content}
					bcast(connUserMap, srvMsg)
					msgBytes, _ := json.Marshal(srvMsg)
					history = append(history, msgBytes)
				}

			case rxMsg.Request == "names":
				names := ""
				for _, user := range connUserMap {
					names = names + user + "\n"
				}
				msgOutCh <- connAndServerMsg{rxSocket,
					T.ServerMsg{time.Now().String(), "server", "info", names}}
			}
		}
	}
}

func netHandler(msgInCh chan<- connAndMsg, msgOutCh <-chan connAndServerMsg, historyOutCh <-chan connAndHistory, connUserMapPtr *map[*net.TCPConn]string) {
	println("Enter host and port to allow connections...")
	reader := bufio.NewReader(os.Stdin)
	print("HOST: ")
	HOST, _ := reader.ReadString('\n')
	HOST = HOST[:len(HOST)-1]
	print("PORT: ")
	PORT, _ := reader.ReadString('\n')
	PORT = PORT[:len(PORT)-1]

	delConnCh := make(chan *net.TCPConn)
	idConnMap := make(map[int]*net.TCPConn)
	connArr := make([]*net.TCPConn, 8)
	tcpAddr, _ := net.ResolveTCPAddr("tcp", HOST+":"+PORT)
	listner, _ := net.ListenTCP("tcp", tcpAddr)

	newConnCh := make(chan connAndId)
	go getTcpConnections(newConnCh, listner)

	txCh := make(chan connAndServerMsg)
	txHistoryCh := make(chan connAndHistory)
	go transmitter(txCh, txHistoryCh)

	var selectReceiver []reflect.SelectCase
	rxCh := make(chan idAndMsg)
	go parseReceivers(&selectReceiver, rxCh)
	time.Sleep(50 * time.Millisecond) // let parser init before allow mod to selecReceiver

	println("Server is READY:\n--------------------------------------------\n")

	for {
		select {
		case newConn := <-newConnCh:
			idConnMap[newConn.Id] = &newConn.Socket
			(*connUserMapPtr)[&newConn.Socket] = ""
			connArr = append(connArr, &newConn.Socket)
			newChan := make(chan []byte)
			go receiver(&newConn.Socket, newConn.Id, newChan, delConnCh)
			selectReceiver = append(selectReceiver,
				reflect.SelectCase{
					Dir:  reflect.SelectRecv,
					Chan: reflect.ValueOf(newChan)},
			)

		case rx := <-rxCh:
			msgInCh <- connAndMsg{idConnMap[rx.Id], rx.Msg}
		case tx := <-msgOutCh:
			txCh <- tx
		case tx := <-historyOutCh:
			txHistoryCh <- tx
		case delConn := <-delConnCh:
			delete(*connUserMapPtr, delConn)
		}
	}
}

func parseReceivers(selectReceiverPtr *[]reflect.SelectCase, rxCh chan idAndMsg) {
	tickCh := time.Tick(5 * time.Second)
	*selectReceiverPtr = append(*selectReceiverPtr, reflect.SelectCase{Dir: reflect.SelectRecv,
		Chan: reflect.ValueOf(tickCh)})
	var rxIdAndMsg idAndMsg
	for {
		index, rxBytes, _ := reflect.Select(*selectReceiverPtr)
		//rif reflect.TypeOf(rxBytes) == time.Ticker {
		if index == 0 {
			continue
		}
		json.Unmarshal(rxBytes.Bytes(), &rxIdAndMsg)
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

func receiver(socket *net.TCPConn, id int, rxCh chan<- []byte, delConnCh chan<- *net.TCPConn) {
	println("Starting a new receiver")
	println("I am using socket: ", socket, " | id: ", id)
	var b [4096]byte
	var clMsg T.ClientMsg
	for {
		n, err := socket.Read(b[:])
		if err != nil {
			println("A receiver encountered an error and is quitting")
			delConnCh <- socket
			return
		}
		json.Unmarshal(b[:n], &clMsg)
		rxBytes, _ := json.Marshal(idAndMsg{id, clMsg})
		rxCh <- rxBytes

	}
}

func transmitter(txCh <-chan connAndServerMsg, txHistoryCh <-chan connAndHistory) {
	i := 0
	for {
		select {
		case tx := <-txCh:
			b, _ := json.Marshal(tx.Msg)
			tx.Socket.Write(b)
		case tx := <-txHistoryCh:
			println("History: ------____--------\n")
			if len(tx.History.Content) > 0 {
				println(len(tx.History.Content))
				b, _ := json.Marshal(tx.History)
				tx.Socket.Write(b)
				i++
			}
		}
	}
}

func bcast(connUserMap map[*net.TCPConn]string, msg T.ServerMsg) {
	for socket, user := range connUserMap {
		if user == "" {
			continue
		}
		b, _ := json.Marshal(msg)
		socket.Write(b)
	}
}

func userNameValid(name string) bool {
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func nameAvailible(name string, connUserMap map[*net.TCPConn]string) bool {
	for _, nameTaken := range connUserMap {
		if name == nameTaken {
			return false
		}
	}
	return true
}
