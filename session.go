package autobahn

import (
	"github.com/gopherjs/gopherjs/js"
)

type Session struct {
	js.Object
}

type SubMessage struct {
	args    []js.Object
	kwargs  js.Object
	details js.Object
}

func (s *Session) GetId() int {
	return s.Get("id").Int()
}

func (s *Session) IsOpen() bool {
	return s.Get("isOpen").Bool()
}

func (s *Session) GetSubscriptions() []Subscription {
	obj := s.Get("subscriptions")
	l := s.Length()
	ret := make([]Subscription, l)
	for i := 0; i < l; i++ {
		ret[i] = Subscription{obj.Index(i)}
	}
	return ret
}

func (s *Session) Log(obj js.Object) {
	s.Call("log", obj)
}

func (s *Session) Prefix(prefix, uri string) {
	s.Call("prefix", prefix, uri)
}

func (s *Session) Subscribe(topic string, ch chan<- *SubMessage, options ...Option) (*Subscription, error) {
	// Create the callback-to-channel handler.
	handler := func(argsobj js.Object, kwargs, details js.Object) {
		l := argsobj.Length()
		args := make([]js.Object, l)
		for i := 0; i < l; i++ {
			args[i] = argsobj.Index(i)
		}
		ch <- &SubMessage{args, kwargs, details}
	}
	// Create subscription with handler.
	obj, err := HandlePromise(s.Call("subscribe", handler))
	if err != nil {
		return nil, err
	}
	return &Subscription{obj}, nil
}

func (s *Session) Unsubscribe(sub *Subscription) (bool, error) {
	ret, err := HandlePromise(s.Call("unsubscribe", sub))
	if err != nil {
		return false, err
	}
	return ret.Bool(), nil
}

func (s *Session) Publish(topic string, args []js.Object, kwargs js.Object, options ...Option) (js.Object, JSError) {
	opts := map[string]js.Object{}
	for _, o := range options {
		opts[o.key] = o.value
	}
	return HandlePromise(s.Call("publish", topic, args, kwargs, opts))
}

func (s *Session) CallRPC(procedure string, args []js.Object, kwargs js.Object, options ...Option) (*Result, JSError) {
	opts := make(map[string]interface{})
	for _, o := range options {
		opts[o.key] = o.value
	}
	opts["receive_progress"] = false
	result, err := HandlePromise(s.Call("call", procedure, args, kwargs, opts))
	if err != nil {
		return nil, err
	}
	return &Result{result}, nil
}

func (s *Session) CallRPCProgressive(procedure string, args []js.Object, kwargs js.Object, options ...Option) (<-chan *Result, <-chan JSError, <-chan js.Object) {
	opts := make(map[string]interface{})
	for _, o := range options {
		opts[o.key] = o.value
	}
	opts["receive_progress"] = false
	promise := s.Call("call", procedure, args, kwargs, opts)

	prog := make(chan js.Object, 100)
	errchan := make(chan JSError)
	final := make(chan *Result)

	go func() {
		promise.Call(
			"then",
			func(obj js.Object) {
				close(prog)
				final <- &Result{obj}
			},
			func(obj js.Object) {
				close(prog)
				errchan <- &jserror{obj}
			},
			func(obj js.Object) {
				prog <- obj
			},
		)
	}()

	return final, errchan, prog
}
