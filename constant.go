// Copyright 2021 by Lukisno Kaharman. All rights reserved.
// This Source Code Form is subject to the terms of the Apache
// License 2.0 that can be found in the LICENSE file.

package gosqlredis

const (
	// MaxUint32 maximum value for uint32
	MaxUint32 = ^uint32(0)

	// MinUint32 minimum value for uint32
	MinUint32 = 0

	// MaxInt32 maximum value for int32
	MaxInt32 = int(MaxUint32 >> 1)

	// MinInt32 minimum value for int32
	MinInt32 = -MaxInt32 - 1

	// MaxUint64 maximum value for uint64
	MaxUint64 = ^uint64(0)

	// MinUint64 minimum value for uint64
	MinUint64 = 0

	// MaxInt64 maximum value for int64
	MaxInt64 = int(MaxUint64 >> 1)

	// MinInt64 maximum value for int64
	MinInt64 = -MaxInt64 - 1
)
