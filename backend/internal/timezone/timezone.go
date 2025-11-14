package timezone

import "time"

// init sets the process-local timezone early during program initialization.
// Import this package with a blank import in main so the init runs before
// other packages that may log during startup.
func init() {
	if loc, err := time.LoadLocation("Asia/Kolkata"); err == nil {
		time.Local = loc
	}
}
