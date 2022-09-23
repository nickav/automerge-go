package main

//
// TODO(nick):
// - finish / clean up API weirdness
// - memory situation
// - get changes
// - list iterate over values, delete
// - object get, iterate, delete
// - counters?
// - hashes
//

import (
	"fmt"
	//"time"
	"unsafe"
)

/*
#include "main.h"
*/
import "C"

type Doc struct {
	handle *C.AMdoc
}

type ObjectType int

const (
	ObjectType_Map  ObjectType = 0
	ObjectType_List            = 1
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
}

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

func MapPut(doc Doc, obj Object, key string, value interface{}) {
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
		/*
			case bool:
				C.AMfree(C.AMmapPutBool(doc.handle, obj.handle, cstr, toCBool(value.(bool))))
		*/
	}
}

func ListSet(doc Doc, obj Object, index uint64, insert bool, value interface{}) {
	objId := obj.handle

	switch value.(type) {
	case string:
		value_cstr := C.CString(value.(string))
		defer C.free(unsafe.Pointer(value_cstr))
		C.AMfree(C.AMlistPutStr(doc.handle, objId, C.ulonglong(index), C._Bool(insert), value_cstr))
	case int:
		C.AMfree(C.AMlistPutInt(doc.handle, objId, C.ulonglong(index), C._Bool(insert), C.longlong(value.(int))))
	}
}

func MapPutObject(doc Doc, obj Object, key string, objType ObjectType) Object {
	objId := obj.handle

	var actualType C.uchar
	if objType == ObjectType_List {
		actualType = C.AM_OBJ_TYPE_LIST
	} else {
		actualType = C.AM_OBJ_TYPE_MAP
	}

	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	result := Object{}
	result.handle = C.toAMobjId(C.AMresultValue(C.AMmapPutObject(doc.handle, objId, cstr, actualType)))
	return result
}

func ListPutObject(doc Doc, obj Object, index int64, insert bool, objType ObjectType) Object {
	objId := obj.handle

	var actualType C.uchar
	if objType == ObjectType_List {
		actualType = C.AM_OBJ_TYPE_LIST
	} else {
		actualType = C.AM_OBJ_TYPE_MAP
	}

	result := Object{}
	result.handle = C.toAMobjId(C.AMresultValue(C.AMlistPutObject(doc.handle, objId, C.ulonglong(index), C._Bool(insert), actualType)))
	return result
}

/*
func MapPutMap(m Map, key string) Map {
	result := Map{}
	result.handle = MapPutObject(m.root, m.handle, key, ObjectType_Map)
	result.root = m.root
	return result
}

func MapPutList(m Map, key string) List {
	result := List{}
	result.handle = MapPutObject(m.root, m.handle, key, ObjectType_List)
	result.root = m.root
	return result
}
*/

func (doc Doc) Root() Map {
	result := Map{}
	result.root = doc
	result.handle = Object{} // nil for root
	return result
}

func (m Map) NewMap(key string) Map {
	result := Map{}
	result.handle = MapPutObject(m.root, m.handle, key, ObjectType_Map)
	result.root = m.root
	return result
}

func (m Map) Set(key string, value interface{}) {
	MapPut(m.root, m.handle, key, value)
}

func (m Map) Get(key string) Value {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	result := Value{}
	result.handle = C.AMG_MapGet(m.root.handle, m.handle.handle, cstr)
	return result
}

func (m Map) Count() uint64 {
	return uint64(C.AMG_GetSize(m.root.handle, m.handle.handle))
}

func (m Map) Remove(key string) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	C.AMG_MapDelete(m.root.handle, m.handle.handle, cstr)
}

func (m Map) NewList(key string) List {
	result := List{}
	result.handle = MapPutObject(m.root, m.handle, key, ObjectType_List)
	result.root = m.root
	return result
}

func (list List) Insert(index uint64, value interface{}) {
	ListSet(list.root, list.handle, index, true, value)
}

func (list List) Set(index uint64, value interface{}) {
	ListSet(list.root, list.handle, index, false, value)
}

func (list List) Push(value interface{}) {
	ListSet(list.root, list.handle, C.SIZE_MAX, true, value)
}

func (list List) Get(index uint64) Value {
	result := Value{}
	result.handle = C.AMG_ListGet(list.root.handle, list.handle.handle, C.ulonglong(index))
	return result
}

func (list List) Count() uint64 {
	return uint64(C.AMG_GetSize(list.root.handle, list.handle.handle))
}

func (list List) Remove(index uint64) {
	C.AMG_ListDelete(list.root.handle, list.handle.handle, C.ulonglong(index))
}

func (list List) Pop() interface{} {
	value := list.Get(C.SIZE_MAX)

	list.Remove(C.SIZE_MAX)

	return value.Value()
}

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

func (doc Doc) NewList(key string) List {
	return doc.Root().NewList(key)
}

func (doc Doc) NewMap(key string) Map {
	return doc.Root().NewMap(key)
}

func (doc Doc) Set(key string, value interface{}) {
	doc.Root().Set(key, value)
}

func (doc Doc) Get(key string) interface{} {
	return doc.Root().Get(key)
}

func toCBool(x bool) C.int {
	if x {
		return C.int(1)
	}
	return C.int(0)
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
	fmt.Println("hello!")

	doc1 := New()
	doc1.Root().Set("hello", "world")
	doc1.Root().Set("foo", 42)
	doc1.Root().Set("bar", 23)
	cards := doc1.Root().NewList("cards")
	cards.Push("a")
	cards.Push("b")
	cards.Push("c")

	fmt.Println("hello:", doc1.Root().Get("hello").Value())
	fmt.Println("foo:", doc1.Root().Get("foo").Value())

	fmt.Println("cards[0]:", cards.Get(0).Value())
	fmt.Println("cards[1]:", cards.Get(1).Value())
	fmt.Println("cards[2]:", cards.Get(2).Value())
	fmt.Println("cards[3]:", cards.Get(3).Value())
	fmt.Println("cards.Count():", cards.Count())

	/*
		doc1.Change("add cards", func(doc Map) {
			//doc.Set("cards", doc.NewList())
			doc.NewList("cards")
		})

		list := doc1.PutList(doc1.Root(), "list")
		list.Insert(0, "a")
	*/

	/*
		const BENCH_TIMES = 10_000
		start := time.Now()

		for i := 0; i < 10_000; i++ {
			Load(doc1.Save())
		}

		end := time.Now()
		fmt.Println("Took", end.Sub(start), "(average:", end.Sub(start)/time.Duration(BENCH_TIMES), ")")

		fmt.Println(doc1)
	*/

	fmt.Println("doc1 ID:", doc1.GetActorID())
	fmt.Println("Save:", doc1.Save())

	str := doc1.Save()
	doc2 := Load(str)
	fmt.Println("doc2:", doc2)
	fmt.Println("doc2 ID:", doc2.GetActorID())
}
