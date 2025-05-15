package generator

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
)

type CopyTask struct {
	FromPath string
	ToPath   string
}

func isDirectory(path string) bool {
	// Check if the path is a directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func copyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create the destination file
	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the contents from source to destination
	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func copyDirectory(src, dst string) error {
	// Create the destination directory
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		return err
	}

	// Read the contents of the source directory
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		dstPath := filepath.Join(dst, file.Name())

		if file.IsDir() {
			// Recursively copy subdirectories
			err = copyDirectory(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			// Copy files
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t CopyTask) Execute() error {
	// If FromPAth is a directory, copy all files
	// If FromPath is a file, copy the file
	if isDirectory(t.FromPath) {
		return copyDirectory(t.FromPath, t.ToPath)
	} else {
		return copyFile(t.FromPath, t.ToPath)
	}

}

func (t CopyTask) Name() string {
	return t.ToPath
}

func (t CopyTask) Generate() template.HTML {
	return template.HTML("")
}

func (t CopyTask) GetOutputPath() string {
	return t.ToPath
}
