package report

import (
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

const (
	similarityCalculatorCmd = "C:\\source\\SimilarityCalculator\\SimilarityCalculator\\bin\\Debug\\netcoreapp3.1\\SimilarityCalculator.exe"
)

type SimilarityCalculator interface {
	Calculate(
		filename1, name1, line1, column1, methodLine1,
		filename2, name2, line2, column2, methodLine2,
		queryID string,
	) (string, error)
}

type Similarity struct{}

func NewSimilarity() *Similarity {
	return &Similarity{}
}

func (e *Similarity) Calculate(
	filename1, name1, line1, column1, methodLine1,
	filename2, name2, line2, column2, methodLine2,
	queryID string,
) (string, error) {
	command := exec.Command(
		similarityCalculatorCmd,
		filename1, name1, line1, column1, methodLine1,
		filename2, name2, line2, column2, methodLine2,
		queryID,
	)
	out, err := command.Output()
	if err != nil {
		return "", errors.Wrap(err, "failed running command")
	}
	return strings.TrimSpace(string(out)), nil
}
