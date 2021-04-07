package cloudflare

import "errors"

var (
	E001 = errors.New("failed to decode zones list")
	E002 = errors.New("no matching zones found")
	E003 = errors.New("failed to decode records")
	E004 = errors.New("failed to decode record update")
	E005 = errors.New("no records returned")
	E006 = errors.New("failed to build request")
	E007 = errors.New("failed to get records")
	E008 = errors.New("failed to decode record creation")
	E009 = errors.New("failed to create record")
	E010 = errors.New("failed to zones list")
	E011 = errors.New("too many records found")
	E012 = errors.New("failed to update record")
)

