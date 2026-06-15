package internal

type Hub struct {
	broadcast  chan []byte
	register   chan chan []byte
	unregister chan chan []byte
	clients    map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan chan []byte),
		unregister: make(chan chan []byte),
		clients:    make(map[chan []byte]struct{}),
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
