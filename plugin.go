// Binary plugin is the so entry point for XPlane.  The intended product
// of this is a dll/so.
package main

import(
	"C"
	"unsafe"
)

func main() {
	// An emtpy main func is needed to make shared libraries.
}

//export XPluginStart
func XPluginStart(outName, outSig, outDesc *C.char) C.int{
	return C.int(1)
}

//export XPluginStop
func XPluginStop() {
}

//export XPluginDisable
func XPluginDisable() {
}

//export XPluginEnable
func XPluginEnable() {
}

//export XPluginReceiveMessage
func XPluginReceiveMessage(inFromWho, inMessage C.int, inParam unsafe.Pointer) {
}
