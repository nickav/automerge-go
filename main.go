package main

//
// TODO(nick):
// - fix memory leaks everywhere
// - get history / changes
// - support counters
// - JSON serialize?
//

// API examples:

// https://github.com/progrium/goja-automerge/blob/main/automerge_test.go
// https://github.com/automerge/automerge-rs/blob/main/automerge-c/examples/quickstart.c
// https://github.com/automerge/automerge-rs/blob/c2ed212dbccafd66c85cd0bb9527bb27f81b6e17/automerge-c/test/ported_wasm/basic_tests.c
// https://github.com/automerge/automerge-rs/blob/main/automerge/examples/quickstart.rs

import (
	"fmt"
	//"time"
	"unsafe"
)

/*
#include "main.h"
*/
import "C"

//
// Types
//

type Doc struct {
	handle *C.AMdoc
}

type ObjectType int

const (
	Object_Map  ObjectType = 0
	Object_List            = 1
)

type Object struct {
	handle *C.AMobjId
	//Type   ObjectType
}

type Map struct {
	handle Object
	root   Doc
}

type List struct {
	handle Object
	root   Doc
}

type Value struct {
	handle C.AMvalue
	root   Doc
}

type EmptyList struct {
}

type EmptyMap struct {
}

//
// Doc
//

// @Incomplete: support actor id
func New() Doc {
	result := Doc{}
	result.handle = C.AMG_Create(nil)
	return result
}

// @Incomplete: support timestamp
func (doc Doc) Commit(msg string) {
	if len(msg) > 0 {
		cstr := C.CString(msg)
		C.AMG_Commit(doc.handle, cstr, nil)
		C.free(unsafe.Pointer(cstr))
	} else {
		C.AMG_Commit(doc.handle, nil, nil)
	}
}

func (doc Doc) Save() string {
	byteSpan := C.AMG_Save(doc.handle)
	return resultToString(byteSpan)
}

func Load(raw string) Doc {
	result := Doc{}

	bytes := C.CString(raw)
	result.handle = C.AMG_Load(bytes, C.ulonglong(len(raw)))
	C.free(unsafe.Pointer(bytes))

	return result
}

func Merge(dest Doc, src Doc) {
	C.AMG_Merge(dest.handle, src.handle)
}

func Clone(doc Doc) Doc {
	result := Doc{}
	result.handle = C.AMG_Clone(doc.handle)
	return result
}

type ChangeFn func(root Map)

// @Incomplete: I _think_ this is all the "change" function does. But we should confirm that.
func (doc Doc) Change(msg string, fn ChangeFn) {
	fn(doc.Root())
	doc.Commit(msg)
}

func (doc Doc) GetActorID() string {
	return C.GoString(C.AMG_GetActorIDString(doc.handle))
}

func (doc Doc) PutObject(obj Object, key string, objType ObjectType) Object {
	return MapPutObject(doc, obj, key, objType)
}

func (doc Doc) InsertObject(obj Object, index uint64, insert bool, objType ObjectType) Object {
	return ListPutObject(doc, obj, index, true, objType)
}

func (doc Doc) SetObject(obj Object, index uint64, objType ObjectType) Object {
	return ListPutObject(doc, obj, index, false, objType)
}

func (doc Doc) Root() Map {
	result := Map{}
	result.root = doc
	result.handle = Object{} // NOTE(nick): nil for root
	return result
}

func (doc Doc) Set(key string, value interface{}) interface{} {
	return doc.Root().Set(key, value)
}

func (doc Doc) Get(key string) Value {
	return doc.Root().Get(key)
}

//
// Object
//

func MapPutObject(doc Doc, obj Object, key string, objType ObjectType) Object {
	objId := obj.handle

	var actualType C.uchar
	if objType == Object_List {
		actualType = C.AM_OBJ_TYPE_LIST
	} else {
		actualType = C.AM_OBJ_TYPE_MAP
	}

	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	result := Object{}
	result.handle = C.AMG_ToObjectID(C.AMresultValue(C.AMmapPutObject(doc.handle, objId, cstr, actualType)))
	return result
}

func ListPutObject(doc Doc, obj Object, index uint64, insert bool, objType ObjectType) Object {
	objId := obj.handle

	var actualType C.uchar
	if objType == Object_List {
		actualType = C.AM_OBJ_TYPE_LIST
	} else {
		actualType = C.AM_OBJ_TYPE_MAP
	}

	result := Object{}
	result.handle = C.AMG_ToObjectID(C.AMresultValue(C.AMlistPutObject(doc.handle, objId, C.ulonglong(index), C._Bool(insert), actualType)))
	return result
}

