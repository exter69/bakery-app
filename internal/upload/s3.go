package upload

import "fmt"

// ErrS3NotImplemented is returned when the S3 storage backend is selected but not yet implemented.
var ErrS3NotImplemented = fmt.Errorf("S3 storage backend is not yet implemented; use UPLOAD_STORAGE=local or leave unset")
