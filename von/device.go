package von

type Device interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
}
