package casting

import (
	"strconv"
)

func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func StringToUint(s string) (uint, error) {
	u64, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(u64), nil
}

func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}
