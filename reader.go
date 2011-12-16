package schematic

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
)

const (
	tagEnd       = 0
	tagByte      = 1
	tagShort     = 2
	tagInt       = 3
	tagLong      = 4
	tagFloat     = 5
	tagDouble    = 6
	tagByteArray = 7
	tagString    = 8
	tagList      = 9
	tagCompound  = 10
)

type nbtReader struct {
	r *bufio.Reader
}

func newNbtReader(r io.Reader) (nr *nbtReader, err os.Error) {
	var rd io.Reader
	if rd, err = gzip.NewReader(r); err != nil {
		return
	}
	return &nbtReader{r: bufio.NewReader(rd)}, nil
}

func (r *nbtReader) ReadString() (str string, err os.Error) {
	var l int
	if l, err = r.ReadShort(); err != nil {
		return
	}
	data := make([]byte, l)
	if _, err = io.ReadFull(r.r, data); err != nil {
		return
	}
	return string(data), nil
}

func (r *nbtReader) ReadShort() (val int, err os.Error) {
	buf := [2]byte{}
	if _, err = io.ReadFull(r.r, buf[:]); err != nil {
		return
	}
	val = int(buf[1]) + (int(buf[0]) << 8) // Big Endian
	return
}

func (r *nbtReader) ReadInt() (val int, err os.Error) {
	buf := [4]byte{}
	if _, err = io.ReadFull(r.r, buf[:]); err != nil {
		return
	}
	for i := 0; i < 4; i++ {
		val <<= 8
		val += int(buf[i])
	}
	return
}

func (r *nbtReader) ReadTagTyp() (typ byte, err os.Error) {
	typ, err = r.r.ReadByte()
	return
}

func (r *nbtReader) ReadTagName() (typ byte, name string, err os.Error) {
	if typ, err = r.r.ReadByte(); err != nil {
		return
	}
	if typ == tagEnd {
		return
	}
	name, err = r.ReadString()
	return
}

func (r *nbtReader) ReadByteArray() (data []byte, err os.Error) {
	var l int
	if l, err = r.ReadInt(); err != nil {
		return
	}
	data = make([]byte, l)
	_, err = io.ReadFull(r.r, data)
	return
}

type schematicReader struct {
	r *nbtReader
}

func newSchematicReader(r io.Reader) (sr *schematicReader, err os.Error) {
	var nr *nbtReader
	if nr, err = newNbtReader(r); err != nil {
		return
	}
	return &schematicReader{r: nr}, nil
}

type Entity struct {
	Id string
}

type Schematic struct {
	Width     int
	Length    int
	Height    int
	WEOffsetX int
	WEOffsetY int
	WEOffsetZ int
	Materials string
	Blocks    []byte
	Data      []byte
	Entities  []Entity
}

func (r *schematicReader) ReadEntity() (entity Entity, err os.Error) {
	for {
		var typ byte
		var name string
		if typ, name, err = r.r.ReadTagName(); err != nil {
			return
		}
		if typ == tagEnd {
			break
		}
		switch name {
		default:
			err = fmt.Errorf("Unknown entity field: %s", name)
			return
		}
	}
	return
}

func (r *schematicReader) ReadEntities() (entities []Entity, err os.Error) {
	for {
		var typ byte
		if typ, err = r.r.ReadTagTyp(); err != nil {
			return
		}
		if typ == tagEnd {
			break
		}
		if typ == tagCompound {
		}
		var entity Entity
		if entity, err = r.ReadEntity(); err != nil {
			return
		}
		entities = append(entities, entity)
	}
	return
}

func (r *schematicReader) Parse() (s *Schematic, err os.Error) {
	var typ byte
	var name string
	if typ, name, err = r.r.ReadTagName(); err != nil {
		return
	}
	if typ != tagCompound {
		return nil, fmt.Errorf("Top level tag must be compound. Got: %d", typ)
	}
	if name != "Schematic" {
		return nil, fmt.Errorf("Unexpected tag name: %s, want: Schematic", name)
	}
	s = new(Schematic)
	for {
		if typ, name, err = r.r.ReadTagName(); err != nil {
			return
		}
		if typ == tagEnd {
			break
		}
		switch name {
		case "Width":
			s.Width, err = r.r.ReadShort()
		case "Length":
			s.Length, err = r.r.ReadShort()
		case "Height":
			s.Height, err = r.r.ReadShort()
		case "Materials":
			s.Materials, err = r.r.ReadString()
		case "Blocks":
			s.Blocks, err = r.r.ReadByteArray()
		case "Data":
			s.Data, err = r.r.ReadByteArray()
		case "WEOffsetX":
			s.WEOffsetX, err = r.r.ReadInt()
		case "WEOffsetY":
			s.WEOffsetY, err = r.r.ReadInt()
		case "WEOffsetZ":
			s.WEOffsetZ, err = r.r.ReadInt()
		case "Entities":
			s.Entities, err = r.ReadEntities()
		default:
			return nil, fmt.Errorf("Unexpected tag: %d, name: %s\n", typ, name)
		}
		if err != nil {
			return
		}
	}
	if s.Materials != "Alpha" {
		return nil, fmt.Errorf("Materials must have 'Alpha' value, got: '%s'", s.Materials)
	}
	return
}

func (s *Schematic) XLen() int {
	return s.Width
}

func (s *Schematic) YLen() int {
	return s.Height
}

func (s *Schematic) ZLen() int {
	return s.Length
}

func (s *Schematic) Get(x, y, z int) bool {
	if x < 0 || y < 0 || z < 0 || x >= s.XLen() || y >= s.YLen() || z >= s.ZLen() {
		return false
	}
	index := y*s.XLen()*s.ZLen() + z*s.XLen() + x
	return s.Blocks[index] != 0
}

func ReadSchematic(input io.Reader) (vol *Schematic, err os.Error) {
	var r *schematicReader
	if r, err = newSchematicReader(input); err != nil {
		return
	}
	vol, err = r.Parse()
	return
}
