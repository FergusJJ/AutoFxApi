package main

// #cgo CFLAGS: -I/usr/include/python3.8
// #cgo LDFLAGS: -lpython3.10
// #include <Python.h>
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	C.Py_Initialize()
	defer C.Py_Finalize()

	pyName := C.CString("monitor_module")
	defer C.free(unsafe.Pointer(pyName))
	pyModule := C.PyImport_ImportModule(pyName)
	if pyModule == nil {
		fmt.Println("pyModule is nilk")
	}

	pyFunc := C.PyObject_GetAttrString(pyModule, C.CString("monitor_function"))
	pyArgs := C.PyTuple_New(0)
	pyResult := C.PyObject_CallObject(pyFunc, pyArgs)

	result := C.GoString(C.PyUnicode_AsUTF8(pyResult)) // Python result to a Go string
	fmt.Println(result)
}
