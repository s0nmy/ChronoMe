package gormrepo

import (
	"time"

	"gorm.io/gorm"
)

// updateValues はゼロ値を保持し、指定された列だけを更新値に含める。
func updateValues(fields map[string]any, columns []string, now time.Time) map[string]any {
	values := map[string]any{"version": gorm.Expr("version + 1"), "updated_at": now}
	for _, column := range columns {
		if value, ok := fields[column]; ok {
			values[column] = value
		}
	}
	return values
}
