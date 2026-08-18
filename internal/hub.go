package internal

var userClientMap map[string]chan []byte = make(map[string]chan []byte, 10)

type Hub struct {
	broadcast  chan []byte // Messages of HTML that are sent out to any user showing the global state
	register   chan Client
	unregister chan Client
	clients    map[Client]struct{}
}

type Client chan []byte

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan Client),
		unregister: make(chan Client),
		clients:    make(map[Client]struct{}),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			delete(h.clients, client)
			close(client)
		case html := <-h.broadcast:
			for client := range h.clients {
				select {
				case client <- html:
				default:
				}
			}
		}
	}
}
