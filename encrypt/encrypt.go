package encrypt

import (
	"crypto/md5"
)

func EncryptText1Way(raw []byte, salt []byte) []byte {
	firstMD5 := md5.Sum(raw)
	secondMD5 := md5.Sum(append(firstMD5[:], salt...))
	return secondMD5[:]
}

func EncryptText2Way() {

}
