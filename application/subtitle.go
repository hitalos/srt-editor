package application

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const (
	timeFormat = "15:04:05,999"

	newLineSymbol = "⏎"
)

// Subtitle a text and interval representation
type Subtitle struct {
	Num    int
	Start  Timestamp
	End    Timestamp
	Text   string
	Delete bool
}

// MarshalText converts object to bytes representation
func (s *Subtitle) MarshalText() ([]byte, error) {
	out := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", s.Num, s.Start, s.End, s.Text)
	return []byte(out), nil
}

// UnmarshalText build object from byte representation
func (s *Subtitle) UnmarshalText(bs []byte) error {
	reader := bufio.NewReader(bytes.NewReader(bs))
	line, err := reader.ReadBytes('\n')
	if err != nil {
		return err
	}

	if s.Num, err = strconv.Atoi(string(bytes.TrimSpace(line))); err != nil {
		return fmt.Errorf("sequence error: %q\n%s", string(bs), err)
	}

	interval, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	interval = strings.TrimSpace(interval)
	times := strings.Split(interval, " --> ")
	start, err := time.Parse(timeFormat, times[0])
	if err != nil {
		return err
	}
	end, err := time.Parse(timeFormat, times[1])
	if err != nil {
		return err
	}

	s.Start, s.End = Timestamp(start), Timestamp(end)

	for line, err := reader.ReadBytes('\n'); err == nil; line, err = reader.ReadBytes('\n') {
		s.Text = s.Text + string(bytes.TrimSpace(line)) + "\n"
	}
	s.Text = strings.TrimSpace(s.Text)

	if err == io.EOF {
		return nil
	}

	return err
}
