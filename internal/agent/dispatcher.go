package agent

import "log"

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
		if a.fallback != nil {
			log.Printf("[Dispatcher] no handler for msgID=%d, using fallback", msg.MsgID)
			return a.fallback(a, msg)
		}
		log.Printf("[Dispatcher] no handler for msgID=%d, no fallback, dropping message", msg.MsgID)
		return nil, nil
	}
	log.Printf("[Dispatcher] dispatching msgID=%d", msg.MsgID)
	return handler(a, msg)
}
