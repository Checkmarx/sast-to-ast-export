package report

import (
	"io/ioutil"
	"testing"

	"github.com/checkmarxDev/ast-sast-export/internal/database"
	mock_store "github.com/checkmarxDev/ast-sast-export/test/mocks/database/store"
	mock_report "github.com/checkmarxDev/ast-sast-export/test/mocks/sast/report"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestReport_Parse(t *testing.T) {
	report1, report1Err := ioutil.ReadFile("../../../test/data/sast/report/report1.xml")
	assert.NoError(t, report1Err)
	ctrl := gomock.NewController(t)
	sourceMock := mock_report.NewMockSourceProvider(ctrl)
	nodeResultsMock := mock_store.NewMockNodeResultsStore(ctrl)
	similarityCalculatorMock := mock_report.NewMockSimilarityCalculator(ctrl)

	parser := NewReport(sourceMock, nodeResultsMock, similarityCalculatorMock)
	parseErr := parser.Parse(report1)
	assert.NoError(t, parseErr)
	assert.Equal(t, parser.report.Queries[0].Name, "SQL_Injection")
}

func TestReport_Marshal(t *testing.T) {
	report1, report1Err := ioutil.ReadFile("../../../test/data/sast/report/report1.xml")
	assert.NoError(t, report1Err)
	report1AfterMarshal, report1AfterMarshalErr := ioutil.ReadFile("../../../test/data/sast/report/report1_after_marshal.xml")
	assert.NoError(t, report1AfterMarshalErr)
	ctrl := gomock.NewController(t)
	sourceMock := mock_report.NewMockSourceProvider(ctrl)
	nodeResultsMock := mock_store.NewMockNodeResultsStore(ctrl)
	similarityCalculatorMock := mock_report.NewMockSimilarityCalculator(ctrl)

	parser := NewReport(sourceMock, nodeResultsMock, similarityCalculatorMock)
	parseErr := parser.Parse(report1)
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
	ctrl := gomock.NewController(t)
	sourceMock := mock_report.NewMockSourceProvider(ctrl)
	sourceMock.EXPECT().GetBasePath("1000002").Return("C:\\source", nil).AnyTimes()
	nodeResultsMock := mock_store.NewMockNodeResultsStore(ctrl)
	nodeResultsMock.EXPECT().GetByResultPathAndNode("1000002", "2", 1).Return(&database.NodeResult{MethodLine: 1}, nil).AnyTimes()
	nodeResultsMock.EXPECT().GetByResultPathAndNode("1000002", "2", 3).Return(&database.NodeResult{MethodLine: 2}, nil).AnyTimes()
	nodeResultsMock.EXPECT().GetByResultPathAndNode("1000002", "3", 1).Return(&database.NodeResult{MethodLine: 3}, nil).AnyTimes()
	nodeResultsMock.EXPECT().GetByResultPathAndNode("1000002", "3", 5).Return(&database.NodeResult{MethodLine: 4}, nil).AnyTimes()
	nodeResultsMock.EXPECT().GetByResultPathAndNode("1000002", "6", 1).Return(&database.NodeResult{MethodLine: 5}, nil).AnyTimes()
	similarityCalculatorMock := mock_report.NewMockSimilarityCalculator(ctrl)
	similarityCalculatorMock.EXPECT().Calculate(
		"C:\\source\\Goatlin-develop\\packages\\clients\\android\\app\\src\\main\\java\\com\\cx\\goatlin\\EditNoteActivity.kt", gomock.Any(), gomock.Any(), gomock.Any(), "1", //nolint:lll
		"C:\\source\\Goatlin-develop\\packages\\clients\\android\\app\\src\\main\\java\\com\\cx\\goatlin\\helpers\\DatabaseHelper.kt", gomock.Any(), gomock.Any(), gomock.Any(), "2", //nolint:lll
		gomock.Any(),
	).Return("12000000", nil)
	similarityCalculatorMock.EXPECT().Calculate(
		"C:\\source\\Goatlin-develop\\packages\\clients\\android\\app\\src\\main\\java\\com\\cx\\goatlin\\EditNoteActivity.kt", gomock.Any(), gomock.Any(), gomock.Any(), "3", //nolint:lll
		"C:\\source\\Goatlin-develop\\packages\\clients\\android\\app\\src\\main\\java\\com\\cx\\goatlin\\helpers\\DatabaseHelper.kt", gomock.Any(), gomock.Any(), gomock.Any(), "4", //nolint:lll
		gomock.Any(),
	).Return("34000000", nil)
	similarityCalculatorMock.EXPECT().Calculate(
		"C:\\source\\Goatlin-develop\\packages\\services\\api\\src\\app.js", gomock.Any(), gomock.Any(), gomock.Any(), "5",
		"C:\\source\\Goatlin-develop\\packages\\services\\api\\src\\app.js", gomock.Any(), gomock.Any(), gomock.Any(), "5",
		gomock.Any(),
	).Return("55000000", nil)

	parser := NewReport(sourceMock, nodeResultsMock, similarityCalculatorMock)
	parseErr := parser.Parse(report1)
	assert.NoError(t, parseErr)

	err := parser.AddSimilarity()
	assert.NoError(t, err)

	data, marshalErr := parser.Marshal()
	assert.NoError(t, marshalErr)
	assert.Equal(t, string(enrichedReport1), string(data))
}
