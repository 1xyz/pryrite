package tools

type BytesWriter struct {
	bytes []byte
}

func NewBytesWriter() *BytesWriter {
	return &BytesWriter{
		bytes: []byte{},
	}
}

func (b *BytesWriter) Write(p []byte) (int, error) {
	b.bytes = append(b.bytes, p...)
	return len(p), nil
}

func (b *BytesWriter) GetBytes() []byte {
	return b.bytes
}

func (b *BytesWriter) GetString() string {
	return string(b.bytes)
}
