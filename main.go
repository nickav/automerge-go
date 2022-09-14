package main

import (
	"fmt"
	"time"
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
}

type Map struct {
	handle *C.AMobjId
	root   Doc
}

type List struct {
	handle *C.AMobjId
	root   Doc
}

// @Incomplete: support actor id
func Init() Doc {
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

type ChangeFn func(root Map)

// @Incomplete: I _think_ this is all the "change" function does. But we should confirm that.
func (doc Doc) Change(msg string, fn ChangeFn) {
	fn(doc.Root())
	doc.Commit(msg)
}

func (doc Doc) GetActorID() string {
	return C.GoString(C.AMG_GetActorIDString(doc.handle))
}

func MapPut(doc Doc, objId *C.AMobjId, key string, value interface{}) {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

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

func ListSet(doc Doc, objId *C.AMobjId, index uint64, insert bool, value interface{}) {
	switch value.(type) {
	case string:
		value_cstr := C.CString(value.(string))
		defer C.free(unsafe.Pointer(value_cstr))
		C.AMfree(C.AMlistPutStr(doc.handle, objId, C.ulonglong(index), C._Bool(insert), value_cstr))
	case int:
		C.AMfree(C.AMlistPutInt(doc.handle, objId, C.ulonglong(index), C._Bool(insert), C.longlong(value.(int))))
	}
}

func MapPutObject(doc Doc, objId *C.AMobjId, key string, objType ObjectType) Object {
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

/*

func ListPutObject(doc Doc, objId *C.AMobjId, key string, objType C.uchar) *C.AMobjId {
	cstr := C.CString(key)
	defer C.free(unsafe.Pointer(cstr))

	result := C.toAMobjId(C.AMresultValue(C.AMlistPutObject(doc.handle, objId, cstr, objType)))
	return result
}
*/

func (doc Doc) Root() Map {
	result := Map{}
	result.root = doc
	result.handle = nil
	return result
}

func (doc Doc) Set(key string, value interface{}) {
	MapPut(doc, doc.Root().handle, key, value)
}

func (obj Map) Set(key string, value interface{}) {
	MapPut(obj.root, obj.handle, key, value)
}

func (doc Doc) PutList(at Map, key string) List {
	result := List{}
	object := MapPutObject(doc, at.handle, key, ObjectType_List)
	result.handle = object.handle
	result.root = doc
	return result
}

func (list List) Insert(index uint64, value interface{}) {
	ListSet(list.root, list.handle, index, true, value)
}

func (list List) Set(index uint64, value interface{}) {
	ListSet(list.root, list.handle, index, false, value)
}

/*
func (doc Doc) PutMap(at Map, key string) Map {
}
*/

/*
func (doc Doc) NewMap(key string) Map {
	result := Map{}
	result.handle = MapPutObject(doc, nil, key, C.AM_OBJ_TYPE_MAP)
	result.root = doc
	return result
}

func (doc Doc) NewList(key string) List {
	result := List{}
	result.handle = MapPutObject(doc, nil, key, C.AM_OBJ_TYPE_LIST)
	result.root = doc
	return result
}
*/

func toCBool(x bool) C.int {
	if x {
		return C.int(1)
	}
	return C.int(0)
}

func resultToString(byteSpan C.AMbyteSpan) string {
	return C.GoStringN((*C.char)(unsafe.Pointer(byteSpan.src)), C.int(byteSpan.count))
}

func main() {
	fmt.Println("hello!")

	doc1 := Init()

	doc1.Set("hello", "world")

	list := doc1.PutList(doc1.Root(), "list")
	list.Insert(0, "a")

	const BENCH_TIMES = 10_000
	start := time.Now()

	for i := 0; i < 10_000; i++ {
		Load(doc1.Save())
	}

	end := time.Now()
	fmt.Println("Took", end.Sub(start), "(average: ", end.Sub(start)/time.Duration(BENCH_TIMES), ")")

	/*
		cards := doc1.NewMap("cards")
		doc1.Set("cards", true)
	*/

	fmt.Println(doc1)

	fmt.Println("doc1 ID:", doc1.GetActorID())
	fmt.Println("Save:", doc1.Save())

	str := doc1.Save()
	doc2 := Load(str)
	fmt.Println("doc2:", doc2)
	fmt.Println("doc2 ID:", doc2.GetActorID())
}
