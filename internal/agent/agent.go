package agent

import (
	"sync"

	"github.com/luckinbyte/wg_ai/internal/session"
)

type Message struct {
	MsgID    uint16
	Sequence uint32
	Payload  []byte
	Sess     *session.Session
}

type Agent struct {
	ID         int
	players    map[int64]*session.Session
	msgQueue   chan *Message
	mutex      sync.RWMutex
	stopCh     chan struct{}
	dispatcher *Dispatcher
	fallback   HandlerFunc
}

func New(id, queueSize int) *Agent {
	a := &Agent{
		ID:         id,
		players:    make(map[int64]*session.Session),
		msgQueue:   make(chan *Message, queueSize),
		stopCh:     make(chan struct{}),
		dispatcher: NewDispatcher(),
	}
	RegisterDefaultHandlers(a)
	return a
}

func (a *Agent) SetFallback(fn HandlerFunc) {
	a.fallback = fn
}

func (a *Agent) Push(msg *Message) {
	select {
	case a.msgQueue <- msg:
	default:
		// queue full, drop message
	}
}

func (a *Agent) Run() {
	for {
		select {
		case msg := <-a.msgQueue:
			a.handleMessage(msg)
		case <-a.stopCh:
			return
		}
	}
}

func (a *Agent) Stop() {
	close(a.stopCh)
}

func (a *Agent) handleMessage(msg *Message) {
	resp, err := a.dispatcher.Dispatch(a, msg)
	if err != nil {
		// TODO: send error response
		return
	}
	if resp != nil && msg.Sess != nil {
		msg.Sess.Send(resp)
	}
}

func (a *Agent) BindSession(sess *session.Session) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.players[sess.RID] = sess
}

func (a *Agent) UnbindSession(rid int64) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	delete(a.players, rid)
}

func (a *Agent) GetSession(rid int64) *session.Session {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.players[rid]
}

func (a *Agent) RegisterHandler(msgID uint16, handler HandlerFunc) {
	a.dispatcher.Register(msgID, handler)
}
