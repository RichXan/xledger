package timex

import "time"

func ResolveUserTZ(value string) *time.Location {
	loc, ok := ParseUserTZ(value)
	if ok {
		return loc
	}
	return time.FixedZone("UTC+8", 8*60*60)
}

func ParseUserTZ(value string) (*time.Location, bool) {
	if value != "" {
		if loc, err := time.LoadLocation(value); err == nil {
			return loc, true
		}
		return nil, false
	}
	return time.FixedZone("UTC+8", 8*60*60), true
}
