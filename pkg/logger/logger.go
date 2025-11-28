package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// Уровень для вывода в консоль
	currentLogLevel     LogLevel = INFO
	currentFileLogLevel LogLevel = INFO

	// Раздельные логгеры для разных назначений
	consoleLogger   *log.Logger = log.New(os.Stdout, "", 0)
	debugFileLogger *log.Logger = log.New(io.Discard, "", 0)
	infoFileLogger  *log.Logger = log.New(io.Discard, "", 0)
	warnFileLogger  *log.Logger = log.New(io.Discard, "", 0)

	// Файловые дескрипторы для ротации
	currentDebugFile *os.File
	currentInfoFile  *os.File
	currentWarnFile  *os.File
	currentLogHour   string
	logMutex         = &sync.Mutex{}

	// ПЕРЕМЕННЫЕ ДЛЯ УПРАВЛЕНИЯ ЗАПИСЬЮ В ФАЙЛЫ
	fileLoggingEnabled bool   // Флаг, разрешающий запись в файлы
	logDirectory       string // Имя подпапки для логов текущего режима
)

// LogLevel определяет уровень логирования.
type LogLevel int

const (
	DEBUG LogLevel = iota // 0
	INFO                  // 1
	WARN                  // 2
	ERROR                 // 3
	FATAL                 // 4
)

// String возвращает строковое представление уровня логирования.
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// InitLogger инициализирует все логгеры.
func InitLogger(level LogLevel, fileLogLevel LogLevel) {
	logMutex.Lock()
	defer logMutex.Unlock()

	currentLogLevel = level
	currentFileLogLevel = fileLogLevel

	// Логгеры УЖЕ существуют. Мы просто сообщаем об изменении уровней.
	// Используем consoleLogger, который гарантированно не nil.
	consoleLogger.Printf("INFO: (Консоль) Минимальный уровень логирования установлен: %s", currentLogLevel)
	consoleLogger.Printf("INFO: (Файлы)    Минимальный уровень логирования установлен: %s", currentFileLogLevel)
}

// EnableFileLogging активирует запись логов в файлы для конкретного режима.
// subDir - это имя уникальной папки, например "real_trading" или "backtest".
func EnableFileLogging(subDir string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	fileLoggingEnabled = true
	logDirectory = subDir

	// Логируем сам факт включения, чтобы это было видно и в консоли, и в файле
	msg := fmt.Sprintf("INFO: Запись логов в файлы ВКЛЮЧЕНА. Директория: logs/%s", subDir)
	consoleLogger.Println(formatLogPrefix("INFO") + msg)

	// Принудительно обновляем файлы сразу после включения
	updateLogFilesUnsafe() // Используем версию без блокировки, т.к. мьютекс уже захвачен
	infoFileLogger.Println(msg)
	debugFileLogger.Println(msg)
	warnFileLogger.Println(msg) // Пишем и в новый лог
}

// CloseFileLogs отключает запись, закрывает файловые дескрипторы и сбрасывает состояние.
func CloseFileLogs() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if !fileLoggingEnabled {
		return
	}

	msg := "INFO: Запись логов в файлы остановлена, файлы закрыты."
	consoleLogger.Println(formatLogPrefix("INFO") + msg)
	// Записываем прощальное сообщение во все файлы
	if infoFileLogger != nil {
		infoFileLogger.Println(msg)
	}
	if debugFileLogger != nil {
		debugFileLogger.Println(msg)
	}
	if warnFileLogger != nil {
		warnFileLogger.Println(msg)
	}

	// Закрываем все файловые дескрипторы
	closeFiles()

	fileLoggingEnabled = false
	logDirectory = ""
	currentLogHour = ""

	// Сбрасываем вывод логгеров в "никуда"
	debugFileLogger.SetOutput(io.Discard)
	infoFileLogger.SetOutput(io.Discard)
	warnFileLogger.SetOutput(io.Discard)
}

