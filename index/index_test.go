package index

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type searchTestSuite struct {
	suite.Suite
	index              Index
	firstSearchQuery   []string
	secondSearchQuery  []string
	firstExpectedCase  map[string]int
	secondExpectedCase map[string]int
}

func TestSearchTestSuite(t *testing.T) {
	suite.Run(t, new(searchTestSuite))
}

func (f *searchTestSuite) SetupTest() {
	f.index = make(Index)
	f.firstExpectedCase = make(map[string]int)
	f.secondExpectedCase = make(map[string]int)
	f.index["hello"] = []string{"file1", "file2"}
	f.index["world"] = []string{"file1", "file4"}
	f.index["golang"] = []string{"file2", "file3", "file4"}
	f.index["java"] = []string{"file1"}
	f.index["architectur"] = []string{"file1"}
	f.firstSearchQuery = []string{"hello", "world"}
	f.secondSearchQuery = []string{"golang", "java"}
	f.firstExpectedCase = map[string]int{
		"file1": 2,
		"file2": 1,
		"file4": 1,
	}
	f.secondExpectedCase = map[string]int{
		"file1": 1,
		"file2": 1,
		"file3": 1,
		"file4": 1,
	}
}

func (f *searchTestSuite) TestBuildSearchIndex() {
	actual, err := f.index.BuildSearchIndex(f.firstSearchQuery)
	require.NoError(f.T(), err)
	require.Equal(f.T(), f.firstExpectedCase, actual)
}

func (f *searchTestSuite) TestBuildSearchIndex2() {
	f.secondSearchQuery = append(f.secondSearchQuery, "architecture")
	actual, err := f.index.BuildSearchIndex(f.secondSearchQuery)
	require.NoError(f.T(), err)
	f.secondExpectedCase["file1"] = 2
	require.Equal(f.T(), f.secondExpectedCase, actual)
}

type indexTestSuite struct {
	suite.Suite
	wg          *sync.WaitGroup
	index       Index
	searchQuery []string
	expected    map[string]string
	dataChan    chan map[string]string
	content     string
	file        *os.File
}

func TestIndexTestSuite(t *testing.T) {
	suite.Run(t, new(indexTestSuite))
}
func (f *indexTestSuite) SetupSuite() {
	file, err := ioutil.TempFile(".", "testFile")
	if err != nil {
		require.Fail(f.T(), fmt.Sprintf("can't create tmp file in current dir, error is %s", err))
		return
	}
	f.content = "Hello world \n"
	if _, err := file.WriteString(strings.Repeat(f.content, 10000)); err != nil {
		require.Fail(f.T(), fmt.Sprintf("can't write tmp file content, error is %s", err))
		return
	}
	f.file = file
}

func (f *indexTestSuite) SetupTest() {
	f.wg = &sync.WaitGroup{}
	f.wg.Add(1)
	f.dataChan = make(chan map[string]string, 10)

	f.index = make(Index)
	f.index["hello"] = []string{f.file.Name()}
	f.index["world"] = []string{f.file.Name()}
	f.expected = make(map[string]string)
	f.expected["hello"] = f.file.Name()
	f.expected["world"] = f.file.Name()
}

func (f *indexTestSuite) TearDownTest() {
	close(f.dataChan)
}

func (f *indexTestSuite) TearDownSuite() {
	if err := f.file.Close(); err != nil {
		require.Fail(f.T(), fmt.Sprintf("can't close file with name: %s, error is %s", f.file.Name(), err))
		return
	}
	if err := os.Remove(f.file.Name()); err != nil {
		require.Fail(f.T(), fmt.Sprintf("can't remove file with name: %s, error is %s", f.file.Name(), err))
		return
	}
}

func (f *indexTestSuite) TestConcurrentBuildFileMap() {
	go func() {
		for data := range f.dataChan {
			require.Equal(f.T(), f.expected, data)
		}
	}()
	ConcurrentBuildFileMap(f.wg, f.file.Name(), f.dataChan)
}

func (f *indexTestSuite) TestAsyncConcurrentBuildFileMap() {
	go func() {
		for data := range f.dataChan {
			require.Equal(f.T(), f.expected, data)
		}
	}()
	go ConcurrentBuildFileMap(f.wg, f.file.Name(), f.dataChan)
	f.wg.Add(1)
	go ConcurrentBuildFileMap(f.wg, f.file.Name(), f.dataChan)
	f.wg.Wait()
}

func (f *indexTestSuite) TestCreateInvertedIndex() {
	m, err := CreateInvertedIndex([]string{f.file.Name()})
	require.NoError(f.T(), err)
	require.Equal(f.T(), f.index, *m)
}
