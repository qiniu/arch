package von

import (
	"bytes"
	"encoding/binary"
	"log"
)

// 指令格式：<op uint16> <params []byte>
// 函数调用格式：
//   push <zero-val-of-ret-type>
//   push arg1
//   push arg2
//   ...
//   push argN
//   call procName
// 函数实现格式：
//   pusha -2   ;取倒数第2个参数
//   pusha -1   ;取倒数第1个参数
//   ...
//   seta -3    ;将运算结果设置到倒数第3个参数 (可能就是返回值的位置)
//   ret 2      ;函数返回，将所有参数弹出去 (只留下返回值)
//
const (
	NOOP   = iota
	ADD    // 加法 (arg1, arg2 int64)
	SUB    // 减法 (arg1, arg2 int64)
	MUL    // 乘法 (arg1, arg2 int64)
	DIV    // 除法 (arg1, arg2 int64)
	MOD    // 取余 (arg1, arg2 int64)
	NEG    // 取负 (arg1 int64)
	LTI    // 小于 (arg1, arg2 int64)
	LTS    // 小于 (arg1, arg2 []byte)
	EQI    // 等于 (arg1, arg2 int64)
	EQS    // 等于 (arg1, arg2 []byte)
	NOT    // 非 (arg1 int64)
	CONCAT // 字符串连接 (arg1, arg2 []byte)
	INDEX  // 取字符 (s []byte, index int64)
	STRING // 字符转为字符串 (ch int64)
	ALLOC  // 申请空间 (n int64)
	READ   // 读取端口数据 <port uint16> (arg1 []byte)
	WRITE  // 写端口数据 <port uint16> (arg1 []byte)
	PUSHI  // 入栈 <val int64>
	PUSHS  // 入栈 <len uint16> <val [len]byte>
	PUSHA  // 取参数入栈 <index int16>
	SETA   // 修改参数 <index int16> (arg1 interface{})
	JMP    // 跳转到 <delta int64>
	JZ     // 如果为假跳转到 <delta int64>
	CALL   // 调用函数 <delta int64>
	RET    // 返回 <narg uint16>
	HALT   // 终止
)

type CPU struct {
	mem  *Memory
	devs map[int]Device
	stk  []interface{}
	bp   int
}

func NewCPU(mem *Memory) *CPU {
	devs := make(map[int]Device)
	return &CPU{mem: mem, devs: devs, bp: -1}
}

func (p *CPU) AddDevice(port int, dev Device) {
	p.devs[port] = dev
}