// updateLogFilesUnsafe обновляет все лог-файлы, если наступил новый час.
func updateLogFilesUnsafe() {
	if !fileLoggingEnabled {
		return
	}

	hourFormat := "2006-01-02_15"
	currentHour := time.Now().Format(hourFormat)

	if currentHour == currentLogHour {
		return
	}

	// Закрываем старые файлы
	closeFiles()

	// --- DEBUG ---
	debugDir := fmt.Sprintf("logs/%s/debug", logDirectory)
	if err := os.MkdirAll(debugDir, 0755); err != nil {
		log.Fatalf("FATAL: Не удалось создать директорию %s: %v", debugDir, err)
	}
	debugFileName := fmt.Sprintf("%s/%s.log", debugDir, currentHour)
	// --- 2. ДОБАВИТЬ ЭТОТ БЛОК ДЛЯ ДИАГНОСТИКИ ---
	absDebugPath, errPath := filepath.Abs(debugFileName)
	if errPath != nil {
		consoleLogger.Printf("WARN: (Logger) Не удалось получить абс. путь для %s: %v", debugFileName, errPath)
	} else {
		// ВОТ ЭТА СТРОКА ПОКАЖЕТ, ГДЕ ИСКАТЬ ФАЙЛ
		consoleLogger.Printf("INFO: (Logger) Открываю DEBUG файл: %s", absDebugPath)
	}
	// --- КОНЕЦ БЛОКА ДИАГNOСТИКИ ---
	debugFile, err := os.OpenFile(debugFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("FATAL: Не удалось открыть файл лога %s: %v", debugFileName, err)
	}
	currentDebugFile = debugFile
	debugFileLogger.SetOutput(currentDebugFile)

	// --- INFO+ ---
	infoDir := fmt.Sprintf("logs/%s/info+", logDirectory)
	if err := os.MkdirAll(infoDir, 0755); err != nil {
		log.Fatalf("FATAL: Не удалось создать директорию %s: %v", infoDir, err)
	}
	infoFileName := fmt.Sprintf("%s/%s.log", infoDir, currentHour)
	infoFile, err := os.OpenFile(infoFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("FATAL: Не удалось открыть файл лога %s: %v", infoFileName, err)
	}
	currentInfoFile = infoFile
	infoFileLogger.SetOutput(currentInfoFile)

	// --- НОВЫЙ БЛОК: WARN+ ---
	warnDir := fmt.Sprintf("logs/%s/warn+", logDirectory)
	if err := os.MkdirAll(warnDir, 0755); err != nil {
		log.Fatalf("FATAL: Не удалось создать директорию %s: %v", warnDir, err)
	}
	warnFileName := fmt.Sprintf("%s/%s.log", warnDir, currentHour)
	warnFile, err := os.OpenFile(warnFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("FATAL: Не удалось открыть файл лога %s: %v", warnFileName, err)
	}
	currentWarnFile = warnFile
	warnFileLogger.SetOutput(currentWarnFile)

	currentLogHour = currentHour
}

// updateLogFiles теперь просто обертка с мьютексом
func updateLogFiles() {
	logMutex.Lock()
	defer logMutex.Unlock()
	updateLogFilesUnsafe()
}

// formatLogPrefix генерирует префикс для сообщения лога.
func formatLogPrefix(level string) string {
	return fmt.Sprintf("%s %s: ", time.Now().Format("2006/01/02 15:04:05.0"), level)
}

// Debug пишет только в debug-файл и, если уровень позволяет, в консоль.
func Debugf(format string, v ...interface{}) {
	// 1. Определяем, куда нужно писать
	writeToConsole := currentLogLevel <= DEBUG
	// Проверка: Пишем в файл, если МИН. УРОВЕНЬ ФАЙЛА (current) <= УРОВНЮ СООБЩЕНИЯ (DEBUG)
	writeToFile := currentFileLogLevel <= DEBUG

	if !writeToConsole && !writeToFile {
		return // Ничего не делаем
	}

	// 2. Форматируем (только если нужно)
	msg := formatLogPrefix("DEBUG") + fmt.Sprintf(format, v...)

	// 3. Пишем в консоль
	if writeToConsole {
		consoleLogger.Println(msg)
	}

	// 4. Пишем в файл
	if writeToFile {
		updateLogFiles() // Проверяем ротацию
		debugFileLogger.Println(msg)
	}
}

// Info пишет в debug и info файлы и, если уровень позволяет, в консоль.
func Infof(format string, v ...interface{}) {
	// 1. Определяем, куда нужно писать
	writeToConsole := currentLogLevel <= INFO
	// Проверка: Пишем в файл, если МИН. УРОВЕНЬ ФАЙЛА (current) <= УРОВНЮ СООБЩЕНИЯ (INFO)
	writeToFile := currentFileLogLevel <= INFO

	if !writeToConsole && !writeToFile {
		return // Ничего не делаем
	}

	// 2. Форматируем (только если нужно)
	msg := formatLogPrefix("INFO") + fmt.Sprintf(format, v...)

	// 3. Пишем в консоль
	if writeToConsole {
		consoleLogger.Println(msg)
	}

	// 4. Пишем в файл
	if writeToFile {
		updateLogFiles()
		debugFileLogger.Println(msg)
		infoFileLogger.Println(msg)
	}
}

