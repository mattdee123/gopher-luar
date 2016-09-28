package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func chanIndex(L *lua.LState) int {
	_, _, mt, isPtr := check(L, 1, reflect.Chan)
	key := L.CheckString(2)

	if !isPtr {
		if fn := mt.method(key); fn != nil {
			L.Push(fn)
			return 1
		}
	}

	if fn := mt.ptrMethod(key); fn != nil {
		L.Push(fn)
		return 1
	}

	return 0
}

func chanLen(L *lua.LState) int {
	ref, _, _, isPtr := check(L, 1, reflect.Chan)
	if isPtr {
		L.RaiseError("invalid operation on chan pointer")
	}
	L.Push(lua.LNumber(ref.Len()))
	return 1
}

// chan methods

func chanSend(L *lua.LState) int {
	ref, _, _, _ := check(L, 1, reflect.Chan)
	value := L.CheckAny(2)
	convertedValue := lValueToReflect(L, value, ref.Type().Elem(), nil)
	if convertedValue.Type() != ref.Type().Elem() {
		L.ArgError(2, "incorrect type")
	}
	ref.Send(convertedValue)
	return 0
}

func chanReceive(L *lua.LState) int {
	ref, _, _, _ := check(L, 1, reflect.Chan)

	value, ok := ref.Recv()
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LBool(false))
		return 2
	}
	L.Push(New(L, value.Interface()))
	L.Push(lua.LBool(true))
	return 2
}

func chanClose(L *lua.LState) int {
	ref, opts, _, _ := check(L, 1, reflect.Chan)

	if opts.Immutable {
		L.RaiseError("cannot close immutable channel")
	}

	ref.Close()
	return 0
}
