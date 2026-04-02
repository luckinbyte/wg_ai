package agent

type HandlerFunc func(a *Agent, msg *Message) ([]byte, error)

type Dispatcher struct {
	handlers map[uint16]HandlerFunc
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[uint16]HandlerFunc),
	}
}

func (d *Dispatcher) Register(msgID uint16, handler HandlerFunc) {
	d.handlers[msgID] = handler
}

func (d *Dispatcher) Get(msgID uint16) HandlerFunc {
	return d.handlers[msgID]
}

func (d *Dispatcher) Dispatch(a *Agent, msg *Message) ([]byte, error) {
	handler := d.handlers[msg.MsgID]
	if handler == nil {
		return nil, nil
	}
	return handler(a, msg)
}
