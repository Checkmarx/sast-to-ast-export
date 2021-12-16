package report

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReport(t *testing.T) {
	report1, report1Err := ioutil.ReadFile("../../../test/data/sast/report/report1.xml")
	assert.NoError(t, report1Err)

	result, err := NewReport(report1)
	assert.NoError(t, err)
	assert.Equal(t, result.report.Queries[0].Name, "SQL_Injection")
}

func TestReport_Marshal(t *testing.T) {
	report1, report1Err := ioutil.ReadFile("../../../test/data/sast/report/report1.xml")
	assert.NoError(t, report1Err)
	report1AfterMarshal, report1AfterMarshalErr := ioutil.ReadFile("../../../test/data/sast/report/report1_after_marshal.xml")
	assert.NoError(t, report1AfterMarshalErr)

	parser, parseErr := NewReport(report1)
	assert.NoError(t, parseErr)

	result, err := parser.Marshal()
	assert.NoError(t, err)
	assert.Equal(t, string(report1AfterMarshal), string(result))
}

func TestReport_AddSimilarity(t *testing.T) {
	report1, report1Err := ioutil.ReadFile("../../../test/data/sast/report/report1.xml")
	assert.NoError(t, report1Err)
	enrichedReport1, enrichedReport1Err := ioutil.ReadFile("../../../test/data/sast/report/report1_enriched.xml")
	assert.NoError(t, enrichedReport1Err)

	parser, parseErr := NewReport(report1)
	assert.NoError(t, parseErr)

	err := parser.AddSimilarity()
	assert.NoError(t, err)

	data, marshalErr := parser.Marshal()
	assert.NoError(t, marshalErr)
	assert.Equal(t, string(enrichedReport1), string(data))
}
