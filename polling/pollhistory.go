package polling

import (
	"io/ioutil"
	"os"
	"strconv"
)

type PollHistory interface {
	ReadHistory() (uint64, error)
	WriteHistory(uint64) error
}

type FileHistory struct {
	filename string
}

func NewFileHistory(filename string) *FileHistory {
	return &FileHistory{filename: filename}
}

// Read the last historyId from the file
func (f *FileHistory) ReadHistory() (uint64, error) {
	data, err := ioutil.ReadFile(f.filename)
	if err != nil {
		// If the file does not exist, return 0 and no error
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	historyId, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, err
	}
	return historyId, nil
}

// Write the last historyId to the file
func (f *FileHistory) WriteHistory(historyId uint64) error {
	return ioutil.WriteFile(f.filename, []byte(strconv.FormatUint(historyId, 10)), 0644)
}
