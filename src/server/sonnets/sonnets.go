package sonnets

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Sonnet struct {
	File   os.DirEntry
	Title  string
	Author string
}

func (s *Sonnet) GetLine(dir string, line int) (string, error) {
	lastLine := 0
	content, err := os.ReadFile(dir + s.File.Name())
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(
		strings.NewReader(string(content)),
	)

	for scanner.Scan() {
		if lastLine == 0 {
			lastLine++
			continue // skip author - title
		}

		if lastLine == line {
			return scanner.Text(), scanner.Err()
		}

		lastLine++
	}

	return "", io.EOF
}

func GetSonnets(dir string) ([]Sonnet, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	sonnets := make([]Sonnet, len(files))

	for i, file := range files {
		filePtr, err := os.Open(dir + file.Name())
		if err != nil {
			return nil, err
		}
		defer filePtr.Close()

		reader := bufio.NewReader(filePtr)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Printf("Error reading %v: %v\n", file.Name(), err)
			return nil, err
		}

		metadata := strings.Split(string(line), "-")

		sonnets[i] = Sonnet{
			file,
			strings.Trim(metadata[1][:len(metadata[1])-1], " "),
			strings.Trim(metadata[0], " "),
		}
	}

	return sonnets, nil
}