func MapGet(doc Doc, obj Object, key string) Value {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	result := Value{}
	result.handle = C.AMG_MapGet(doc.handle, obj.handle, cstr)
	result.root = doc
	return result
}

func ListGet(doc Doc, obj Object, index uint64) Value {
	result := Value{}
	result.handle = C.AMG_ListGet(doc.handle, obj.handle, C.ulonglong(index))
	result.root = doc
	return result
}

func MapPut(doc Doc, obj Object, key string, value interface{}) Value {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	objId := obj.handle

	switch value.(type) {
	case string:
		value_cstr := C.CString(value.(string))
		defer C.free(unsafe.Pointer(value_cstr))
		C.AMfree(C.AMmapPutStr(doc.handle, objId, cstr, value_cstr))
	case int:
		C.AMfree(C.AMmapPutInt(doc.handle, objId, cstr, C.longlong(value.(int))))
	case uint64:
		C.AMfree(C.AMmapPutInt(doc.handle, objId, cstr, C.longlong(value.(uint64))))
	case float64:
		C.AMfree(C.AMmapPutF64(doc.handle, objId, cstr, C.double(value.(float64))))
	case bool:
		C.AMfree(C.AMmapPutBool(doc.handle, obj.handle, cstr, C._Bool(value.(bool))))
	case EmptyList:
		MapPutObject(doc, obj, key, Object_List)
	case EmptyMap:
		MapPutObject(doc, obj, key, Object_Map)
	default:
		fmt.Println("[automerge] MapPut - unhandled case!")
	}

	result := Value{}
	result.root = doc
	result.handle = MapGet(doc, obj, key).handle
	return result
}

func ListSet(doc Doc, obj Object, index uint64, insert bool, value interface{}) Value {
	objId := obj.handle

	switch value.(type) {
	case string:
		value_cstr := C.CString(value.(string))
		defer C.free(unsafe.Pointer(value_cstr))
		C.AMfree(C.AMlistPutStr(doc.handle, objId, C.ulonglong(index), C._Bool(insert), value_cstr))
	case int:
		C.AMfree(C.AMlistPutInt(doc.handle, objId, C.ulonglong(index), C._Bool(insert), C.longlong(value.(int))))
	case uint64:
		C.AMfree(C.AMlistPutInt(doc.handle, objId, C.ulonglong(index), C._Bool(insert), C.longlong(value.(uint64))))
	case float64:
		C.AMfree(C.AMlistPutF64(doc.handle, objId, C.ulonglong(index), C._Bool(insert), C.double(value.(float64))))
	case bool:
		C.AMfree(C.AMlistPutBool(doc.handle, objId, C.ulonglong(index), C._Bool(insert), C._Bool(value.(bool))))
	case EmptyList:
		ListPutObject(doc, obj, index, insert, Object_List)
	case EmptyMap:
		ListPutObject(doc, obj, index, insert, Object_Map)
	default:
		fmt.Println("[automerge] ListSet - unhandled case!")
	}

	result := Value{}
	result.root = doc
	result.handle = ListGet(doc, obj, index).handle
	return result
}

// Returns "size" of object (for lists this is length, for maps this is number of keys?)
func GetSize(doc Doc, obj Object) uint64 {
	return uint64(C.AMG_GetSize(doc.handle, obj.handle))
}

//
// Map
//

func (m Map) Set(key string, value interface{}) Value {
	return MapPut(m.root, m.handle, key, value)
}

func (m Map) Get(key string) Value {
	return MapGet(m.root, m.handle, key)
}

func (m Map) Count() uint64 {
	return GetSize(m.root, m.handle)
}

func (m Map) Delete(key string) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	C.AMG_MapDelete(m.root.handle, m.handle.handle, cstr)
}

func (m Map) Keys() []string {
	amhandle := C.AMkeys(m.root.handle, m.handle.handle, nil)
	strs := C.AMG_ToStrs(C.AMresultValue(amhandle))

	result := []string{}
	for str := C.AMstrsNext(&strs, 1); str != nil; str = C.AMstrsNext(&strs, 1) {
		result = append(result, C.GoString(str))
	}

	C.AMfree(amhandle)
	return result
}

//
// List
//

func (list List) Insert(index uint64, value interface{}) Value {
	return ListSet(list.root, list.handle, index, true, value)
}

func (list List) Set(index uint64, value interface{}) Value {
	return ListSet(list.root, list.handle, index, false, value)
}

func (list List) Push(value interface{}) Value {
	return ListSet(list.root, list.handle, C.SIZE_MAX, true, value)
}

