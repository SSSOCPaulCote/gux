/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/11
*/

package gux

import (
	"sync"

	bg "github.com/SSSOCPaulCote/blunderguard"
)

const (
	ErrInvalidPayloadType = bg.Error("Invalid payload type")
	ErrInvalidStateType   = bg.Error("Invalid state type")
	ErrInvalidAction      = bg.Error("Invalid action")
	ErrAlreadySubscribed  = bg.Error("subscriber with given name already subscribed")
)

type (
	Reducer func(interface{}, Action) (interface{}, error)

	Action struct {
		Type    string
		Payload interface{}
	}

	Listener struct {
		IsConnected bool
		Signal      chan struct{}
	}

	Store struct {
		state     interface{}
		reducer   Reducer
		listeners map[string]*Listener
		sync.RWMutex
	}
)

// CreateStore creates a new state store object
func CreateStore(initialState interface{}, rootReducer Reducer) *Store {
	return &Store{
		state:     initialState,
		reducer:   rootReducer,
		listeners: make(map[string]*Listener),
	}
}

// GetState returns the current state object
func (s *Store) GetState() interface{} {
	s.RLock()
	defer s.RUnlock()
	return s.state
}

// Dispatch takes an action and returns an error. It is the only way to change the state
func (s *Store) Dispatch(action Action) error {
	s.Lock()
	defer s.Unlock()
	newState, err := s.reducer(s.state, action)
	if err != nil {
		return err
	}
	s.state = newState
	// Remove disconnected listeners
	// Update subscribers on state change
	newListenerMap := make(map[string]*Listener)
	for n, l := range s.listeners {
		if !l.IsConnected {
			close(l.Signal)
			continue
		}
		l.Signal <- struct{}{}
		newListenerMap[n] = l
	}
	s.listeners = newListenerMap
	return nil
}

// Subscribe adds a callback function to the list of listeners which will be executed upon each Dispatch call.
// Returns the index in the listener slice belonging to callback and unsubscribe function
func (s *Store) Subscribe(name string) (chan struct{}, func(), error) {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.listeners[name]; ok {
		return nil, nil, ErrAlreadySubscribed
	}
	s.listeners[name] = &Listener{IsConnected: true, Signal: make(chan struct{}, 2)} // made channel buffered for edge case where unsub() and l.Signal<-struct{}{} listener disconnects, it won't hang
	unsub := func() {
		s.Lock()
		defer s.Unlock()
		s.listeners[name].IsConnected = false
	}
	return s.listeners[name].Signal, unsub, nil
}

// CombineReducers combines any number of reducers and returns one combined reducer
func CombineReducers(reducers ...Reducer) Reducer {
	var combinedReducer Reducer = func(s interface{}, a Action) (interface{}, error) {
		newState := make([]interface{}, len(reducers))
		for i, r := range reducers {
			newS, err := r(s, a)
			if err != nil {
				return nil, err
			}
			newState[i] = newS
		}
		return newState, nil
	}
	return combinedReducer
}
