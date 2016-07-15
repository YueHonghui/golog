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
	ErrLevelInvalid = errors.New("Level invalid. Level must be TRC|DBG|INF|WRN|ERR")
)

var (
	glock     sync.RWMutex
	glevel    int = LevelTRC
	gnewline  string
	gwriter   io.WriteCloser
	glogFatal *log.Logger
	glogERR   *log.Logger
	glogWRN   *log.Logger
	glogINF   *log.Logger
	glogDBG   *log.Logger
	glogTRC   *log.Logger
)

func ParseLevel(lv string) (int, bool) {
	if lv == "TRC" {
		return LevelTRC, true
	} else if lv == "DBG" {
		return LevelDBG, true
	} else if lv == "INF" {
		return LevelINF, true
	} else if lv == "WRN" {
		return LevelWRN, true
	} else if lv == "ERR" {
		return LevelERR, true
	} else {
		return LevelTRC, false
	}
}

func init() {
	glock.Lock()
	defer glock.Unlock()
	if runtime.GOOS == "windows" {
		gnewline = "\r\n"
	} else {
		gnewline = "\n"
	}
	glogFatal = log.New(os.Stderr, "[FAT] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogERR = log.New(os.Stdout, "[ERR] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogWRN = log.New(os.Stdout, "[WRN] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogINF = log.New(os.Stdout, "[INF] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogDBG = log.New(os.Stdout, "[DBG] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogTRC = log.New(os.Stdout, "[TRC] ", log.Ldate|log.Ltime|log.Lshortfile)
}

func parseLogUrl(url string) (schema, uri string, keyvalues map[string]string, err error) {
	keyvalues = make(map[string]string)
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
//	file:///home/logs/demo,[rotate=day|none],[level=TRC|DBG|INF|WRN|ERR]
func Init(logurl string) (err error) {
	var uri string
	var kvs map[string]string
	_, uri, kvs, err = parseLogUrl(logurl)
	if err != nil {
		return
	}
	glock.Lock()
	defer glock.Unlock()
	if gwriter != nil {
		glogERR.Fatalf("InitLogger must called only one time%s", gnewline)
		return
	}
	if rt, ok := kvs["rotate"]; ok && rt == "day" {
		gwriter, err = rfw.New(uri)
		if err != nil {
			return
		}
	} else {
		gwriter, err = os.OpenFile(uri, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
		if err != nil {
			return
		}
	}
	if lv, ok := kvs["level"]; ok {
		level, valid := ParseLevel(lv)
		if !valid {
			gwriter.Close()
			return ErrLevelInvalid
		}
		glevel = level
	}
	glogFatal = log.New(gwriter, "[FAT] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogERR = log.New(gwriter, "[ERR] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogWRN = log.New(gwriter, "[WRN] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogINF = log.New(gwriter, "[INF] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogDBG = log.New(gwriter, "[DBG] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogTRC = log.New(gwriter, "[TRC] ", log.Ldate|log.Ltime|log.Lshortfile)
	return
}

func SetLevel(level int) {
	glock.Lock()
	defer glock.Unlock()
	glevel = level
}

func GetLevel() int {
	return glevel
}

func TRC(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelTRC {
		glogTRC.Output(2, fmt.Sprintf(format+gnewline, v...))
	}
}

func DBG(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelDBG {
		glogDBG.Output(2, fmt.Sprintf(format+gnewline, v...))
	}
}

func INF(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelINF {
		glogINF.Output(2, fmt.Sprintf(format+gnewline, v...))
	}
}

func WRN(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelWRN {
		glogWRN.Output(2, fmt.Sprintf(format+gnewline, v...))
	}
}

func ERR(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelERR {
		glogERR.Output(2, fmt.Sprintf(format+gnewline, v...))
	}
}

func FAT(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	glogFatal.Output(2, fmt.Sprintf(format+gnewline, v...))
	os.Exit(1)
}

func TRCf(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelTRC {
		glogTRC.Output(2, fmt.Sprintf(format, v...))
	}
}

func DBGf(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelDBG {
		glogDBG.Output(2, fmt.Sprintf(format, v...))
	}
}

func INFf(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelINF {
		glogINF.Output(2, fmt.Sprintf(format, v...))
	}
}

func WRNf(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelWRN {
		glogWRN.Output(2, fmt.Sprintf(format, v...))
	}
}

func ERRf(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	if glevel <= LevelERR {
		glogERR.Output(2, fmt.Sprintf(format, v...))
	}
}
func Fatalf(format string, v ...interface{}) {
	glock.RLock()
	defer glock.RUnlock()
	glogFatal.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Fini() {
	glock.Lock()
	defer glock.Unlock()
	if gwriter != nil {
		gwriter.Close()
		gwriter = nil
	}
	glogFatal = log.New(os.Stderr, "[FAT] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogERR = log.New(os.Stdout, "[ERR] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogWRN = log.New(os.Stdout, "[WRN] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogINF = log.New(os.Stdout, "[INF] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogDBG = log.New(os.Stdout, "[DBG] ", log.Ldate|log.Ltime|log.Lshortfile)
	glogTRC = log.New(os.Stdout, "[TRC] ", log.Ldate|log.Ltime|log.Lshortfile)
}
