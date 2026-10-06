package snowflake

import "time"

func Deconstruct(id int64) (time.Time, int64, int64) {
	sequence := id & SequenceMask
	workerID := (id >> workerShift) & WorkerIDMask
	timestampMS := (id >> timestampShift) + DefaultEpoch

	return time.UnixMilli(timestampMS), workerID, sequence
}
