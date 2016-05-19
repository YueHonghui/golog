package golog

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/YueHonghui/rfw"
)

const (
	LevelTRC = iota
	LevelDBG
	LevelINF
	LevelWRN
	LevelERR
)

var (
	lock     sync.RWMutex
	level    int = LevelTRC
	newline  string
	writer   io.WriteCloser
	logFatal *log.Logger
	logERR   *log.Logger
	logWRN   *log.Logger
	logINF   *log.Logger
	logDBG   *log.Logger
	logTRC   *log.Logger
)

func init() {
	lock.Lock()
	defer lock.Unlock()
	if runtime.GOOS == "windows" {
		newline = "\r\n"
	} else {
		newline = "\n"
	}
	logFatal = log.New(os.Stderr, "[FAT]", log.Ldate|log.Ltime|log.Lshortfile)
	logERR = log.New(os.Stdout, "[ERR]", log.Ldate|log.Ltime|log.Lshortfile)
	logWRN = log.New(os.Stdout, "[WRN]", log.Ldate|log.Ltime|log.Lshortfile)
	logINF = log.New(os.Stdout, "[INF]", log.Ldate|log.Ltime|log.Lshortfile)
	logDBG = log.New(os.Stdout, "[DBG]", log.Ldate|log.Ltime|log.Lshortfile)
	logTRC = log.New(os.Stdout, "[TRC]", log.Ldate|log.Ltime|log.Lshortfile)
}

func parseLogUrl(url string) (schema, uri string, keyvalues map[string]string, err error) {
	itms := strings.Split(url, ",")
	schitms := strings.Split(itms[0], "://")
	if len(schitms) != 2 {
		err = errors.New(fmt.Sprintf("logurl invalid: %s", url))
		return
	}
	if schitms[0] != "file" {
		err = errors.New(fmt.Sprintf("schema %s in logurl %s not supported yet", schitms[0], url))
		return
	}
	schema = schitms[0]
	uri = schitms[1]
	for _, v := range itms[1:] {
		kvs := strings.Split(v, "=")
		if len(kvs) != 2 {
			err = errors.New(fmt.Sprintf("keyvalue %s in logurl %s invalid", v, url))
			return
		}
		keyvalues[kvs[0]] = kvs[1]
	}
	return
}

//@logurl
//    logurl used to determin to which place and how the log will be write, generate format is "[schema]://[uri],[key=value]...", for example
//	file:///home/logs/demo,[rotate=day|none]
func Init(logurl string) (err error) {
	var uri string
	var kvs map[string]string
	_, uri, kvs, err = parseLogUrl(logurl)
	if err != nil {
		return
	}
	lock.Lock()
	defer lock.Unlock()
	if writer != nil {
		logERR.Fatalf("InitLogger must called only one time%s", newline)
		return
	}
	if rt, ok := kvs["rotate"]; ok && rt == "day" {
		writer, err = rfw.New(uri)
		if err != nil {
			return
		}
	} else {
		writer, err = os.OpenFile(uri, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			return
		}
	}
	logFatal = log.New(writer, "[FAT]", log.Ldate|log.Ltime|log.Lshortfile)
	logERR = log.New(writer, "[ERR]", log.Ldate|log.Ltime|log.Lshortfile)
	logWRN = log.New(writer, "[WRN]", log.Ldate|log.Ltime|log.Lshortfile)
	logINF = log.New(writer, "[INF]", log.Ldate|log.Ltime|log.Lshortfile)
	logDBG = log.New(writer, "[DBG]", log.Ldate|log.Ltime|log.Lshortfile)
	logTRC = log.New(writer, "[TRC]", log.Ldate|log.Ltime|log.Lshortfile)
	return
}

func SetLevel(level int) {
	lock.Lock()
	defer lock.Unlock()
	level = level
}

func GetLevel() int {
	return level
}

func TRC(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelTRC {
		logTRC.Printf(format, v...)
		logTRC.Print(newline)
	}
}

func DBG(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelDBG {
		logDBG.Printf(format, v...)
		logDBG.Print(newline)
	}
}

func INF(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelINF {
		logINF.Printf(format, v...)
		logINF.Print(newline)
	}
}

func WRN(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelWRN {
		logWRN.Printf(format, v...)
		logWRN.Print(newline)
	}
}

func ERR(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelERR {
		logERR.Printf(format, v...)
		logERR.Print(newline)
	}
}

func Fatal(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelERR {
		logFatal.Printf(format, v...)
		logFatal.Fatal(newline)
	}
}

func TRCf(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelTRC {
		logTRC.Printf(format, v...)
	}
}

func DBGf(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelDBG {
		logDBG.Printf(format, v...)
	}
}

func INFf(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelINF {
		logINF.Printf(format, v...)
	}
}

func WRNf(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelWRN {
		logWRN.Printf(format, v...)
	}
}

func ERRf(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelERR {
		logERR.Printf(format, v...)
	}
}
func Fatalf(format string, v ...interface{}) {
	lock.RLock()
	defer lock.RUnlock()
	if level <= LevelERR {
		logFatal.Fatalf(format, v...)
	}
}

func Fini() {
	lock.Lock()
	defer lock.Unlock()
	if writer != nil {
		writer.Close()
		writer = nil
	}
	logFatal = log.New(os.Stderr, "[FAT]", log.Ldate|log.Ltime|log.Lshortfile)
	logERR = log.New(os.Stdout, "[ERR]", log.Ldate|log.Ltime|log.Lshortfile)
	logWRN = log.New(os.Stdout, "[WRN]", log.Ldate|log.Ltime|log.Lshortfile)
	logINF = log.New(os.Stdout, "[INF]", log.Ldate|log.Ltime|log.Lshortfile)
	logDBG = log.New(os.Stdout, "[DBG]", log.Ldate|log.Ltime|log.Lshortfile)
	logTRC = log.New(os.Stdout, "[TRC]", log.Ldate|log.Ltime|log.Lshortfile)
}
