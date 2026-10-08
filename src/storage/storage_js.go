//go:build js

package storage

import (
	"fmt"
	"io/fs"
	"syscall/js"
)

// keyPrefix namespaces the game's keys, since localStorage is shared by
// everything served from the same origin.
const keyPrefix = "dungeoneer/"

// call invokes a localStorage method. syscall/js panics when JavaScript
// throws (quota exceeded, storage blocked in a sandboxed iframe or private
// window), so the panic is turned into an error and the game carries on as
// if there were no save.
func call(method string, args ...any) (v js.Value, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("storage: localStorage.%s: %v", method, r)
		}
	}()
	return js.Global().Get("localStorage").Call(method, args...), nil
}

// ReadFile returns the save called name.
func ReadFile(name string) ([]byte, error) {
	v, err := call("getItem", keyPrefix+name)
	if err != nil {
		return nil, err
	}
	if v.IsNull() {
		return nil, &fs.PathError{Op: "read", Path: name, Err: fs.ErrNotExist}
	}
	return []byte(v.String()), nil
}

// WriteFile stores data under name.
func WriteFile(name string, data []byte) error {
	_, err := call("setItem", keyPrefix+name, string(data))
	return err
}

// Remove deletes the save called name. Removing a name that does not exist
// is not an error.
func Remove(name string) error {
	_, err := call("removeItem", keyPrefix+name)
	return err
}
