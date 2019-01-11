package types

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

// Srt a collection of Subtitle
type Srt struct {
	Filename  string
	Subtitles []Subtitle
}

// Load loads Srt from a file
func (srt *Srt) Load(file string) error {
	srt.Filename = file
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	return srt.UnmarshalText(bs)
}

// MarshalText converts Srt to bytes
func (srt *Srt) MarshalText() ([]byte, error) {
	out := []byte{}
	deletedCount := 0
	for i, sub := range srt.Subtitles {
		if sub.Delete {
			deletedCount++
			continue
		}
		sub.Num = i + 1 - deletedCount
		bs, err := sub.MarshalText()
		if err != nil {
			return nil, err
		}
		out = append(out, bs...)
	}
	return out, nil
}

// UnmarshalText converts from bytes to Srt object
func (srt *Srt) UnmarshalText(content []byte) error {
	var err error
	var bs []byte
	reader := bufio.NewReader(bytes.NewReader(content))
	for line, err := reader.ReadBytes('\n'); err == nil; line, err = reader.ReadBytes('\n') {
		if len(bytes.TrimSpace(line)) == 0 {
			sub := Subtitle{}
			if err = sub.UnmarshalText(bs); err != nil {
				return err
			}
			srt.Subtitles = append(srt.Subtitles, sub)
			bs = []byte{}
			continue
		}
		bs = append(bs, line...)
	}
	if err != nil && err != io.EOF {
		return err
	}

	return nil
}

// Save saves Srt to file
func (srt *Srt) Save(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	bs, err := srt.MarshalText()
	if err != nil {
		return err
	}
	if _, err = f.Write(bs); err != nil {
		return err
	}
	return nil
}

// Shift add interval to all subtitle
func (srt *Srt) Shift(d time.Duration) {
	for i := range srt.Subtitles {
		srt.Subtitles[i].Start = srt.Subtitles[i].Start.Add(d)
		srt.Subtitles[i].End = srt.Subtitles[i].End.Add(d)
	}
}

// Search returns next row with a string
func (srt *Srt) Search(text string, init int) (int, error) {
	if text == "" {
		return 0, errors.New("not found")
	}
	for row, sub := range srt.Subtitles {
		if row <= init {
			continue
		}
		if strings.Index(sub.Text, text) >= 0 {
			return row, nil
		}
	}
	return 0, errors.New("not found")
}
