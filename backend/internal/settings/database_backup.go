package settings

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// DownloadCloudBackup exports a transactionally consistent, gzipped JSON
// snapshot of the cloud database. The route is protected by OwnerOnly.
func (h *DatabaseHandler) DownloadCloudBackup(c *gin.Context) {
	if h.db == nil || (!strings.EqualFold(h.db.DriverName(), "pgx") && !strings.EqualFold(h.db.DriverName(), "postgres")) {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "cloud database backup is available only on the PostgreSQL service", "code": "CLOUD_DATABASE_REQUIRED"})
		return
	}

	tx, err := h.db.BeginTxx(c.Request.Context(), &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start database snapshot"})
		return
	}
	defer tx.Rollback()

	var schema string
	if err := tx.GetContext(c.Request.Context(), &schema, `SELECT current_schema()`); err != nil || schema == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to identify database schema"})
		return
	}
	var tables []string
	if err := tx.SelectContext(c.Request.Context(), &tables, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = current_schema() AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list database tables"})
		return
	}

	file, err := os.CreateTemp("", "partflow-cloud-backup-*.json.gz")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare backup file"})
		return
	}
	fileName := file.Name()
	defer os.Remove(fileName)
	defer file.Close()
	if err := file.Chmod(0600); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to secure backup file"})
		return
	}

	compressed := gzip.NewWriter(file)
	write := func(value string) error {
		_, err := io.WriteString(compressed, value)
		return err
	}
	generatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	metadata, _ := json.Marshal(map[string]any{
		"format_version": 1,
		"generated_at":   generatedAt,
		"schema":         schema,
	})
	if _, err := compressed.Write([]byte(`{"metadata":`)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write backup"})
		return
	}
	if _, err := compressed.Write(metadata); err != nil || write(`,"tables":{`) != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write backup"})
		return
	}

	for index, table := range tables {
		if index > 0 {
			if err := write(","); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write backup"})
				return
			}
		}
		encodedName, _ := json.Marshal(table)
		if _, err := compressed.Write(encodedName); err != nil || write(`:[`) != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write backup"})
			return
		}
		rows, err := tx.QueryContext(c.Request.Context(), `SELECT row_to_json(t)::text FROM `+quoteSQLIdentifier(schema)+`.`+quoteSQLIdentifier(table)+` AS t`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export database table"})
			return
		}
		first := true
		for rows.Next() {
			var row string
			if err := rows.Scan(&row); err != nil || !json.Valid([]byte(row)) {
				_ = rows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read database row"})
				return
			}
			if !first {
				if err := write(","); err != nil {
					_ = rows.Close()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write backup"})
					return
				}
			}
			if err := write(row); err != nil {
				_ = rows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write backup"})
				return
			}
			first = false
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read database table"})
			return
		}
		if err := rows.Close(); err != nil || write("]") != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finish database table"})
			return
		}
	}
	if err := write("}}"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to finish backup"})
		return
	}
	if err := compressed.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compress backup"})
		return
	}
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete database snapshot"})
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare backup download"})
		return
	}
	info, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to inspect backup file"})
		return
	}

	filename := fmt.Sprintf("partflow-cloud-backup-%s.json.gz", time.Now().UTC().Format("20060102-150405"))
	c.Header("Content-Type", "application/gzip")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Header("Content-Length", fmt.Sprintf("%d", info.Size()))
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, file)
}

func quoteSQLIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
