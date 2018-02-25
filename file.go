package log

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type FileConfig struct {
	FileName    string
	DailyRotate bool // Rotate daily
	MaxDays     uint16
	MaxSize     uint32
	Perm        string
}

type FileStore struct {
	sync.RWMutex   // write log order by order
	Filename       string
	fileNamePrefix string   // file name prefix, if change file output, only add date information
	fp             *os.File // The opened file
	DailyRotate    bool     // change file output everyday
	MaxDays        uint16   // if log files exist more than max days, delete log files
	MaxSize        uint32   // unit: byte
	nowSize        uint32   // unit: byte
	Perm           string
}

func NewFileStore(cfg *FileConfig) (Store, error) {
	if len(cfg.FileName) == 0 {
		return nil, errors.New("filename empty")
	}
	if filepath.Ext(cfg.FileName) == "" {
		return nil, errors.New("filename must in *.* format")
	}
	f := &FileStore{}
	f.Filename = cfg.FileName
	f.DailyRotate = cfg.DailyRotate
	f.MaxDays = cfg.MaxDays
	f.MaxSize = cfg.MaxSize
	f.Perm = "0660" //default is 0660
	f.fileNamePrefix = strings.TrimSuffix(f.Filename, filepath.Ext(f.Filename))
	return f, nil
}

func (w *FileStore) Init() error {
	err := w.startLogger()
	return err
}

// start file logger. create log file and set to locker-inside file writer.
func (w *FileStore) startLogger() (err error) {
	//file, err := w.createLogFile()

	perm, err := strconv.ParseInt(w.Perm, 8, 64)
	if err != nil {
		return err
	}

	fp, err := os.OpenFile(w.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err == nil {
		// Make sure file perm is user set perm cause of `os.OpenFile` will obey umask
		os.Chmod(w.Filename, os.FileMode(perm))
	}

	fInfo, _ := fp.Stat()

	// record size now
	w.nowSize = uint32(fInfo.Size())

	if err != nil {
		return err
	}

	if w.fp != nil {
		w.fp.Close()
	}
	w.fp = fp

	return w.initLogFile()
}

// initialize log file
func (w *FileStore) initLogFile() (err error) {
	fd := w.fp
	fInfo, err := fd.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s", err)
	}

	//if need a new log file daily
	if w.DailyRotate {
		go w.run() // check is need rename timer task
	}

	// if this file contains data
	if fInfo.Size() > 0 {
	}
	return nil
}

// WriteMsg write logger message into file.
func (w *FileStore) WriteMsg(s *string) error {
	w.Lock()
	w.nowSize += uint32(len(*s))
	fmt.Fprintf(os.Stderr, "nowSize %d, maxsize: %d\n", w.nowSize, w.MaxSize)
	w.checkFileSizeRotate()
	_, err := w.fp.Write([]byte(*s))
	w.Unlock()
	return err
}

func (w *FileStore) run() {
	timer := time.NewTimer(time.Minute)
	for {
		select {

		case <-timer.C:
			timer.Reset(time.Minute)
			if hour, _, _ := time.Now().Clock(); hour == 0 {
				w.Lock()
				if w.DailyRotate {
					if err := w.rename(); err != nil {
						fmt.Fprintf(os.Stderr, "fileLogger(%q): %s\n", w.Filename, err)
					}
				}
				w.Unlock()
			}

		}
	}

}

func (w *FileStore) checkFileSizeRotate() {
	// check size
	if w.nowSize > w.MaxSize {
		if err := w.rename(); err != nil {
			fmt.Fprintf(os.Stderr, "fileName( %s ) err: %s\n", w.Filename, err)
		}
	}
}

// check if need rename a file and start a new file
func (w *FileStore) rename() error {
	// Find the next available number
	num := 1
	newFileName := ""
	isNewFileNameAvailable := false
	logTime := time.Now()
	_, err := os.Lstat(w.Filename)
	if err != nil {
		//even if the file is not exist or other ,we should RESTART the logger
		goto RESTART_LOGGER
	}

	// according to maxLines setting, generate a new file name

	for ; err == nil && num <= 9999; num++ {

		newFileName = w.fileNamePrefix + fmt.Sprintf("_%s_%03d%s", logTime.Format("2006-01-02"), num, ".log")

		// if err appeared, it means that newFileName is available
		_, err = os.Lstat(newFileName)
		if err != nil {
			isNewFileNameAvailable = true
			break
		}

	}

	if !isNewFileNameAvailable {
		return fmt.Errorf("cannot find free log number to rename %s", w.Filename)
	}

	// close file before rename
	w.fp.Close()

	// Rename the file
	// even if occurs error,we MUST guarantee to  restart new logger
	err = os.Rename(w.Filename, newFileName)
	if err != nil {
		goto RESTART_LOGGER
	}
	err = os.Chmod(newFileName, os.FileMode(0440))
	if err != nil {
		return fmt.Errorf("Chmod err: %s", err)
	}

	// restart logger
RESTART_LOGGER:

	startLoggerErr := w.startLogger()
	go w.deleteExpiredFile()

	if startLoggerErr != nil {
		return fmt.Errorf("startLogger err: %s", startLoggerErr)
	}

	return nil
}

// exec this every day
func (w *FileStore) deleteExpiredFile() {
	dir := filepath.Dir(w.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Unable to delete old log '%s', error: %v\n", path, r)
			}
		}()

		if info == nil {
			return
		}

		//delete log file if create time+maxDays  > now time
		if !info.IsDir() && info.ModTime().Add(24*time.Hour*time.Duration(w.MaxDays)).Before(time.Now()) {
			//check if the prefix is matched
			if strings.HasPrefix(filepath.Base(path), filepath.Base(w.fileNamePrefix)) {
				os.Remove(path)
			}
		}
		return
	})
}

// Destroy close the file description, close file writer.
func (w *FileStore) Destroy() {
	w.fp.Close()
}

// flush file means sync file from disk.
func (w *FileStore) Flush() {
	w.fp.Sync()
}
