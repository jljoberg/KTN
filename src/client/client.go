package main

//import "fmt"
import "encoding/json"
import "bufio"
import "net"
import "os"
import "time"
import T "typeDef"

func main() {

	netOutCh := make(chan T.ClientMsg)
	netInCh := make(chan T.ServerMsg)

	go netMsgHandler(netOutCh, netInCh, netInit())
	time.Sleep(50 * time.Millisecond)

	inputCh := make(chan T.ClientMsg)
	go userInput(inputCh)

	for {
		select {
		case input := <-inputCh:
			netOutCh <- input
		case rx := <-netInCh:
			if rx.Response == "msg" {
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

func netMsgHandler(netOutCh <-chan T.ClientMsg, netInCh chan<- T.ServerMsg, hostAndPort string) {
	tcpAddr, _ := net.ResolveTCPAddr("tcp", hostAndPort)
	socket, _ := net.DialTCP("tcp", nil, tcpAddr)
	defer socket.Close()

	rxCh := make(chan T.ServerMsg)
	txCh := make(chan T.ClientMsg)
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

func receiver(rxCh chan<- T.ServerMsg, socket *net.TCPConn) {
	var b [4096]byte
	var srvMsg T.ServerMsg
	var hMsg T.HistoryMsg
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
func transmitter(txCh <-chan T.ClientMsg, socket *net.TCPConn) {
	for {
		msg := <-txCh
		b, _ := json.Marshal(msg)
		socket.Write(b)
	}
}

func clientMsgConstructor(request, content string) (T.ClientMsg, bool) {
	var validRequests []string
	validRequests = append(validRequests, "login", "logout")
	validRequests = append(validRequests, "msg")
	validRequests = append(validRequests, "names")
	validRequests = append(validRequests, "help")

	var clientMsg T.ClientMsg
	for _, legalRequest := range validRequests {
		if request == legalRequest {
			clientMsg.Request = request
			clientMsg.Content = content
			return clientMsg, true
		}
	}
	return clientMsg, false
}

func userInput(inputCh chan<- T.ClientMsg) {
	reader := bufio.NewReader(os.Stdin)
	for {
		print("Request: ")
		request, _ := reader.ReadString('\n')
		request = request[:len(request)-1]
		print("Content: ")
		content, _ := reader.ReadString('\n')
		content = content[:len(content)-1]
		println()

		clientMsg, ok := clientMsgConstructor(request, content)
		if ok {
			inputCh <- clientMsg
		} else {
			println("ClientMsg NOT ok ---///---///---///")
		}
		time.Sleep(100 * time.Millisecond)
	}
}
