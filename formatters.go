package wlog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Formatter is a base interface for output formatters, it has
// one method called Format which will be called when outputting
// the write entry
type Formatter interface {
	Format(w io.Writer, l *Logger, msg string, entryTime time.Time) error
}

// JSONFormatter used to output logs in JSON format
type JSONFormatter struct{}

// Implements Formatter.Format
func (j JSONFormatter) Format(w io.Writer, l *Logger, msg string, entryTime time.Time) error {
	l.fields["msg"] = msg
	l.fields["timestamp"] = getTimestamp(entryTime)
	l.fields["level"] = l.logLevel.String()

	encoder := json.NewEncoder(w)

	if err := encoder.Encode(l.fields); err != nil {
		return fmt.Errorf("failed to marshal fields to JSON, %v", err)
	}

	return nil
}

// TextFormatter used to output logs in text format. This is the default
// formatter when creating a instance of wlog.
type TextFormatter struct{}

func (t TextFormatter) Format(w io.Writer, l *Logger, msg string, entryTime time.Time) error {

	// Write Date
	year, month, day := entryTime.Date()
	itoa(w, year, 4)
	writeString(w, "-")
	itoa(w, int(month), 2)
	writeString(w, "-")
	itoa(w, day, 2)

	writeString(w, " ")

	// Write time
	hour, min, sec := entryTime.Clock()
	itoa(w, hour, 2)
	writeString(w, ":")
	itoa(w, min, 2)
	writeString(w, ":")
	itoa(w, sec, 2)
	writeString(w, ":")
	itoa(w, entryTime.Nanosecond()/1e3, 6)

	writeString(w, " ")

	// Write log level
	var level string
	switch l.logLevel {
	case Dbg:
		level = "DBG "
	case Nfo:
		level = "NFO "
	case Wrn:
		level = "WRN "
	case Err:
		level = "ERR "
	case Ftl:
		level = "FTL "
	}

	writeString(w, level)

	// Append log message to buffer
	writeString(w, msg)

	if len(msg) == 0 || msg[len(msg)-1] != '\n' {
		writeString(w, "\n")
	}

	return nil
}

func getTimestamp(now time.Time) string {
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	nano := now.Nanosecond() / 1e3

	format := "%d-%02d-%02d %02d:%02d:%02d:%06d"

	return fmt.Sprintf(format, year, month, day, hour, min, sec, nano)
}

func writeString(w io.Writer, str string) {
	if _, err := io.WriteString(w, str); err != nil {
		fmt.Fprintf(os.Stderr, "could not write entry log, err: %s", err)
	}
}

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
// NOTE: Taken from Go's std log package
func itoa(w io.Writer, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)

	if _, err := w.Write(b[bp:]); err != nil {
		fmt.Fprintf(os.Stderr, "failed adding zero padding to %d", i)
	}
}