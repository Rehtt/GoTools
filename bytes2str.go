package GoTools

import "unsafe"

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}