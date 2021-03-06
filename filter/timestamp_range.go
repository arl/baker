package filter

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/AdRoll/baker"
)

// TimestampRangeDesc describes the TimestampRange filter.
var TimestampRangeDesc = baker.FilterDesc{
	Name:   "TimestampRange",
	New:    NewTimestampRange,
	Config: &TimestampRangeConfig{},
	Help:   "Discard records if the value of a field containing a timestamp is out of the given time range (i.e StartDateTime <= value < EndDateTime)",
}

// TimestampRangeConfig holds configuration paramters for the TimestampRange filter.
type TimestampRangeConfig struct {
	StartDatetime string `help:"Lower bound of the accepted time interval (inclusive, UTC) format:'2006-01-31 15:04:05'. Also accepts 'now'" default:"no bound" required:"true"`
	EndDatetime   string `help:"Upper bound of the accepted time interval (exclusive, UTC) format:'2006-01-31 15:04:05'. Also accepts 'now'" default:"no bound" required:"true"`
	Field         string `help:"Name of the field containing the Unix EPOCH timestamp" required:"true"`
}

// TimestampRange is a baker filter that discards records depending on the
// value of a field representing a Unix timestamp.
type TimestampRange struct {
	numProcessedLines int64
	numFilteredLines  int64
	startDate         int64
	endDate           int64

	fidx baker.FieldIndex
}

// NewTimestampRange creates and configures a TimestampRange filter.
func NewTimestampRange(cfg baker.FilterParams) (baker.Filter, error) {
	if cfg.DecodedConfig == nil {
		cfg.DecodedConfig = &TimestampRangeConfig{}
	}
	dcfg := cfg.DecodedConfig.(*TimestampRangeConfig)

	fidx, ok := cfg.FieldByName(dcfg.Field)
	if !ok {
		return nil, fmt.Errorf("unknown field %q", dcfg.Field)
	}

	f := &TimestampRange{
		fidx: fidx,
	}
	if err := f.setTimes(dcfg.StartDatetime, dcfg.EndDatetime); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *TimestampRange) setTimes(start, end string) error {
	const timeLayout = "2006-01-02 15:04:05"

	if start == "now" {
		f.startDate = time.Now().Unix()
	} else {
		t, err := time.Parse(timeLayout, start)
		if err != nil {
			return fmt.Errorf("start time is invalid: %s", err)
		}
		f.startDate = t.Unix()
	}

	if end == "now" {
		f.endDate = time.Now().Unix()
	} else {
		t, err := time.Parse(timeLayout, end)
		if err != nil {
			return fmt.Errorf("end time is invalid: %s", err)
		}
		f.endDate = t.Unix()
	}

	return nil
}

// Stats implements baker.Filter.
func (f *TimestampRange) Stats() baker.FilterStats {
	return baker.FilterStats{
		NumProcessedLines: atomic.LoadInt64(&f.numProcessedLines),
		NumFilteredLines:  atomic.LoadInt64(&f.numFilteredLines),
	}
}

// Process implements baker.Filter.
func (f *TimestampRange) Process(l baker.Record, next func(baker.Record)) {
	atomic.AddInt64(&f.numProcessedLines, 1)

	// Convert the record timestamp to unix time (int64)
	ts, err := strconv.ParseInt(string(l.Get(f.fidx)), 10, 64)
	if err != nil {
		atomic.AddInt64(&f.numFilteredLines, 1)
		return
	}

	// Discard records having an out-of-bounds timestamp.
	if ts < f.startDate || ts >= f.endDate {
		atomic.AddInt64(&f.numFilteredLines, 1)
		return
	}

	next(l)
}
