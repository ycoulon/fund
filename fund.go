package main

import (
	"bufio"
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	fundPath = ""
)

func visitFile(fp string, fi os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err) // can't walk here,
		return nil       // but continue walking elsewhere
	}

	if strings.Contains(fp, ".fund") {
		return nil
	}

	if !!fi.Mode().IsDir() {
		fmt.Println("Processing directory ", fp)
		return nil
	}

	calculHash(fp, fi)
	return nil
}
func calculHash(fp string, fi os.FileInfo) {
	file, err := os.Open(fp)
	if err != nil {
		log.Fatal(err)
	}

	// close file on exit and check for its returned error
	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// make a read buffer
	r := bufio.NewReader(file)
	// make a buffer to keep chunks that are read
	var buffsize int64
	buffsize = 1024

	if fi.Size() < 1024 {
		buffsize = fi.Size()
	}
	h := sha256.New()
	buf := make([]byte, buffsize)
	for {
		// read a chunk
		n, err := r.Read(buf)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		h.Write(buf)
		if n == 0 {
			break
		}

	}
	hash := hex.EncodeToString(h.Sum(nil))
	hashFile := path.Join(fundPath, hash)
	if _, err := os.Stat(hashFile); os.IsNotExist(err) {
		f, err := os.Create(hashFile)
		if err != nil {
			panic(err)
		}
		f.WriteString(fp + "\n")
		f.Close()
	} else {
		f, err := os.OpenFile(hashFile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if _, err = f.WriteString(fp + "\n"); err != nil {
			panic(err)
		}

	}
}

func getResult(fp string, fi os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err) // can't walk here,
		return nil       // but continue walking elsewhere
	}

	if !!fi.Mode().IsDir() {
		return nil
	}

	file, err := os.Open(fp)
	if err != nil {
		log.Fatal(err)
	}

	// close file on exit and check for its returned error
	defer file.Close()

	scanner := bufio.NewScanner(file)
	l := list.New()
	// scanner.Scan() advances to the next token returning false if an error was encountered
	for scanner.Scan() {
		l.PushBack(scanner.Text())
	}

	if l.Len() > 1 {
		fmt.Println("Duplicate files :")
		// Iterate through list and print its contents.
		for e := l.Front(); e != nil; e = e.Next() {
			fmt.Println(e.Value)
		}
	}
	return nil
}

func processResult() error {
	fmt.Println("Processing result...")
	filepath.Walk(fundPath, getResult)
	return nil
}
func main() {
	rootDir := os.Args[1]
	wd, _ := os.Getwd()
	fundPath = path.Join(wd, ".fund")

	if _, err := os.Stat(fundPath); !os.IsNotExist(err) {
		err := os.RemoveAll(fundPath)
		if err != nil {
			panic(err)
		}

	}

	err := os.Mkdir(".fund", 0755)
	if err != nil {
		panic(err)
	}

	fmt.Println("Processing ", rootDir)
	fmt.Println("Result will output to ", fundPath)
	filepath.Walk(rootDir, visitFile)

	processResult()
}
