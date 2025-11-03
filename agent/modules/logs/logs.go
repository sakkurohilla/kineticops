package logs

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sakkurohilla/kineticops/agent/config"
	"github.com/sakkurohilla/kineticops/agent/pipelines"
	"github.com/sakkurohilla/kineticops/agent/state"
	"github.com/sakkurohilla/kineticops/agent/utils"
)

// LogsModule collects log files
type LogsModule struct {
	config   *config.LogsModule
	pipeline *pipelines.PipelineManager
	state    *state.Manager
	logger   *utils.Logger
	stopChan chan struct{}
	watchers map[string]*LogWatcher
}

// LogWatcher watches a single log file
type LogWatcher struct {
	path     string
	file     *os.File
	scanner  *bufio.Scanner
	offset   int64
	watcher  *fsnotify.Watcher
	stopChan chan struct{}
}

// NewLogsModule creates a new logs module
func NewLogsModule(cfg *config.LogsModule, pipeline *pipelines.PipelineManager, stateManager *state.Manager, logger *utils.Logger) (*LogsModule, error) {
	return &LogsModule{
		config:   cfg,
		pipeline: pipeline,
		state:    stateManager,
		logger:   logger,
		stopChan: make(chan struct{}),
		watchers: make(map[string]*LogWatcher),
	}, nil
}

// Name returns the module name
func (l *LogsModule) Name() string {
	return "logs"
}

// IsEnabled returns whether the module is enabled
func (l *LogsModule) IsEnabled() bool {
	return l.config.Enabled
}

// Start begins collecting logs
func (l *LogsModule) Start(ctx context.Context) error {
	l.logger.Info("Starting log collection", "inputs", len(l.config.Inputs))

	// Start watching each input
	for _, input := range l.config.Inputs {
		if err := l.startInput(ctx, &input); err != nil {
			l.logger.Error("Failed to start input", "paths", input.Paths, "error", err)
			continue
		}
	}

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Stop stops log collection
func (l *LogsModule) Stop() error {
	close(l.stopChan)
	
	// Stop all watchers
	for path, watcher := range l.watchers {
		if err := watcher.Stop(); err != nil {
			l.logger.Error("Error stopping watcher", "path", path, "error", err)
		}
	}
	
	return nil
}

// startInput starts watching files for a log input
func (l *LogsModule) startInput(ctx context.Context, input *config.LogInput) error {
	// Expand glob patterns
	var files []string
	for _, pattern := range input.Paths {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			l.logger.Error("Invalid glob pattern", "pattern", pattern, "error", err)
			continue
		}
		files = append(files, matches...)
	}

	// Start watching each file
	for _, file := range files {
		// Check if file should be excluded
		if l.shouldExclude(file, input.Exclude) {
			continue
		}

		if err := l.startWatching(ctx, file, input); err != nil {
			l.logger.Error("Failed to start watching file", "file", file, "error", err)
			continue
		}
	}

	return nil
}

// shouldExclude checks if a file should be excluded
func (l *LogsModule) shouldExclude(file string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if matched, _ := filepath.Match(pattern, file); matched {
			return true
		}
	}
	return false
}

// startWatching starts watching a single log file
func (l *LogsModule) startWatching(ctx context.Context, filePath string, input *config.LogInput) error {
	// Check if already watching
	if _, exists := l.watchers[filePath]; exists {
		return nil
	}

	// Create file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Add file to watcher
	if err := watcher.Add(filePath); err != nil {
		watcher.Close()
		return fmt.Errorf("failed to watch file: %w", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		watcher.Close()
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Get last read position from state
	offset := l.state.GetOffset(filePath)
	if offset > 0 {
		if _, err := file.Seek(offset, 0); err != nil {
			l.logger.Warn("Failed to seek to last position", "file", filePath, "offset", offset, "error", err)
			offset = 0
		}
	}

	logWatcher := &LogWatcher{
		path:     filePath,
		file:     file,
		scanner:  bufio.NewScanner(file),
		offset:   offset,
		watcher:  watcher,
		stopChan: make(chan struct{}),
	}

	l.watchers[filePath] = logWatcher

	// Start watching in goroutine
	go l.watchFile(ctx, logWatcher, input)

	l.logger.Info("Started watching file", "file", filePath, "offset", offset)
	return nil
}

// watchFile watches a single file for changes
func (l *LogsModule) watchFile(ctx context.Context, watcher *LogWatcher, input *config.LogInput) {
	defer func() {
		watcher.file.Close()
		watcher.watcher.Close()
		delete(l.watchers, watcher.path)
	}()

	// Read existing content first
	l.readLines(watcher, input)

	for {
		select {
		case <-ctx.Done():
			return
		case <-watcher.stopChan:
			return
		case event, ok := <-watcher.watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				l.readLines(watcher, input)
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
				l.logger.Info("File removed or renamed", "file", watcher.path)
				return
			}

		case err, ok := <-watcher.watcher.Errors:
			if !ok {
				return
			}
			l.logger.Error("File watcher error", "file", watcher.path, "error", err)
		}
	}
}

// readLines reads new lines from the file
func (l *LogsModule) readLines(watcher *LogWatcher, input *config.LogInput) {
	for watcher.scanner.Scan() {
		line := watcher.scanner.Text()
		if line == "" {
			continue
		}

		// Create log event
		event := l.createLogEvent(line, watcher.path, input)
		
		// Send to pipeline
		if err := l.pipeline.Send(event); err != nil {
			l.logger.Error("Failed to send log event", "error", err)
			continue
		}

		// Update offset
		watcher.offset += int64(len(line)) + 1 // +1 for newline
		l.state.SetOffset(watcher.path, watcher.offset)
	}

	if err := watcher.scanner.Err(); err != nil {
		l.logger.Error("Scanner error", "file", watcher.path, "error", err)
	}
}

// createLogEvent creates a log event from a line
func (l *LogsModule) createLogEvent(line, filePath string, input *config.LogInput) map[string]interface{} {
	timestamp := time.Now().UTC()
	hostname, _ := os.Hostname()

	event := map[string]interface{}{
		"@timestamp": timestamp.Format(time.RFC3339),
		"agent": map[string]interface{}{
			"name":    "kineticops-agent",
			"type":    "filebeat",
			"version": "1.0.0",
		},
		"host": map[string]interface{}{
			"hostname": hostname,
		},
		"event": map[string]interface{}{
			"kind":     "event",
			"category": "file",
			"type":     "info",
		},
		"log": map[string]interface{}{
			"file": map[string]interface{}{
				"path": filePath,
			},
			"offset": 0, // Will be set by the pipeline
		},
		"message": line,
		"input": map[string]interface{}{
			"type": input.Type,
		},
	}

	// Add custom fields
	if len(input.Fields) > 0 {
		fields := make(map[string]interface{})
		for k, v := range input.Fields {
			fields[k] = v
		}
		event["fields"] = fields
	}

	// Parse log level from message
	if level := l.extractLogLevel(line); level != "" {
		event["log"].(map[string]interface{})["level"] = level
	}

	return event
}

// extractLogLevel tries to extract log level from the message
func (l *LogsModule) extractLogLevel(message string) string {
	message = strings.ToUpper(message)
	
	levels := []string{"ERROR", "WARN", "WARNING", "INFO", "DEBUG", "TRACE", "FATAL"}
	for _, level := range levels {
		if strings.Contains(message, level) {
			return strings.ToLower(level)
		}
	}
	
	return ""
}

// Stop stops a specific watcher
func (w *LogWatcher) Stop() error {
	close(w.stopChan)
	return nil
}