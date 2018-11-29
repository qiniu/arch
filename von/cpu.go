package von

import "encoding/binary"

// <op byte> <oplen byte> <params [len]byte>

const (
	NOOP  = iota
	JMP   // 跳转到 <oplen=8 byte> <pc int64>
	READ  // 读取端口数据 <oplen=2 byte> <port uint16>
	WRITE // 写端口数据 <oplen=2 byte> <port uint16>
)

type CPU struct {
	mem  *Memory
	pc   int64
	devs map[int]Device
}

func NewCPU(mem *Memory, pc int64) *CPU {
	devs := make(map[int]Device)
	return &CPU{mem, pc, devs}
}

func (p *CPU) AddDevice(port int, dev Device) {
	p.devs[port] = dev
}

func (p *CPU) Run() {
	pc := p.pc
	mem := p.mem
	for {
		op := readU8(mem, pc)
		switch op {
		case JMP:
		case READ:
		case WRITE:
			//TODO
		}
	}
}

func readU8(mem *Memory, off int64) (v byte) {
	var buf [1]byte
	if _, err := mem.ReadAt(buf[:], off); err != nil {
		panic(err)
	}
	return buf[0]
}

func readU16(mem *Memory, off int64) (v uint16) {
	var buf [2]byte
	if _, err := mem.ReadAt(buf[:], off); err != nil {
		panic(err)
	}
	return binary.LittleEndian.Uint16(buf[:])
}

func readI64(mem *Memory, off int64) (v int64) {
	var buf [8]byte
	if _, err := mem.ReadAt(buf[:], off); err != nil {
		panic(err)
	}
	return int64(binary.LittleEndian.Uint64(buf[:]))
}
