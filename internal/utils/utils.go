package utils

import (
	"fmt"
	"time"
)

func GenerateQRCode(visitorID, societyCode string) string {
	// Generate a unique QR code based on visitor ID, society code and current time
	timestamp := time.Now().Unix()
	return fmt.Sprintf("BMS-%s-%s-%d", societyCode, visitorID, timestamp)
}