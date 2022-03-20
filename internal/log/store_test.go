package log

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	data  = []byte("Hello, world!")
	width = uint64(len(data)) + lenWidth
)

func BenchmarkStoreAppendRead(t *testing.B) {
	total := 0
	f, err := ioutil.TempFile("", "store_append_read_benchmark")

	require.NoError(t, err)

	defer os.Remove(f.Name())

	s, err := newStore(f)
	require.NoError(t, err)
	total += testAppend(t, s, uint64(t.N))
	testRead(t, s, uint64(t.N))

	t.SetBytes(int64(total))
}

func TestStoreAppendRead(t *testing.T) {
	f, err := ioutil.TempFile("", "store_append_read_test") // Create a temporary file for our tests

	require.NoError(t, err)

	defer os.Remove(f.Name()) // Remove tempfile when we're done

	s, err := newStore(f) // create a new store, using our tempfile as the store file
	require.NoError(t, err)

	testAppend(t, s, 4)
	testRead(t, s, 4)
	testReadAt(t, s)

	s, err = newStore(f) // recreate our store, using the same file, to ensure consistency.

	require.NoError(t, err)
	testRead(t, s, 4)
}

func testReadAt(t require.TestingT, s *store) int {
	//t.Helper()
	total := 0
	for i, off := uint64(1), int64(0); i < 4; i++ {
		b := make([]byte, lenWidth)
		n, err := s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, lenWidth, n)

		off += int64(n)

		size := enc.Uint64(b)

		b = make([]byte, size)
		n, err = s.ReadAt(b, off)
		require.NoError(t, err)
		require.Equal(t, int(size), n)
		off += int64(n)
		total += n
	}

	return total
}

func testAppend(t require.TestingT, s *store, i uint64) int {
	//	t.Helper()
	total := 0
	for j := uint64(1); j < i; j++ {
		n, pos, err := s.Append(data) // write the data to our store
		require.NoError(t, err)
		require.Equal(t, pos+n, width*j) // ensure our offset is correct
		total += int(n)
	}

	return total
}

func testRead(t require.TestingT, s *store, i uint64) int {
	//	t.Helper()

	total := 0
	var pos uint64
	for j := uint64(1); j < i; j++ {
		b, err := s.Read(pos)
		require.NoError(t, err)
		require.Equal(t, data, b)

		pos += width
		total += len(b)
	}

	return total
}

func TestStoreClose(t *testing.T) {
	f, err := ioutil.TempFile("", "store_close_test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	s, err := newStore(f)
	require.NoError(t, err)

	_, _, err = s.Append(data)
	require.NoError(t, err)

	f, beforeSize, err := openFile(f.Name())
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)

	_, afterSize, err := openFile(f.Name())
	require.NoError(t, err)

	require.True(t, afterSize > beforeSize)
}

func openFile(name string) (file *os.File, size int64, err error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, 0, err
	}

	fi, err := f.Stat()

	if err != nil {
		return nil, 0, err
	}

	return f, fi.Size(), nil
}
