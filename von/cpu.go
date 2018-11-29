package von

// <op byte> <len byte> <params [len]byte>

const (
	NOOP  = iota
	JMP   // 跳转到 <len=8 byte> <pc int64>
	READ  // 读取端口数据 <port>
	WRITE // 写端口数据 <port>
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
