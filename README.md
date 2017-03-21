
# PROSJEKT TTM4100 - Chat server and client #

## TODO

In server: I can't send socket adress using json. Try implementing some kind of id system to relate to sockets


## KTN1 - desctription of design ##

I will implement the chat in golang
The messages will be represented using a struct, but have to be transferred as strings to satisfy protocol

### Server ###
The main server thread will communicate with serverIFace using a single channel. This channel will transfer messages using structs.
The main thread will keep a map of all active usernames, and refer to either the channel or socket it's bound to.
The serverIFace thread will manage threads called serversideConnection for each connection to clients.
The connection threads will be accessed by serverIFace using the []reflect.SelectCase method.


### Client 
The main client thread will use a channel to clientIFace to send and recieve messages.
clientIFace will set up a TCP-connection with a dedicated serversideConnection thread.

### Misc ###
The two different structs used for client- and server-messages might need to be placed into a superstruct in order to be able to send both struct over a single channel.
Alteratively, I might use a set of two channels to send the respective structs. 
