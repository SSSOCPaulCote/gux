# gux
A redux-like state management library written in Go.

# Usage

Your state can be as simple as an `int` or a more complex `struct`. The `Reducer` function interface can take any type. It is up to you to create a reducer that does type assertion and for you to write code handling incorrect state types when getting the state.

```go
import (
    "log"
    "time"

    "github.com/SSSOCPaulCote/gux"
)

type (
    ExampleState struct {
        Counter int
    }
)

var (
    initialState = ExampleState{}
    rootReducer gux.Reducer = func(s interface{}, a gux.Action) (interface{}, error) {
        // assert type of s
		oldState, ok := s.(ExampleState)
		if !ok {
			return nil, gux.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "counter/increment":
			// assert type of payload
			newCnt, ok := a.Payload.(int)
			if !ok {
				return nil, gux.ErrInvalidPayloadType
			}
			oldState.Counter = oldState.Counter+newCnt
			return oldState, nil
		case "counter/decrement":
			// assert type of payload
			newCnt, ok := a.Payload.(int)
			if !ok {
				return nil, gux.ErrInvalidPayloadType
			}
			oldState.Counter = oldState.Counter-newCnt
			return oldState, nil
		default:
			return nil, gux.ErrInvalidAction
		}
    }
    incrementAction = gux.Action{
        Type: "counter/increment",
        Payload: 1,
    }
    decrementAction = gux.Action{
        Type: "counter/decrement",
        Payload: 1,
    }
)

func main() {
    store := gux.CreateStore(initialState, rootReducer)
    quitChan := make(chan struct{})
    defer close(quitChan)
    go func() {
        for i := 0; i < 5; i++ {
            select {
            case <-quitChan:
                return
            default:
                err := store.Dispatch(incrementAction)
                if err != nil {
                    log.Println(err)
                }
                time.Sleep(1 * time.Second)
            }
        }
    }()
    updateChan, unsub := store.Subscribe("name")
    defer unsub(store, "name")
    for {
        select {
        case <-updateChan:
            state := store.GetState()
            cState, ok := state.(ExampleState)
            if !ok {
                log.Println(gux.ErrInvalidStateType)
                return
            }
            log.Printf("Counter: %v\n", cState.Counter)
            if cState.Counter == 3 {
                log.Println("End")
                return
            }
        }
    }
}
```
