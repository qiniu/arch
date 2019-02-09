package von

import "fmt"

const PageSize = 1024

type PageEvent interface {
	OnPageMiss(ipage int64) ([]byte, error)
}

type Memory struct {
	data map[int64][]byte
	ev   PageEvent
}

func NewMemory(ev PageEvent) *Memory {
	data := make(map[int64][]byte)
	return &Memory{data, ev}
}

func (p *Memory) requirePage(ipage int64) []byte {
	page, ok := p.data[ipage]
	if !ok {
		newpage, err := p.ev.OnPageMiss(ipage)
		if err != nil {
			panic(err)
		}
		if len(newpage) != PageSize {
			panic(fmt.Errorf("OnPageMiss: len(newpage) != PageSize"))
		}
		p.data[ipage] = newpage
		return newpage
	}
	return page
}

func (p *Memory) ReadAt(b []byte, pos int64) (n int, err error) {
	ipage := pos / PageSize
	off := int(pos % PageSize)
	for {
		page := p.requirePage(ipage)
		readed := copy(b, page[off:])
		n += readed
		if len(b) == readed {
			return
		}
		b = b[readed:]
		off = 0
		ipage++
	}
}
