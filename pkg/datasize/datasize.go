package datasize

import "fmt"

const (
	BitsInByte          = 8
	BytesInKiloByte     = 1000
	KiloBytesInMegaByte = 1000
	MegaBytesInGigaByte = 1000
	GigaBytesInTeraByte = 1000

	Bit      Size = 1
	Byte     Size = 8 * Bit
	KiloByte Size = 1000 * Byte
	MegaByte Size = 1000 * KiloByte
	GigaByte Size = 1000 * MegaByte
	TeraByte Size = 1000 * GigaByte
)

type Size int

func (s Size) InBits() Size {
	return s
}

func (s Size) InBytes() Size {
	return s / BitsInByte
}

func (s Size) InKiloBytes() Size {
	return s.InBytes() / BytesInKiloByte
}

func (s Size) InMegaBytes() Size {
	return s.InKiloBytes() / KiloBytesInMegaByte
}

func (s Size) InGigaBytes() Size {
	return s.InMegaBytes() / MegaBytesInGigaByte
}

func (s Size) InTeraBytes() Size {
	return s.InGigaBytes() / GigaBytesInTeraByte
}

// https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
func (s Size) String() string {
	if s <= BitsInByte {
		return fmt.Sprintf("%d Bits", s)
	}

	b := s.InBytes()
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
