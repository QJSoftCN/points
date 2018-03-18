package points

import (
	"fmt"
	"unsafe"
	"github.com/axgle/mahonia"
)

func INT8FromString(s string) ([]byte, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return nil, nil
		}
	}

	return []byte(s), nil
}

func StringToINT8(s string) []byte {

	a, err := INT8FromString(s)

	if err != nil {
		fmt.Println(" string is nil")
	}

	return a
}

func StringToINT8Ptr(s string) *byte {
	return &StringToINT8(s)[0]
}


func Uintptr(ptr unsafe.Pointer)uintptr{
	return uintptr(ptr)
}

func BytePtr(strs []string) []*byte {

	bptrs := make([]*byte, len(strs))

	for index, s := range strs {
		bptrs[index] = StringToINT8Ptr(s)
	}

	return bptrs
}

func IntPtr(n int) uintptr {
	return uintptr(n)
}

var (
	GBK = mahonia.NewDecoder("gbk")
)

func BytePtrToString(ptr *byte) string {

	buf := make([]byte, 0, 256)
	a := uintptr(unsafe.Pointer(ptr))
	for ; ; {
		b := *((*byte)(unsafe.Pointer(a)))
		if b == 0 {
			break
		}
		buf = append(buf, b)
		a += 1

	}

	str := string(buf)

	str = GBK.ConvertString(str)

	return str
}
