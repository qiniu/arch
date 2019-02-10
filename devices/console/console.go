package console

import (
	"encoding/binary"
	"fmt"

	"github.com/qiniu/arch/drivers/console"
)

type Device struct {
}

func New() *Device {
	return &Device{}
}

func (p *Device) Read(b []byte) (int, error) {
	panic("can't read from console")
}

func (p *Device) Write(b []byte) (int, error) {
	switch b[0] {
	case console.PUTI:
		v := int64(binary.LittleEndian.Uint64(b[1:]))
		fmt.Print(v)
		return 9, nil
	case console.PUTS:
		fmt.Print(string(b[1:]))
		return len(b), nil
	default:
		panic("Device console: unknown instruction")
	}
}
