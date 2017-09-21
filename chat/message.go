package chat

type Message struct {
	Path string `json:"path"`
}

func (self *Message) String() string {
	return self.Path
}