// Warn пишет во все файлы и, если уровень позволяет, в консоль.
func Warnf(format string, v ...interface{}) {
	// 1. Определяем, куда нужно писать
	writeToConsole := currentLogLevel <= WARN
	// Проверка: Пишем в файл, если МИН. УРОВЕНЬ ФАЙЛА (current) <= УРОВНЮ СООБЩЕНИЯ (WARN)
	writeToFile := currentFileLogLevel <= WARN

	if !writeToConsole && !writeToFile {
		return // Ничего не делаем
	}

	// 2. Форматируем (только если нужно)
	msg := formatLogPrefix("WARN") + fmt.Sprintf(format, v...)

	// 3. Пишем в консоль
	if writeToConsole {
		consoleLogger.Println(msg)
	}

	// 4. Пишем в файл
	if writeToFile {
		updateLogFiles()
		debugFileLogger.Println(msg)
		infoFileLogger.Println(msg)
		warnFileLogger.Println(msg)
	}
}

// Error пишет в оба файла и, если уровень позволяет, в консоль.
func Errorf(format string, v ...interface{}) {
	// 1. Определяем, куда нужно писать
	writeToConsole := currentLogLevel <= ERROR
	// Проверка: Пишем в файл, если МИН. УРОВЕНЬ ФАЙЛА (current) <= УРОВНЮ СООБЩЕНИЯ (ERROR)
	writeToFile := currentFileLogLevel <= ERROR

	if !writeToConsole && !writeToFile {
		return // Ничего не делаем
	}

	// 2. Форматируем (только если нужно)
	msg := formatLogPrefix("ERROR") + fmt.Sprintf(format, v...)

	// 3. Пишем в консоль
	if writeToConsole {
		consoleLogger.Println(msg)
	}

	// 4. Пишем в файл
	if writeToFile {
		updateLogFiles()
		debugFileLogger.Println(msg)
		infoFileLogger.Println(msg)
		warnFileLogger.Println(msg)
	}
}

// Fatal пишет во все места и завершает программу.
// Fatal пишет во все места и завершает программу.
func Fatalf(format string, v ...interface{}) {
	msg := formatLogPrefix("FATAL") + fmt.Sprintf(format, v...)

	logMutex.Lock()
	defer logMutex.Unlock()

	// Пишем везде
	consoleLogger.Println(msg)
	if infoFileLogger != nil {
		infoFileLogger.Println(msg)
	}
	if debugFileLogger != nil {
		debugFileLogger.Println(msg)
	}
	if warnFileLogger != nil { // ДОБАВЛЕНО
		warnFileLogger.Println(msg)
	}

	// Принудительно закрываем файлы
	closeFiles()

	os.Exit(1)
}

// closeFiles — внутренняя функция для закрытия всех открытых файлов логов.
func closeFiles() {
	if currentDebugFile != nil {
		currentDebugFile.Close()
		currentDebugFile = nil
	}
	if currentInfoFile != nil {
		currentInfoFile.Close()
		currentInfoFile = nil
	}
	if currentWarnFile != nil { // ДОБАВЛЕНО
		currentWarnFile.Close()
		currentWarnFile = nil
	}
}

// Debug обеспечивает обратную совместимость с logger.Debug(msg...)
func Debug(v ...interface{}) {
	Debugf("%s", fmt.Sprintln(v...))
}

// Info обеспечивает обратную совместимость с logger.Info(msg...)
func Info(v ...interface{}) {
	Infof("%s", fmt.Sprintln(v...))
}

// Warn обеспечивает обратную совместимость с logger.Warn(msg...)
func Warn(v ...interface{}) {
	Warnf("%s", fmt.Sprintln(v...))
}

// Error обеспечивает обратную совместимость с logger.Error(msg...)
func Error(v ...interface{}) {
	Errorf("%s", fmt.Sprintln(v...))
}

// Fatal обеспечивает обратную совместимость с logger.Fatal(msg...)
func Fatal(v ...interface{}) {
	Fatalf("%s", fmt.Sprintln(v...))
}
