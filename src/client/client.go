package main

//import "fmt"
import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"strings"
	"time"
	T "typeDef"
)

func main() {

	netOutCh := make(chan T.ClientPayload)
	netInCh := make(chan T.ServerPayload)

	go netMsgHandler(netOutCh, netInCh, netInit())
	time.Sleep(50 * time.Millisecond)

	inputCh := make(chan T.ClientPayload)
	go userInput(inputCh)

	for {
		select {
		case input := <-inputCh:
			netOutCh <- input
		case rx := <-netInCh:
			if rx.Response == "message" || rx.Response == "msg" {
				println(rx.Timestamp, " -:- ", rx.Sender, ":")
			} else {
				println(rx.Response, "------- : ")
			}
			println(rx.Content)
		}
		println("-------------------------------------------------------")
	}
}

func netInit() string {
	println("Enter host and port to connect to a server...")
	reader := bufio.NewReader(os.Stdin)
	print("HOST: ")
	HOST, _ := reader.ReadString('\n')
	HOST = HOST[:len(HOST)-1]
	print("PORT: ")
	PORT, _ := reader.ReadString('\n')
	PORT = PORT[:len(PORT)-1]
	return HOST + ":" + PORT
}

func netMsgHandler(netOutCh <-chan T.ClientPayload, netInCh chan<- T.ServerPayload, hostAndPort string) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", hostAndPort)
	socket, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer socket.Close()

	rxCh := make(chan T.ServerPayload)
	txCh := make(chan T.ClientPayload)
	go receiver(rxCh, socket)
	go transmitter(txCh, socket)

	for {
		select {
		case rx := <-rxCh:
			netInCh <- rx
		case tx := <-netOutCh:
			txCh <- tx
		}
	}

}

func receiver(rxCh chan<- T.ServerPayload, socket *net.TCPConn) {
	var b [4096]byte
	var srvMsg T.ServerPayload
	var hMsg T.HistoryPayload
	for {
		n, _ := socket.Read(b[:])
		err := json.Unmarshal(b[:n], &srvMsg)
		if err != nil {
			err := json.Unmarshal(b[:n], &hMsg)
			if err != nil {
				println("unmarshal error in second loop")
				continue
			}
			println("   ---     GETTING HISTORY   ---\n---------------------------------------")
			for _, hB := range hMsg.Content {
				json.Unmarshal(hB, &srvMsg)
				rxCh <- srvMsg
				continue
			}
			continue
		}
		rxCh <- srvMsg
	}
}
func transmitter(txCh <-chan T.ClientPayload, socket *net.TCPConn) {
	for {
		msg := <-txCh
		b, _ := json.Marshal(msg)
		socket.Write(b)
	}
}

func clientMsgConstructor(request, content string) (T.ClientPayload, bool) {
	var validRequests []string
	validRequests = append(validRequests, "login", "logout")
	validRequests = append(validRequests, "msg")
	validRequests = append(validRequests, "names")
	validRequests = append(validRequests, "help")

	var clientMsg T.ClientPayload
	for _, legalRequest := range validRequests {
		if request == legalRequest {
			clientMsg.Request = request
			clientMsg.Content = content
			return clientMsg, true
		}
	}
	clientMsg.Request = request
	clientMsg.Content = content
	return clientMsg, false
}

func userInput(inputCh chan<- T.ClientPayload) {
	reader := bufio.NewReader(os.Stdin)
	var request, content string
	println("Prefix commands with '/'. Type '/help' for availible commands")
	for {
		input, _ := reader.ReadString('\n')
		input = input[:len(input)-1]
		if len(input) == 0 {
			continue
		}
		if input[0] == '/' {
			temp := strings.SplitAfterN(input, " ", 2)
			request = (temp[0])[1:]
			if len(temp) > 1 {
				content = temp[1]
				println(request)
			}
		} else {
			request = "msg"
			content = input
		}
		for request[len(request)-1] == ' ' {
			request = request[:len(request)-1]
		}

		clientMsg, ok := clientMsgConstructor(request, content)
		if ok {
			inputCh <- clientMsg
		} else {
			println("ClientMsg NOT ok ---///---///---///")
			inputCh <- clientMsg
		}
		time.Sleep(100 * time.Millisecond)
	}
}
