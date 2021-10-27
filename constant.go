package gosqlredis

const (
	MaxUint32 = ^uint32(0)
	MinUint32 = 0
	MaxInt32  = int(MaxUint32 >> 1)
	MinInt32  = -MaxInt32 - 1
	MaxUint64 = ^uint64(0)
	MinUint64 = 0
	MaxInt64  = int(MaxUint64 >> 1)
	MinInt64  = -MaxInt64 - 1
)
