package tsi1_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/tsdb/engine/tsi1"
)

// Ensure log file can append series.
func TestLogFile_AddSeries(t *testing.T) {
	f := MustOpenLogFile()
	defer f.Close()

	// Add test data.
	if err := f.AddSeries([]byte("mem"), models.Tags{{Key: []byte("host"), Value: []byte("serverA")}}); err != nil {
		t.Fatal(err)
	} else if err := f.AddSeries([]byte("cpu"), models.Tags{{Key: []byte("region"), Value: []byte("us-east")}}); err != nil {
		t.Fatal(err)
	} else if err := f.AddSeries([]byte("cpu"), models.Tags{{Key: []byte("region"), Value: []byte("us-west")}}); err != nil {
		t.Fatal(err)
	}

	// Verify data.
	itr := f.MeasurementIterator()
	if e := itr.Next(); e == nil || string(e.Name) != "cpu" {
		t.Fatalf("unexpected measurement: %#v", e)
	} else if e := itr.Next(); e == nil || string(e.Name) != "mem" {
		t.Fatalf("unexpected measurement: %#v", e)
	} else if e := itr.Next(); e != nil {
		t.Fatalf("expected eof, got: %#v", e)
	}
}

// Ensure log file can delete an existing measurement.
func TestLogFile_DeleteMeasurement(t *testing.T) {
	f := MustOpenLogFile()
	defer f.Close()

	// Add test data.
	if err := f.AddSeries([]byte("mem"), models.Tags{{Key: []byte("host"), Value: []byte("serverA")}}); err != nil {
		t.Fatal(err)
	} else if err := f.AddSeries([]byte("cpu"), models.Tags{{Key: []byte("region"), Value: []byte("us-east")}}); err != nil {
		t.Fatal(err)
	} else if err := f.AddSeries([]byte("cpu"), models.Tags{{Key: []byte("region"), Value: []byte("us-west")}}); err != nil {
		t.Fatal(err)
	}

	// Remove measurement.
	if err := f.DeleteMeasurement([]byte("cpu")); err != nil {
		t.Fatal(err)
	}

	// Verify data.
	itr := f.MeasurementIterator()
	if e := itr.Next(); !reflect.DeepEqual(e, &tsi1.MeasurementElem{Name: []byte("cpu"), Deleted: true}) {
		t.Fatalf("unexpected measurement: %#v", e)
	} else if e := itr.Next(); !reflect.DeepEqual(e, &tsi1.MeasurementElem{Name: []byte("mem")}) {
		t.Fatalf("unexpected measurement: %#v", e)
	} else if e := itr.Next(); e != nil {
		t.Fatalf("expected eof, got: %#v", e)
	}
}

// LogFile is a test wrapper for tsi1.LogFile.
type LogFile struct {
	*tsi1.LogFile
}

// NewLogFile returns a new instance of LogFile with a temporary file path.
func NewLogFile() *LogFile {
	file, err := ioutil.TempFile("", "tsi1-log-file-")
	if err != nil {
		panic(err)
	}
	file.Close()

	f := &LogFile{LogFile: tsi1.NewLogFile()}
	f.Path = file.Name()

	return f
}

// MustOpenLogFile returns a new, open instance of LogFile. Panic on error.
func MustOpenLogFile() *LogFile {
	f := NewLogFile()
	if err := f.Open(); err != nil {
		panic(err)
	}
	return f
}

// Close closes the log file and removes it from disk.
func (f *LogFile) Close() error {
	defer os.Remove(f.Path)
	return f.LogFile.Close()
}
