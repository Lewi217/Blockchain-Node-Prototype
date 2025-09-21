package utils

import (
	"encoding/json"
	"fmt"
	"time"
)

// FormatBalance formats balance from satoshis to readable format
func FormatBalance(satoshis int64) string {
	coins := float64(satoshis) / 100000000.0
	return fmt.Sprintf("%.8f coins (%d satoshis)", coins, satoshis)
}

// SatoshisToCoins converts satoshis to coins
func SatoshisToCoins(satoshis int64) float64 {
	return float64(satoshis) / 100000000.0
}

// CoinsToSatoshis converts coins to satoshis
func CoinsToSatoshis(coins float64) int64 {
	return int64(coins * 100000000.0)
}

// FormatTimestamp formats Unix timestamp to readable string
func FormatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05 UTC")
}

// PrettyPrint prints JSON with indentation
func PrettyPrint(v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MinInt64 returns the minimum of two int64 values
func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// MaxInt64 returns the maximum of two int64 values
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// Contains checks if a string slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Reverse reverses a slice of strings
func Reverse(slice []string) []string {
	result := make([]string, len(slice))
	for i, v := range slice {
		result[len(slice)-1-i] = v
	}
	return result
}

// IsValidAmount checks if an amount is valid (positive and not zero)
func IsValidAmount(amount int64) bool {
	return amount > 0
}

// TruncateString truncates a string to specified length with ellipsis
func TruncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}