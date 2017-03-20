package main

//import "fmt"
import "bufio"

func main() bool {
	reader := buifio.NewReader(os.Stdin)
	print("Type something: ")
	text, _ := reader.ReadString('\n')
	// fmt.Scan()
	println("You entered: ", text)

	return true

}
