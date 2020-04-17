package files

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/rs/zerolog/log"

	"github.com/polisgo2020/search-Arkronzxc/util"
)

// ConcurrentReadFile concurrently read file and returns word array from file
func ConcurrentReadFile(filename string) (wordArr []string, err error) {
	log.Debug().Str("filename", filename)
	wg := sync.WaitGroup{}

	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	//chunkSize is 1 mb
	const chunkSize = 1024 * 1024
	goRoutineCount := int(math.Ceil(float64(util.FileSize(filename)/chunkSize))) + 1
	log.Debug().Int("goroutine count", goRoutineCount)

	wordChannel := make(chan string, goRoutineCount)

	ctx, finish := context.WithCancel(context.Background())
	log.Debug().Interface("context", ctx)

	errChannel := make(chan error, goRoutineCount)

	// Current signifies the counter for bytes of the file.
	var current int64

	// Limit signifies the chunk size of file to be processed by every thread.
	var limit int64 = chunkSize
	log.Debug().Int64("limit", limit)

	// adds goroutine to the wait group
	for i := 0; i < goRoutineCount; i++ {
		wg.Add(1)
		//start read goroutine which reads and handles curtain part of file
		go read(ctx, &wg, current, limit, file, wordChannel, errChannel)
		//point start of the next chunk by adding limit + 1 byte.
		//Adding one byte will prevent the start of the next chunk exactly on the end of the previous chunk
		current += limit + 1
	}

	// starts goroutine which waits end of all the goroutines
	go func(chW chan string, errChan chan error, wg *sync.WaitGroup) {
		wg.Wait()
		close(chW)
		close(errChan)
	}(wordChannel, errChannel, &wg)

	wordArr = make([]string, 0)

	// receiving values from either word channel or error channel until one of them won't close
ReadLoop:
	for {
		select {
		case data, ok := <-wordChannel:
			// means the channel is already empty and closed
			if !ok {
				break ReadLoop
			}
			wordArr = append(wordArr, data)

		case errData, ok := <-errChannel:
			if !ok {
				break ReadLoop
			}
			// if some data came to err channel we send terminating signal to all other goroutines which got that context
			finish()
			return nil, errData
		}
	}
	log.Debug().Strs("word array", wordArr)
	return wordArr, nil
}

// read writes in the word channel the words
func read(ctx context.Context, wg *sync.WaitGroup, offset int64, limit int64, file *os.File,
	wordChannel chan<- string, errChan chan<- error) {
	log.Debug().Interface("context", ctx).Interface("wg", wg).Int64("offset", offset).
		Int64("limit", limit).Interface("file", file)

	defer wg.Done()

	// shifts the pointer to the offset value.
	_, _ = file.Seek(offset, 0)
	reader := bufio.NewReader(file)

	// skips all the bytes before first space because they refer to the previous chunk if it's not the first chunk.
	// If it is then starting to read from the start.
	if offset != 0 {
		_, err := reader.ReadBytes(' ')
		if err == io.EOF {
			fmt.Println("EOF")
			return
		}

		if err != nil {
			log.Err(err)
			errChan <- err
			return
		}
	}
	log.Debug().Int64("offset", offset)

	// size of read bytes
	var cumulativeSize int64

	// iterates over a space separated byte buffer. Another case is it terminates if context is done.
	// It becomes done if some error occurred in any reading goroutine.
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if cumulativeSize > limit {
				return
			}

			b, err := reader.ReadBytes(' ')

			if err == io.EOF {
				return
			}

			if err != nil {
				errChan <- err
				return
			}

			cumulativeSize += int64(len(b))
			s := string(b)
			if s != "" {
				t, _ := utf8.DecodeRune([]byte("'"))
				f := func(c rune) bool {
					return !unicode.IsLetter(c) && t != c
				}
				str := strings.FieldsFunc(s, f)

				for i := range str {
					w, err := util.CleanUserData(str[i])
					if err != nil {
						errChan <- err
						return

					}
					if w != "" {
						wordChannel <- w
					}
				}
			}
		}
	}
}
