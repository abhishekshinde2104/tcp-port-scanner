package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckExt returns an error unless path ends in .json or .csv.
func CheckExt(path string) error {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json", ".csv":
		return nil
	default:
		return fmt.Errorf("unsupported output file %q (use .json or .csv)", path)
	}
}

// Save writes rep to path as JSON or CSV, chosen by the file extension.
func Save(path string, rep Report) error {
	if err := CheckExt(path); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if strings.ToLower(filepath.Ext(path)) == ".json" {
		return JSON(f, rep)
	}
	return CSV(f, rep.Ports)
}
