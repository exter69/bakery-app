//go:build tools

package tools

// This file ensures dependencies are tracked in go.mod even before
// they are used in production code. They will be imported in later tasks.

import (
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/stretchr/testify/assert"
	_ "pgregory.net/rapid"
)
