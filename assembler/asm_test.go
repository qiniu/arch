package assembler

import (
	"testing"

	"github.com/qiniu/arch/devices/console"
	"github.com/qiniu/arch/devices/keyboard"
	"github.com/qiniu/arch/drivers"
	"github.com/qiniu/arch/von"
)

type shortMem struct {
	b []byte
}

func newShortMem(b []byte) *shortMem {
	code := make([]byte, von.PageSize)
	copy(code, b)
	return &shortMem{code}
}

func (p *shortMem) OnPageMiss(ipage int64) ([]byte, error) {
	if ipage > 0 {
		panic("out of range")
	}
	return p.b, nil
}

func run(b *Builder) *von.CPU {
	b.Halt()
	code := b.Bytes()
	mem := von.NewMemory(newShortMem(code))
	cpu := von.NewCPU(mem)
	cpu.AddDevice(drivers.KEYBOARD, keyboard.New())
	cpu.AddDevice(drivers.CONSOLE, console.New())
	cpu.Run(0)
	return cpu
}

func notEq(ret interface{}, v string) bool {
	return string(ret.([]byte)) != v
}

func TestInt(t *testing.T) {
	asm := New(nil)
	asm.PushInt(2).
		PushInt(3).
		Mul()
	ret := run(asm).Top(1)
	if ret != int64(6) {
		t.Fatal("TestInt:", ret)
	}
}

func TestString(t *testing.T) {
	asm := New(nil)
	asm.PushString("Hello, ").
		PushString("World").
		Concat()
	ret := run(asm).Top(1)
	if notEq(ret, "Hello, World") {
		t.Fatal("TestString:", ret)
	}
}

func TestJZ_true(t *testing.T) {
	asm := New(nil)
	asm.PushString("Hello").
		PushString("World").
		LessThanString().
		JZ("else").
		PushString("true").
		Halt().
		Label("else").
		PushString("false")
	ret := run(asm).Top(1)
	if notEq(ret, "true") {
		t.Fatal("TestJZ_true:", ret)
	}
}

func TestJZ_false(t *testing.T) {
	asm := New(nil)
	asm.PushInt(3).
		PushInt(2).
		LessThanInt().
		JZ("else").
		PushString("true").
		Halt().
		Label("else").
		PushString("false")
	ret := run(asm).Top(1)
	if notEq(ret, "false") {
		t.Fatal("TestJZ_false:", ret)
	}
}

func TestProc(t *testing.T) {
	von.Debug = true
	asm := New(nil)
	asm.PushInt(0).
		PushInt(2).
		PushInt(3).
		Call("sub").
		Halt().
		Label("sub").
		PushArg(-2).
		PushArg(-1).
		Sub().
		SetArg(-3).
		Ret(2)
	ret := run(asm).Top(1)
	if ret != int64(-1) {
		t.Fatal("TestProc:", ret)
	}
}

func TestKeyboard(t *testing.T) {

}
