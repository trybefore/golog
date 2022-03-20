package log

import (
	"io"
	"os"

	"github.com/edsrzf/mmap-go"
)

var (
	offsetWidth   uint64 = 4
	positionWidth uint64 = 8

	entireWidth = offsetWidth + positionWidth
)

type index struct {
	file *os.File
	mmap mmap.MMap
	size uint64
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size()) // set index size equal to file size

	// Truncate the file to our max for each segment
	if err = os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
		return nil, err
	}

	// Memory map the index file
	if idx.mmap, err = mmap.Map(f, mmap.RDWR, 0); err != nil {
		return nil, err
	}

	return idx, nil
}

func (i *index) Close() error {

	// Write pending data to disk
	if err := i.mmap.Flush(); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}

	// Truncate our file to match our index size
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}

	return i.file.Close()
}

// Takes in the offset of a record and returns it's position in the related store
func (i *index) Read(in int64) (offset uint32, position uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}

	if in == -1 {
		offset = uint32((i.size / entireWidth) - 1) // if in is -1, return latest value in index
	} else {
		offset = uint32(in)
	}

	position = uint64(offset) * entireWidth //

	// Ensure we're not reading outside the file
	if i.size < position+entireWidth {
		return 0, 0, io.EOF
	}

	offset = enc.Uint32(i.mmap[position : position+offsetWidth])               // offset where we store the record's position
	position = enc.Uint64(i.mmap[position+offsetWidth : position+entireWidth]) // position of the record

	return offset, position, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entireWidth {
		return io.EOF
	}

	enc.PutUint32(i.mmap[i.size:i.size+offsetWidth], off)
	enc.PutUint64(i.mmap[i.size+offsetWidth:i.size+entireWidth], pos)
	i.size += uint64(entireWidth)

	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
