package bitmap

const MaxSize = uint16(65535)

type Bitmap struct {
	vals []byte
	size uint16
}

func New(size uint16) *Bitmap {
	if size == 0 || size >= MaxSize {
		size = MaxSize
	} else if remainder := size % 8; remainder != 0 {
		size += 8 - remainder
	}
	return &Bitmap{size: size, vals: make([]byte, size>>3+1)}
}

func (b *Bitmap) Size() uint16 {
	return b.size
}

func (b *Bitmap) Set(offset uint16, value uint8) bool {
	if b.size < offset {
		return false
	}

	index, pos := offset>>3, offset&0x07

	if value == 0 {
		b.vals[index] &^= 0x01 << pos
	} else {
		b.vals[index] |= 0x01 << pos
	}

	return true
}

func (b *Bitmap) Get(offset uint16) uint8 {
	if b.size < offset {
		return 0
	}

	index, pos := offset>>3, offset&0x07

	return (b.vals[index] >> pos) & 0x01
}
