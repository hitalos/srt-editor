package types

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

const timeFormat = "15:04:05.999"

// Subtitle a text and interval representation
type Subtitle struct {
	Num    int
	Start  time.Time
	End    time.Time
	Text   string
	Delete bool
}

// MarshalText converts object to bytes representation
func (s *Subtitle) MarshalText() ([]byte, error) {
	start := strings.Replace(s.Start.Format(timeFormat), ".", ",", -1)
	end := strings.Replace(s.End.Format(timeFormat), ".", ",", -1)
	out := fmt.Sprintf("%d\n%s --> %s\n%s\n\n", s.Num, start, end, s.Text)
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
		return err
	}

	interval, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	interval = strings.TrimSpace(strings.Replace(interval, ",", ".", -1))
	times := strings.Split(interval, " --> ")
	s.Start, err = time.Parse(timeFormat, times[0])
	if err != nil {
		return err
	}
	s.End, err = time.Parse(timeFormat, times[1])
	if err != nil {
		return err
	}

	for line, err := reader.ReadBytes('\n'); err == nil; line, err = reader.ReadBytes('\n') {
		s.Text = s.Text + string(bytes.TrimSpace(line)) + "\n"
	}
	s.Text = strings.TrimSpace(s.Text)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}
