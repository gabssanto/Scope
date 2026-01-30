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
	defer func() { _ = tx.Rollback() }()

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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = rows.Close() }()

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
	defer func() { _ = rows.Close() }()

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

// ListAllFolders returns all unique folders that have at least one tag
func ListAllFolders() ([]string, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := database.Query(`
		SELECT DISTINCT f.path
		FROM folders f
		JOIN folder_tags ft ON f.id = ft.folder_id
		ORDER BY f.path
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	defer func() { _ = rows.Close() }()

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

// RenameTag renames a tag across all folders
func RenameTag(oldName, newName string) error {
	database := db.GetDB()
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	// Check if old tag exists
	var oldID int64
	err := database.QueryRow("SELECT id FROM tags WHERE name = ?", oldName).Scan(&oldID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("tag not found: %s", oldName)
	}
	if err != nil {
		return fmt.Errorf("failed to query tag: %w", err)
	}

	// Check if new tag name already exists
	var existingID int64
	err = database.QueryRow("SELECT id FROM tags WHERE name = ?", newName).Scan(&existingID)
	if err == nil {
		return fmt.Errorf("tag already exists: %s", newName)
	}
	if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing tag: %w", err)
	}

	// Rename the tag
	_, err = database.Exec("UPDATE tags SET name = ? WHERE id = ?", newName, oldID)
	if err != nil {
		return fmt.Errorf("failed to rename tag: %w", err)
	}

	return nil
}

// MergeTag merges source tag into destination tag (moves all folders, deletes source)
func MergeTag(srcName, dstName string) (int, error) {
	database := db.GetDB()
	if database == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	// Check if source tag exists
	var srcID int64
	err := database.QueryRow("SELECT id FROM tags WHERE name = ?", srcName).Scan(&srcID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("source tag not found: %s", srcName)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to query source tag: %w", err)
	}

	// Check if destination tag exists, create if not
	var dstID int64
	err = database.QueryRow("SELECT id FROM tags WHERE name = ?", dstName).Scan(&dstID)
	if err == sql.ErrNoRows {
		// Create destination tag
		result, err := database.Exec("INSERT INTO tags (name, created_at) VALUES (?, ?)", dstName, time.Now().Unix())
		if err != nil {
			return 0, fmt.Errorf("failed to create destination tag: %w", err)
		}
		dstID, _ = result.LastInsertId()
	} else if err != nil {
		return 0, fmt.Errorf("failed to query destination tag: %w", err)
	}

	// Get folders from source tag
	folders, err := ListFoldersByTag(srcName)
	if err != nil {
		return 0, fmt.Errorf("failed to get folders for source tag: %w", err)
	}

	// Move folder associations to destination tag
	movedCount := 0
	for _, folder := range folders {
		// Get folder ID
		var folderID int64
		err := database.QueryRow("SELECT id FROM folders WHERE path = ?", folder).Scan(&folderID)
		if err != nil {
			continue
		}

		// Insert into destination (ignore if already exists)
		_, err = database.Exec("INSERT OR IGNORE INTO folder_tags (folder_id, tag_id, created_at) VALUES (?, ?, ?)",
			folderID, dstID, time.Now().Unix())
		if err == nil {
			movedCount++
		}
	}

	// Delete source tag (cascade deletes folder_tags)
	_, err = database.Exec("DELETE FROM tags WHERE id = ?", srcID)
	if err != nil {
		return movedCount, fmt.Errorf("failed to delete source tag: %w", err)
	}

	return movedCount, nil
}

// CloneTag copies all folder associations from source tag to a new tag
func CloneTag(srcName, newName string) (int, error) {
	database := db.GetDB()
	if database == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	// Check if source tag exists
	var srcID int64
	err := database.QueryRow("SELECT id FROM tags WHERE name = ?", srcName).Scan(&srcID)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("source tag not found: %s", srcName)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to query source tag: %w", err)
	}

	// Check if new tag already exists
	var existingID int64
	err = database.QueryRow("SELECT id FROM tags WHERE name = ?", newName).Scan(&existingID)
	if err == nil {
		return 0, fmt.Errorf("tag already exists: %s", newName)
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to check existing tag: %w", err)
	}

	// Create new tag
	result, err := database.Exec("INSERT INTO tags (name, created_at) VALUES (?, ?)", newName, time.Now().Unix())
	if err != nil {
		return 0, fmt.Errorf("failed to create new tag: %w", err)
	}
	newID, _ := result.LastInsertId()

	// Copy folder associations
	res, err := database.Exec(`
		INSERT INTO folder_tags (folder_id, tag_id, created_at)
		SELECT folder_id, ?, ? FROM folder_tags WHERE tag_id = ?
	`, newID, time.Now().Unix(), srcID)
	if err != nil {
		return 0, fmt.Errorf("failed to copy folder associations: %w", err)
	}

	count, _ := res.RowsAffected()
	return int(count), nil
}

// DoctorResult holds the results of a health check
type DoctorResult struct {
	TotalTags         int
	TotalFolders      int
	TotalAssociations int
	OrphanedTags      []string // Tags with no folders
	MissingFolders    []string // Folders that don't exist on disk
	DuplicateFolders  []string // Same path registered multiple times
}

// Doctor performs health checks on the database
func Doctor() (*DoctorResult, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	result := &DoctorResult{}

	// Count tags
	err := database.QueryRow("SELECT COUNT(*) FROM tags").Scan(&result.TotalTags)
	if err != nil {
		return nil, fmt.Errorf("failed to count tags: %w", err)
	}

	// Count folders
	err = database.QueryRow("SELECT COUNT(*) FROM folders").Scan(&result.TotalFolders)
	if err != nil {
		return nil, fmt.Errorf("failed to count folders: %w", err)
	}

	// Count associations
	err = database.QueryRow("SELECT COUNT(*) FROM folder_tags").Scan(&result.TotalAssociations)
	if err != nil {
		return nil, fmt.Errorf("failed to count associations: %w", err)
	}

	// Find orphaned tags (tags with no folders)
	rows, err := database.Query(`
		SELECT t.name FROM tags t
		LEFT JOIN folder_tags ft ON t.id = ft.tag_id
		WHERE ft.tag_id IS NULL
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query orphaned tags: %w", err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			result.OrphanedTags = append(result.OrphanedTags, name)
		}
	}
	_ = rows.Close()

	// Find missing folders (folders that don't exist on disk)
	rows, err = database.Query("SELECT path FROM folders")
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			if _, err := os.Stat(path); os.IsNotExist(err) {
				result.MissingFolders = append(result.MissingFolders, path)
			}
		}
	}
	_ = rows.Close()

	return result, nil
}

