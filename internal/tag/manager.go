package tag

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/gabssanto/Scope/internal/db"
)

// AddTag adds a tag to a folder
func AddTag(path, tagName string) error {
	// Validate folder exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("folder does not exist: %s", path)
	}

	database := db.GetDB()
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().Unix()

	// Insert or get folder
	var folderID int64
	err = tx.QueryRow("SELECT id FROM folders WHERE path = ?", path).Scan(&folderID)
	if err == sql.ErrNoRows {
		result, err := tx.Exec("INSERT INTO folders (path, created_at) VALUES (?, ?)", path, now)
		if err != nil {
			return fmt.Errorf("failed to insert folder: %w", err)
		}
		folderID, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get folder ID: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query folder: %w", err)
	}

	// Insert or get tag
	var tagID int64
	err = tx.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if err == sql.ErrNoRows {
		result, err := tx.Exec("INSERT INTO tags (name, created_at) VALUES (?, ?)", tagName, now)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}
		tagID, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get tag ID: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to query tag: %w", err)
	}

	// Insert folder_tag relationship (ignore if already exists)
	_, err = tx.Exec("INSERT OR IGNORE INTO folder_tags (folder_id, tag_id, created_at) VALUES (?, ?, ?)",
		folderID, tagID, now)
	if err != nil {
		return fmt.Errorf("failed to insert folder_tag: %w", err)
	}

	return tx.Commit()
}

// RemoveTag removes a specific tag from a folder
func RemoveTag(path, tagName string) error {
	database := db.GetDB()
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	result, err := database.Exec(`
		DELETE FROM folder_tags
		WHERE folder_id = (SELECT id FROM folders WHERE path = ?)
		AND tag_id = (SELECT id FROM tags WHERE name = ?)
	`, path, tagName)
	if err != nil {
		return fmt.Errorf("failed to remove tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("tag '%s' not found on folder: %s", tagName, path)
	}

	return nil
}

// DeleteTag deletes a tag entirely (removes from all folders)
func DeleteTag(tagName string) error {
	database := db.GetDB()
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	result, err := database.Exec("DELETE FROM tags WHERE name = ?", tagName)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("tag not found: %s", tagName)
	}

	return nil
}

// ListTags returns all tags with their folder counts
func ListTags() (map[string]int, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := database.Query(`
		SELECT t.name, COUNT(ft.folder_id) as count
		FROM tags t
		LEFT JOIN folder_tags ft ON t.id = ft.tag_id
		GROUP BY t.id, t.name
		ORDER BY t.name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	tags := make(map[string]int)
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags[name] = count
	}

	return tags, nil
}

// ListFoldersByTag returns all folders with a specific tag
func ListFoldersByTag(tagName string) ([]string, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := database.Query(`
		SELECT f.path
		FROM folders f
		JOIN folder_tags ft ON f.id = ft.folder_id
		JOIN tags t ON ft.tag_id = t.id
		WHERE t.name = ?
		ORDER BY f.path
	`, tagName)
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	defer rows.Close()

	var folders []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, path)
	}

	return folders, nil
}

// GetTagsForFolder returns all tags for a specific folder
func GetTagsForFolder(path string) ([]string, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := database.Query(`
		SELECT t.name
		FROM tags t
		JOIN folder_tags ft ON t.id = ft.tag_id
		JOIN folders f ON ft.folder_id = f.id
		WHERE f.path = ?
		ORDER BY t.name
	`, path)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, name)
	}

	return tags, nil
}
