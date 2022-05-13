package main

import "fmt"

func indexTextFile(
	command string,
	path string,
	name string,
	part int,
	originalPath string,
	originalName string,
	content []byte,
) error {
	// index the file -- if we are appending, we should only incrementally index
	_, err := theDB.Exec(
		`INSERT INTO filesearch (cmd, path, name, part, original_path, original_name, content) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		command,
		path,
		name,
		part,
		originalPath,
		originalName,
		content,
	)
	if err != nil {
		return fmt.Errorf("ERR while indexing %s %s%s: %v", command, path, name, err)
	}
	return nil
}