func (list List) Get(index uint64) Value {
	return ListGet(list.root, list.handle, index)
}

func (list List) Count() uint64 {
	return GetSize(list.root, list.handle)
}

func (list List) Delete(index uint64) {
	C.AMG_ListDelete(list.root.handle, list.handle.handle, C.ulonglong(index))
}

func (list List) Pop() interface{} {
	value := list.Get(C.SIZE_MAX)
	list.Delete(C.SIZE_MAX)
	return value.Value()
}

//
// Value
//

// @Incomplete: do we want to define our own ValueType?
func (v Value) Type() uint8 {
	tag := C.AMG_GetType(v.handle)
	return uint8(tag)
}

func (v Value) Value() interface{} {
	tag := C.AMG_GetType(v.handle)

	switch tag {
	case C.AM_VALUE_BOOLEAN:
		return bool(fromCBool(C.AMG_ToBool(v.handle)))
	case C.AM_VALUE_ACTOR_ID:
		return string(C.GoString(C.AMG_ActorIDToString(C.AMG_ToActorID(v.handle))))
	case C.AM_VALUE_INT:
		return int64(C.AMG_ToInt(v.handle))
	case C.AM_VALUE_F64:
		return float64(C.AMG_ToF64(v.handle))
	case C.AM_VALUE_NULL:
		return nil
	case C.AM_VALUE_STR:
		return string(C.GoString(C.AMG_ToString(v.handle)))
	case C.AM_VALUE_UINT:
		return uint64(C.AMG_ToUint(v.handle))
	}

	return nil
}

func (v Value) ToList() List {
	tag := C.AMG_GetType(v.handle)
	// @Incomplete: check if object is list
	result := List{}
	if tag == C.AM_VALUE_OBJ_ID {
		result.handle.handle = C.AMG_ToObjectID(v.handle)
		result.root = v.root
	}
	return result
}

func (v Value) ToMap() Map {
	tag := C.AMG_GetType(v.handle)
	// @Incomplete: check if object is map
	result := Map{}
	if tag == C.AM_VALUE_OBJ_ID {
		result.handle.handle = C.AMG_ToObjectID(v.handle)
		result.root = v.root
	}
	return result
}

func (doc Doc) List() EmptyList {
	return EmptyList{}
}

func (doc Doc) Map() EmptyMap {
	return EmptyMap{}
}

//
// Helpers
//

func toCBool(x bool) C._Bool {
	if x {
		return C._Bool(true)
	}
	return C._Bool(false)
}

func fromCBool(x C.int) bool {
	if x == 0 {
		return false
	}
	return true
}

func resultToString(byteSpan C.AMbyteSpan) string {
	return C.GoStringN((*C.char)(unsafe.Pointer(byteSpan.src)), C.int(byteSpan.count))
}

func main() {
	doc1 := New()
	root := doc1.Root()
	root.Set("hello", "world")
	root.Set("foo", 42)
	root.Set("bar", 23)
	root.Delete("bar")

	mappy := doc1.Root().Set("mappy", doc1.Map()).ToMap()
	mappy.Set("a", true)

	mappy2 := doc1.Root().Get("mappy").ToMap()
	mappy2.Set("b", 123123)

	cards := doc1.Root().Set("cards", doc1.List()).ToList()
	cards.Push("a")
	cards.Push("b")
	cards.Push("c")
	cards.Pop()

	doc1.Root().Get("cards").ToList().Push("d")

	fmt.Println("hello:", doc1.Root().Get("hello").Value())
	fmt.Println("foo:", doc1.Root().Get("foo").Value())

	fmt.Println("cards[0]:", cards.Get(0).Value())
	fmt.Println("cards[1]:", cards.Get(1).Value())
	fmt.Println("cards[2]:", cards.Get(2).Value())
	fmt.Println("cards[3]:", cards.Get(3).Value())
	fmt.Println("cards.Count():", cards.Count())

	doc1.Change("add cards2", func(root Map) {
		root.Set("cards2", doc1.List())
		root.Get("cards2").ToList().Push(true)
	})

	fmt.Println("doc1.Root().Keys():", doc1.Root().Keys())

	fmt.Println("mappy.Keys():", mappy.Keys())
	fmt.Println("mappy2.Keys():", mappy2.Keys())

	fmt.Println("doc1 ID:", doc1.GetActorID())
	fmt.Println("Save:", doc1.Save())

	str := doc1.Save()
	doc2 := Load(str)
	fmt.Println("doc2:", doc2)
	fmt.Println("doc2 ID:", doc2.GetActorID())
}
