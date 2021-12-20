package sast

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	similarityCalculatorCmd = "SimilarityCalculator.exe"
)

type SimilarityCalculator interface {
	Calculate(
		filename1, name1, line1, column1, methodLine1,
		filename2, name2, line2, column2, methodLine2,
		queryID string,
	) (string, error)
}

type Similarity struct {
	calculatorCmd string
}

func NewSimilarity() (*Similarity, error) {
	executableFilename, executableErr := os.Executable()
	if executableErr != nil {
		return nil, executableErr
	}
	executablePath := filepath.Dir(executableFilename)
	calculatorCmd := filepath.Join(executablePath, similarityCalculatorCmd)
	return &Similarity{calculatorCmd: calculatorCmd}, nil
}

func (e *Similarity) Calculate(
	filename1, name1, line1, column1, methodLine1,
	filename2, name2, line2, column2, methodLine2,
	queryID string,
) (string, error) {
	command := exec.Command( //nolint:gosec
		e.calculatorCmd,
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