func (p *CPU) Run(pc int64) {
	mem := p.mem
	for {
		op := readU16(mem, pc)
		switch op {
		case ADD:
			v := p.pop().(int64)
			ret := p.top(1)
			*ret = (*ret).(int64) + v
			pc += 2
			debug("ADD:", p.stk)
		case SUB:
			v := p.pop().(int64)
			ret := p.top(1)
			*ret = (*ret).(int64) - v
			pc += 2
			debug("SUB:", p.stk)
		case MUL:
			v := p.pop().(int64)
			ret := p.top(1)
			*ret = (*ret).(int64) * v
			pc += 2
			debug("MUL:", p.stk)
		case DIV:
			v := p.pop().(int64)
			ret := p.top(1)
			*ret = (*ret).(int64) / v
			pc += 2
			debug("DIV:", p.stk)
		case MOD:
			v := p.pop().(int64)
			ret := p.top(1)
			*ret = (*ret).(int64) % v
			pc += 2
			debug("MOD:", p.stk)
		case NEG:
			ret := p.top(1)
			*ret = -(*ret).(int64)
			pc += 2
			debug("NEG:", p.stk)
		case LTI:
			v := p.pop().(int64)
			ret := p.top(1)
			*ret = fromBool((*ret).(int64) < v)
			pc += 2
			debug("LTI:", p.stk)
		case LTS:
			v := p.pop().([]byte)
			ret := p.top(1)
			*ret = fromBool(bytes.Compare((*ret).([]byte), v) < 0)
			pc += 2
			debug("LTS:", p.stk)
		case EQI:
			v := p.pop().(int64)
			ret := p.top(1)
			*ret = fromBool((*ret).(int64) == v)
			pc += 2
			debug("EQI:", p.stk)
		case EQS:
			v := p.pop().([]byte)
			ret := p.top(1)
			*ret = fromBool(bytes.Equal((*ret).([]byte), v))
			pc += 2
			debug("EQS:", p.stk)
		case NOT:
			ret := p.top(1)
			*ret = not((*ret).(int64))
			pc += 2
			debug("NOT:", p.stk)
		case CONCAT:
			v := p.pop().([]byte)
			ret := p.top(1)
			*ret = append((*ret).([]byte), v...)
			pc += 2
			debug("CONCAT:", p.stk)
		case INDEX:
			v := p.pop().(int64)
			ret := p.top(1)
			*ret = int64(((*ret).([]byte))[v])
			pc += 2
			debug("INDEX:", p.stk)
		case STRING:
			ret := p.top(1)
			*ret = []byte(string(rune((*ret).(int64))))
			pc += 2
			debug("STRING:", p.stk)
		case ALLOC:
			ret := p.top(1)
			*ret = make([]byte, (*ret).(int64))
			pc += 2
			debug("ALLOC:", p.stk)
		case READ:
			port := readU16(mem, pc+2)
			buf := p.pop().([]byte)
			dev := p.devs[int(port)]
			n, err := dev.Read(buf)
			if err != nil {
				panic(err)
			}
			p.push(int64(n))
			pc += 4
			debug("READ:", p.stk)
		case WRITE:
			port := readU16(mem, pc+2)
			buf := p.pop().([]byte)
			dev := p.devs[int(port)]
			n, err := dev.Write(buf)
			if err != nil {
				panic(err)
			}
			p.push(int64(n))
			pc += 4
			debug("WRITE:", p.stk)
		case PUSHI:
			v := readI64(mem, pc+2)
			p.push(v)
			pc += 10
			debug("PUSHI:", p.stk)
		case PUSHS:
			n := readU16(mem, pc+2)
			v := readBytes(mem, pc+4, int(n))
			p.push(v)
			pc += 4 + int64(n)
			debug("PUSHS:", p.stk)
		case PUSHA:
			index := readU16(mem, pc+2)
			p.push(p.arg(int16(index)))
			pc += 4
			debug("PUSHA:", p.stk, "BP:", p.bp)
		case SETA:
			index := readU16(mem, pc+2)
			v := p.pop()
			p.stk[p.bp+int(int16(index))] = v
			debug("SETA:", p.stk)
			pc += 4
		case JMP:
			delta := readI64(mem, pc+2)
			pc += delta
			debug("JMP")
		case JZ:
			delta := readI64(mem, pc+2)
			v := p.pop().(int64)
			if v == 0 {
				pc += delta
			} else {
				pc += 10
			}
			debug("JZ:", pc)
		case CALL:
			base := p.bp
			delta := readI64(mem, pc+2)
			p.bp = len(p.stk)
			p.push(&frame{pc: pc + 10, bp: base})
			pc += delta
			debug("CALL:", p.stk)
		case RET:
			narg := readU16(mem, pc+2)
			f := p.arg(0).(*frame)
			p.stk = p.stk[:p.bp-int(narg)]
			p.bp = f.bp
			pc = f.pc
			debug("RET:", p.stk, "PC:", pc)
		case HALT:
			debug("HALT")
			return
		default:
			debug("Unknown instruction:", op)
			panic("Unknown instruction")
		}
	}
}

type frame struct {
	pc int64
	bp int
}

func (p *CPU) Top(index int) interface{} {
	last := len(p.stk) - index
	return p.stk[last]
}

func (p *CPU) arg(index int16) interface{} {
	return p.stk[p.bp+int(index)]
}

func (p *CPU) top(index int) *interface{} {
	last := len(p.stk) - index
	return &p.stk[last]
}

func (p *CPU) pop() (v interface{}) {
	last := len(p.stk) - 1
	v = p.stk[last]
	p.stk = p.stk[:last]
	return
}

func (p *CPU) push(v interface{}) {
	p.stk = append(p.stk, v)
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

func readBytes(mem *Memory, off int64, n int) (v []byte) {
	v = make([]byte, n)
	if _, err := mem.ReadAt(v, off); err != nil {
		panic(err)
	}
	return
}

func fromBool(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func not(b int64) int64 {
	if b == 0 {
		return 1
	}
	return 0
}

func debug(a ...interface{}) {
	if Debug {
		log.Println(a...)
	}
}

var (
	Debug bool
)
