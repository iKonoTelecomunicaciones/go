// goolm is a pure Go implementation of libolm. Libolm is a cryptographic library used for end-to-end encryption in Matrix and wirtten in C++.
// With goolm there is no need to use cgo when building Matrix clients in go.
/*
This package contains the possible errors which can occur as well as some simple functions. All the 'action' happens in the subdirectories.
*/
package goolm

func GetLibaryVersion() (major, minor, patch uint8) {
	return 3, 2, 14
}