// PruneResult holds the result of a prune operation
type PruneResult struct {
	RemovedFolders []string
	RemovedCount   int
}

// Prune removes folders that no longer exist from the database
func Prune(dryRun bool) (*PruneResult, error) {
	database := db.GetDB()
	if database == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Get all folders
	rows, err := database.Query("SELECT id, path FROM folders")
	if err != nil {
		return nil, fmt.Errorf("failed to query folders: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var toRemove []struct {
		id   int64
		path string
	}

	for rows.Next() {
		var id int64
		var path string
		if err := rows.Scan(&id, &path); err != nil {
			return nil, fmt.Errorf("failed to scan folder: %w", err)
		}

		// Check if folder exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			toRemove = append(toRemove, struct {
				id   int64
				path string
			}{id, path})
		}
	}

	result := &PruneResult{
		RemovedFolders: make([]string, 0, len(toRemove)),
	}

	if dryRun {
		for _, f := range toRemove {
			result.RemovedFolders = append(result.RemovedFolders, f.path)
		}
		result.RemovedCount = len(toRemove)
		return result, nil
	}

	// Remove non-existent folders
	for _, f := range toRemove {
		_, err := database.Exec("DELETE FROM folders WHERE id = ?", f.id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete folder %s: %w", f.path, err)
		}
		result.RemovedFolders = append(result.RemovedFolders, f.path)
	}
	result.RemovedCount = len(toRemove)

	return result, nil
}
