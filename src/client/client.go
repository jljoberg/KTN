package main

//import "fmt"
import "bufio"
import "os"
import T "typeDef"

func main() {

	netOutCh := make(chan T.ClientMsg)
	netInCh := make(chan T.ServerMsg)

	go netMsgHandler(netOutCh, netInCh)

	reader := bufio.NewReader(os.Stdin)
	print("Request: ")
	request, _ := reader.ReadString('\n')
	request = request[:len(request)-1]
	print("Content(if neccessary)")
	content, _ := reader.ReadString('\n')
	println()

	clientMsg, ok := clientMsgConstructor(request, content)
	if ok {
		println("ClientMsg ok -------------")
		println("Request is: ", clientMsg.Request)
		if len(clientMsg.Content) > 0 {
			println("Content is: ", clientMsg.Content)
		}
	} else {
		println("ClientMsg NOT ok ---///---///---///")
	}

}

func netMsgHandler(netOutCh <-chan T.ClientMsg, netInCh chan<- T.ServerMsg) {

	//	for err:=
	println("Enter host and port to connect to a server...")
	reader := bufio.NewReader(os.Sdin)
	print("HOST: ")
	HOST, err := reader.ReadString('\n')
	HOST = HOST[:len(HOST)-1]

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
