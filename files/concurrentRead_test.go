package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/suite"
)

type concurrencyTestSuite struct {
	suite.Suite
	file     *os.File
	content  string
	expected []string
}

func TestFileSuitStart(t *testing.T) {
	suite.Run(t, new(concurrencyTestSuite))
}

func (f *concurrencyTestSuite) SetupSuite() {
	file, err := ioutil.TempFile(".", "testFile")
	if err != nil {
		require.Fail(f.T(), fmt.Sprintf("can't create tmp file in current dir, error is %s", err))
		return
	}
	f.content = "Hello world \n"
	if _, err = file.WriteString(strings.Repeat(f.content, 10000)); err != nil {
		require.Fail(f.T(), fmt.Sprintf("can't write tmp file content, error is %s", err))
		return
	}
	for i := 0; i < 10000; i++ {
		f.expected = append(f.expected, "hello", "world")
	}
	f.file = file
}

func (f *concurrencyTestSuite) TearDownSuite() {
	if err := f.file.Close(); err != nil {
		require.Fail(f.T(), fmt.Sprintf("can't close file with name: %s, error is %s", f.file.Name(), err))
		return
	}
	if err := os.Remove(f.file.Name()); err != nil {
		require.Fail(f.T(), fmt.Sprintf("can't remove file with name: %s, error is %s", f.file.Name(), err))
		return
	}
}

func (f *concurrencyTestSuite) TestConcurrentReadFile() {
	wordArr, _ := ConcurrentReadFile(f.file.Name())
	require.Equal(f.T(), f.expected, wordArr)
}

func (f *concurrencyTestSuite) TestConcurrentReadFile2() {
	if _, err := f.file.WriteString(strings.Repeat("filling the Ice \n", 10000)); err != nil {
		require.Fail(f.T(), "can't write tmp file content")
		return
	}
	for i := 0; i < 10000; i++ {
		f.expected = append(f.expected, "fill", "ice")
	}
	wordArr, _ := ConcurrentReadFile(f.file.Name())
	require.Equal(f.T(), f.expected, wordArr)
}
