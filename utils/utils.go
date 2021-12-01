package utils

import "github.com/spf13/viper"

// EqualError reports whether errors a and b are considered equal.
// They're equal if both are nil, or both are not nil and a.Error() == b.Error().
func EqualError(a, b error) bool {
	return a == nil && b == nil || a != nil && b != nil && a.Error() == b.Error()
}

func InitViper() {
	viper.Set("konfActiveList", "./konf/active")
	viper.Set("konfStore", "./konf/store")
}
