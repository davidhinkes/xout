// Binary plugin is the so entry point for XPlane.  The intended product
// of this is a dll/so.
package main

import(
	// #cgo darwin CFLAGS: -I ${SRCDIR}/XPSDK213/CHeaders -DAPL
	// #cgo darwin LDFLAGS: -F ${SRCDIR}/XPSDK213/Libraries/Mac -framework XPLM
	/*
	#include <XPLM/XPLMDataAccess.h>
	#include <XPLM/XPLMPlugin.h>
	*/
	"C"
	"unsafe"

	"github.com/davidhinkes/xout/internal/cutils"
)

func main() {
	// An emtpy main func is needed to make shared libraries.
}

//export XPluginStart
func XPluginStart(outName, outSig, outDesc *C.char) C.int{
	cutils.Strcpy(unsafe.Pointer(outName), "xout")
	cutils.Strcpy(unsafe.Pointer(outSig), "xout")
	cutils.Strcpy(unsafe.Pointer(outDesc), "xout")
	return C.int(1)
}

//export XPluginStop
func XPluginStop() {
}

//export XPluginDisable
func XPluginDisable() {
}

//export XPluginEnable
func XPluginEnable() C.int {
	return C.int(1)
}

//export XPluginReceiveMessage
func XPluginReceiveMessage(inFromWho, inMessage C.int, inParam unsafe.Pointer) {
}
