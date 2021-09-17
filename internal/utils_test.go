package internal

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertTriagedScansResponseToScansList(t *testing.T) {
	triagedScansResponse := TriagedScansResponse{
		OdataContext: "http://localhost/CxWebInterface/odata/...",
		Value: []ValueOdata{
			{ID: 1, ScanID: 10001, Date: time.Now(), Scan: ValueOdataScan{1}},
			{ID: 2, ScanID: 10002, Date: time.Now(), Scan: ValueOdataScan{1}},
			{ID: 3, ScanID: 10003, Date: time.Now(), Scan: ValueOdataScan{2}},
			{ID: 4, ScanID: 10004, Date: time.Now(), Scan: ValueOdataScan{2}},
			{ID: 5, ScanID: 10005, Date: time.Now(), Scan: ValueOdataScan{3}},
		},
	}

	scansList := convertTriagedScansResponseToScansList(triagedScansResponse)

	expectedScansList := []TriagedScan{
		{1, 10001},
		{1, 10002},
		{2, 10003},
		{2, 10004},
		{3, 10005},
	}
	assert.Equal(t, expectedScansList, scansList)
}

func TestGetLastScansByProject(t *testing.T) {
	scansList := []TriagedScan{
		{1, 10001},
		{1, 10002},
		{2, 10003},
		{2, 10004},
		{3, 10005},
		{4, 10006},
		{5, 10007},
		{5, 10008},
	}

	result := getLastScansByProject(scansList)

	expected := []TriagedScan{
		{1, 10002},
		{2, 10004},
		{3, 10005},
		{4, 10006},
		{5, 10008},
	}
	assert.Equal(t, expected, result)
}

func TestGetDateFromDays(t *testing.T) {
	now := time.Date(2021, 9, 16, 17, 52, 0, 0, time.UTC)
	numDays := 10

	result := GetDateFromDays(numDays, now)

	expected := "2021-9-6"
	assert.Equal(t, expected, result)
}
