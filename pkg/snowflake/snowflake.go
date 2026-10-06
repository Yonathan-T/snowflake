package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	workerBits   = 10
	sequenceBits = 12

	WorkerIDMask = int64(-1) ^ (int64(-1) << workerBits)
	SequenceMask = int64(-1) ^ (int64(-1) << sequenceBits)

	DefaultEpoch = 1791118800000 //october 4th, cuz that is when i started doing this project :)

	workerShift    = sequenceBits
	timestampShift = sequenceBits + workerBits
)

type Snowflake struct {
	mu            sync.Mutex
	epoch         int64
	workerID      int64
	lastTimestamp int64
	sequence      int64
}

func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > WorkerIDMask {
		return nil, fmt.Errorf("workerID must be between 0 and %d", WorkerIDMask)
	}
	return &Snowflake{
		workerID:      workerID,
		epoch:         DefaultEpoch,
		lastTimestamp: -1,
		sequence:      0,
	}, nil
}

func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()

	if now < s.lastTimestamp {
		diff := s.lastTimestamp - now
		if diff <= 5 {
			time.Sleep(time.Duration(diff) * time.Millisecond)
			now = time.Now().UnixMilli()
		}
		if now < s.lastTimestamp {
			return 0, fmt.Errorf("clock moved backwards: refusing to generate id for %d ms", s.lastTimestamp-now)
		}
	}
	if now == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & SequenceMask
		if s.sequence == 0 {
			for now <= s.lastTimestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else if now > s.lastTimestamp {
		s.sequence = 0
	}

	s.lastTimestamp = now

	id := ((now - s.epoch) << timestampShift) | (s.workerID << workerShift) | s.sequence
	return id, nil
}

func (s *Snowflake) WorkerID() int64 {
	return s.workerID
}
