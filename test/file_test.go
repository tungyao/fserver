package test

import "testing"
import "../../file"
func TestFile(t *testing.T) {
	file.Start()
}
func TestSp(t *testing.T) {
	str:=` file="asi12kasldjkasd"; filename="asdjasjdajsdasd";`
	t.Log(file.SplitString([]byte(str), []byte("filename=")))
}
