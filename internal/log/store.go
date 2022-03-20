package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var enc = binary.BigEndian // encoding used to store records

const lenWidth = 8 // bytes used to store length of a record

type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())

	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())

	return &store{
		File: f,
		buf:  bufio.NewWriter(f),
		size: size,
	}, nil
}

// Appends given data to the store, returning bytes written and the offset
func (s *store) Append(data []byte) (n uint64, pos uint64, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pos = s.size // Set position to the end of the file

	// Write the length of our data to the store, that when we read the record, we'll know how many bytes to read
	if err := binary.Write(s.buf, enc, uint64(len(data))); err != nil {
		return 0, 0, err
	}
	// Write our data to the store
	w, err := s.buf.Write(data)
	if err != nil {
		return 0, 0, err
	}

	w += lenWidth // add the length offset

	s.size += uint64(w) // increase the size of our file

	return uint64(w), pos, nil
}

// Read the record stored at the given pos (if it exists)
func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Make sure our buffer is up-to-date
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}

	size := make([]byte, lenWidth)

	// Read the length of the record into size
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}

	// Create our data slice, using the size we got earlier
	data := make([]byte, enc.Uint64(size))

	// Read from the store into our data slice, using pos+length as our offset
	if _, err := s.File.ReadAt(data, int64(pos+lenWidth)); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *store) ReadAt(data []byte, offset int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return 0, err
	}

	return s.File.ReadAt(data, offset)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// make sure our buffer is finished writing
	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close() // close our file (and buffer)
}
