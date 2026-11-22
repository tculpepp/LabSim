//go:build js && wasm

package save

import (
	"errors"
	"fmt"
	"syscall/js"
)

// LocalStore saves to the browser's localStorage. Access can throw (private
// mode, blocked storage, quota), so every call recovers and returns an error.
type LocalStore struct{ Key string }

func DefaultStore() Store { return LocalStore{Key: "labsim-save"} }

func (s LocalStore) call(method string, args ...any) (v js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("localStorage %s: %v", method, r)
		}
	}()
	ls := js.Global().Get("localStorage")
	if ls.IsUndefined() || ls.IsNull() {
		return js.Value{}, errors.New("localStorage unavailable")
	}
	return ls.Call(method, args...), nil
}

func (s LocalStore) Load() ([]byte, error) {
	v, err := s.call("getItem", s.Key)
	if err != nil {
		return nil, err
	}
	if v.IsNull() || v.IsUndefined() {
		return nil, ErrNoSave
	}
	return []byte(v.String()), nil
}

func (s LocalStore) Save(b []byte) error {
	_, err := s.call("setItem", s.Key, string(b))
	return err
}

func (s LocalStore) Clear() error {
	_, err := s.call("removeItem", s.Key)
	return err
}

func (s LocalStore) Backup(b []byte) error {
	_, err := s.call("setItem", s.Key+"-corrupt", string(b))
	return err
}
