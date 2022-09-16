package index_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/matryer/is"
)

var root = os.Getenv("ROOT")

func BenchmarkCountWalk(b *testing.B) {
	is := is.New(b)

	count := 0
	for i := 0; i < b.N; i++ {
		count = 0
		err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			count++
			return nil
		})

		is.NoErr(err)
	}

	fmt.Printf("BenchmarkCountWalk: %d\n", count)
}

func BenchmarkCountWalkDir(b *testing.B) {
	is := is.New(b)

	count := 0
	for i := 0; i < b.N; i++ {
		count = 0
		err := filepath.WalkDir(root, func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			count++
			return nil
		})

		is.NoErr(err)
	}

	fmt.Printf("BenchmarkCountWalkDir: %d\n", count)
}

func BenchmarkCountFindShell(b *testing.B) {
	is := is.New(b)

	count := 0
	for i := 0; i < b.N; i++ {
		count = 0
		c1 := exec.Command("find", root, "-type", "f")
		c2 := exec.Command("wc", "-l")

		r, w := io.Pipe()
		c1.Stdout = w
		c2.Stdin = r

		var b2 bytes.Buffer
		c2.Stdout = &b2

		c1.Start()
		c2.Start()
		c1.Wait()
		w.Close()
		c2.Wait()

		countStr := strings.TrimSpace(b2.String())
		c, err := strconv.Atoi(countStr)
		is.NoErr(err)
		count = c
	}

	fmt.Printf("BenchmarkCountFindShell: %d\n", count)
}

// func BenchmarkIndex(b *testing.B) {
// 	is := is.New(b)
//
// 	for i := 0; i < b.N; i++ {
// 		index := New(root, []string{".php"})
// 		total := make(chan uint, 1)
// 		completed := &atomic.Uint64{}
// 		err := index.Index(total, completed)
// 		is.NoErr(err)
// 	}
// }

// Little reference of how progress would look, not really a test.
// func TestIndex(t *testing.T) {
// 	is := is.New(t)
//
// 	index := &index{
// 		root:                root,
// 		validFileExtensions: []string{".php"},
// 		files:               make(map[string]string, 1000),
// 	}
//
// 	total := make(chan uint, 1)
// 	completed := &atomic.Uint64{}
//
// 	var totalFiles uint
// 	ticker := time.NewTicker(time.Millisecond * 50)
// 	done := make(chan bool)
// 	go func() {
// 		for {
// 			select {
// 			case t := <-total:
// 				totalFiles = t
// 			case <-ticker.C:
// 				c := completed.Load()
//
// 				if totalFiles > 0 {
// 					fmt.Printf(
// 						"%d/%d (%.0f%%)\n",
// 						c,
// 						totalFiles,
// 						(float64(c)/float64(totalFiles))*100,
// 					)
// 					continue
// 				}
//
// 				fmt.Printf("%d/unknown\n", c)
// 			case <-done:
// 				fmt.Printf("%d/%d (100%%)\n", completed.Load(), totalFiles)
// 				fmt.Println("Completed")
// 				return
// 			}
// 		}
// 	}()
//
// 	err := index.Index(total, completed)
// 	done <- true
// 	is.NoErr(err)
// }
