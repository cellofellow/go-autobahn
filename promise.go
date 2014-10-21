package autobahn

import (
	"github.com/gopherjs/gopherjs/js"
)

type JSError interface {
	js.Object
	error
}

type jserror struct {
	js.Object
}

func (j *jserror) Error() string {
	return j.Str()
}

func HandlePromise(p js.Object) (js.Object, JSError) {
	ch := make(chan js.Object)
	errchan := make(chan js.Object)
	var obj, err js.Object
	go func() {
		p.Call(
			"then",
			func(obj js.Object) { ch <- obj },
			func(err js.Object) { errchan <- err },
		)
	}()
	select {
	case obj = <-ch:
		return obj, nil
	case err = <-errchan:
		return nil, &jserror{err}
	}
}
