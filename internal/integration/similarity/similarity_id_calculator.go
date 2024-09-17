package similarity

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	calculatorCmd = "SimilarityCalculator.exe"
)

type IDProvider interface {
	Calculate(
		filename1, name1, line1, column1, methodLine1,
		filename2, name2, line2, column2, methodLine2,
		queryID string,
		simIDVersion int,
	) (string, error)
}

type IDCalculator struct {
	calculatorCmd string
}

func NewSimilarityIDCalculator() (*IDCalculator, error) {
	executableFilename, executableErr := os.Executable()
	if executableErr != nil {
		return nil, executableErr
	}
	executablePath := filepath.Dir(executableFilename)
	cmd := filepath.Join(executablePath, calculatorCmd)
	return &IDCalculator{calculatorCmd: cmd}, nil
}

func (e *IDCalculator) Calculate(
	filename1, name1, line1, column1, methodLine1,
	filename2, name2, line2, column2, methodLine2,
	queryID string,
	simIDVersion int,
) (string, error) {
	command := exec.Command( //nolint:gosec
		e.calculatorCmd,
		filename1, name1, line1, column1, methodLine1,
		filename2, name2, line2, column2, methodLine2,
		queryID,
		fmt.Sprint(simIDVersion),
	)
	out, err := command.Output()
	if err != nil {
		return "", errors.Wrapf(
			err,
			"failed running command file1=%s name1=%s line1=%s col1=%s method1=%s file2=%s name2=%s line2=%s col2=%s method2=%s query=%s simIDVersion=%d", //nolint:lll
			filename1, name1, line1, column1, methodLine1, filename2, name2, line2, column2, methodLine2, queryID, simIDVersion,
		)
	}
	return strings.TrimSpace(string(out)), nil
}
