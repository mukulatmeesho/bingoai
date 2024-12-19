package history

import (
	"fmt"
	"log"
	"my-ai-assistant/constants"
	"os"
	"time"
)

func EnsureHistoryDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

func GenerateFileName(fileFor string) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf(constants.HistoryDir+"/"+fileFor+"_%s.json", timestamp)
}

func GenerateTeleFileName(fileFor string) string {
	return fmt.Sprintf(constants.HistoryDir+"/%s.json", fileFor)
}

func ListHistoryFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}
	return fileNames, nil
}
