package luar

import (
	"reflect"

	"github.com/yuin/gopher-lua"
)

func sliceIndex(L *lua.LState) int {
	ref, opts, mt, isPtr := check(L, 1, reflect.Slice)
	ref = reflect.Indirect(ref)
	key := L.CheckAny(2)

	switch converted := key.(type) {
	case lua.LNumber:
		index := int(converted)
		if index < 1 || index > ref.Len() {
			L.ArgError(2, "index out of range")
		}
		val := ref.Index(index - 1)
		if (val.Kind() == reflect.Struct || val.Kind() == reflect.Array) && val.CanAddr() {
			val = val.Addr()
		}
		L.Push(New(L, val.Interface(), opts))
	case lua.LString:
		if !isPtr {
			if fn := mt.method(string(converted)); fn != nil {
				L.Push(fn)
				return 1
			}
		}
		if fn := mt.ptrMethod(string(converted)); fn != nil {
			L.Push(fn)
			return 1
		}
		return 0
	default:
		L.ArgError(2, "must be a number or string")
	}
	return 1
}

func sliceNewIndex(L *lua.LState) int {
	ref, opts, _, isPtr := check(L, 1, reflect.Slice)
	index := L.CheckInt(2)
	value := L.CheckAny(3)

	if isPtr {
		L.RaiseError("invalid operation on slice pointer")
	}

	if opts.Immutable {
		L.RaiseError("invalid operation on immutable slice")
	}

	if index < 1 || index > ref.Len() {
		L.ArgError(2, "index out of range")
	}
	ref.Index(index - 1).Set(lValueToReflect(L, value, ref.Type().Elem(), nil))
	return 0
}

func sliceLen(L *lua.LState) int {
	ref, _, _, isPtr := check(L, 1, reflect.Slice)

	if isPtr {
		L.RaiseError("invalid operation on slice pointer")
	}

	L.Push(lua.LNumber(ref.Len()))
	return 1
}

func sliceCall(L *lua.LState) int {
	ref, opts, _, isPtr := check(L, 1, reflect.Slice)
	if isPtr {
		L.RaiseError("invalid operation on slice pointer")
	}

	i := 0
	fn := func(L *lua.LState) int {
		if i >= ref.Len() {
			return 0
		}
		item := ref.Index(i).Interface()
		L.Push(lua.LNumber(i + 1))
		L.Push(New(L, item, opts))
		i++
		return 2
	}

	L.Push(L.NewFunction(fn))
	return 1
}

func sliceEq(L *lua.LState) int {
	ref1, _, _, isPtr1 := check(L, 1, reflect.Slice)
	ref2, _, _, isPtr2 := check(L, 2, reflect.Slice)

	if isPtr1 && isPtr2 {
		L.Push(lua.LBool(ref1.Pointer() == ref2.Pointer()))
		return 1
	}

	L.RaiseError("invalid operation == on slice")
	return 0 // never reaches
}

// slice methods

func sliceCapacity(L *lua.LState) int {
	ref, _, _, _ := check(L, 1, reflect.Slice)
	L.Push(lua.LNumber(ref.Cap()))
	return 1
}

func sliceAppend(L *lua.LState) int {
	ref, opts, _, isPtr := check(L, 1, reflect.Slice)

	if isPtr {
		L.RaiseError("invalid operation on slice pointer")
	}

	if opts.Immutable {
		L.RaiseError("invalid operation on immutable slice")
	}

	hint := ref.Type().Elem()
	values := make([]reflect.Value, L.GetTop()-1)
	for i := 2; i <= L.GetTop(); i++ {
		value := lValueToReflect(L, L.Get(i), hint, nil)
		if value.Type() != hint {
			L.ArgError(i, "invalid type")
		}
		values[i-2] = value
	}

	newSlice := reflect.Append(ref, values...)
	L.Push(New(L, newSlice.Interface(), opts))
	return 1
}
