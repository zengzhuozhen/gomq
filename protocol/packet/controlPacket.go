package packet

import "io"

type ControlPacket interface {
	Write(w io.Writer) error
	Read(r io.Reader, header FixedHeader) error
	ReadHeadOnly(r io.Reader, header FixedHeader) error
}
