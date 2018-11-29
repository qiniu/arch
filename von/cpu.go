package von

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
	// TODO
}
