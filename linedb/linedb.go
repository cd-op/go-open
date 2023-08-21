// Package linedb provides an OO-like interface to operate on simple
// line-record databases backed by a plain text file.
//
// The rules are as follows:
//   - One file, one database;
//   - One line, one record;
//   - An empty (0 bytes) database file contains exactly zero records;
//   - In all other cases, an empty line is a valid record;
//   - No assumptions are made about the inner structure of each record, even
//     in terms of white space characters;
//   - The only special character is the record separator, '\n'.
package linedb

import (
	"fmt"
	"os"
	"strings"
)

// Database represents a handle to a given backing file.
type Database struct {
	filePath string
	records  []string
}

// Rec pairs the text of a record with its line number in the backing file.
type Rec struct {
	Number int
	Text   string
}

// Open creates a handle to the given backing file.
func Open(filePath string) (*Database, error) {
	records, err := loadFromFile(filePath)

	if err != nil {
		return nil, err
	}

	return &Database{filePath, records}, nil
}

// Record returns the number-th record in the database. Like lines in files,
// records are 1-indexed.
func (db *Database) Record(number int) (string, error) {
	err := db.checkEmpty()
	if err != nil {
		return "", err
	}

	err = db.checkBounds(1, number)
	if err != nil {
		return "", err
	}

	return db.records[number-1], nil
}

// Length returns the number of records in the database.
func (db *Database) Length() int {
	return len(db.records)
}

// All returns all the records with their respective number
func (db *Database) All() []Rec {
	records := make([]Rec, db.Length())

	for i := range db.records {
		records[i] = Rec{i + 1, db.records[i]}
	}

	return records
}

// Select returns the records which cause the filter function to return true.
func (db *Database) Select(filter func(Rec) bool) []Rec {
	res := []Rec{}

	for i := range db.records {
		if filter(Rec{
			i + 1,
			db.records[i],
		}) {
			res = append(res, Rec{i + 1, db.records[i]})
		}
	}

	return res
}

// Insert creates a new record at position n, pushing the n-th record and
// all subsequent records one position forward, i.e, Insert(1, ...) would
// create a new record at the beginning of the file.
//
// Special case: Insert(0, ...) places the new record at the end of the file.
func (db *Database) Insert(number int, record string) error {
	err := db.checkBounds(0, number)
	if err != nil {
		return err
	}

	// special case: inserting empty record on an empty database is a noop
	if db.Length() == 0 && record == "" {
		return nil
	}

	// special case: end of file
	if number == 0 {
		newRecords := append(db.records, record)

		err := saveToFile(db.filePath, newRecords)
		if err != nil {
			return err
		}

		db.records = newRecords

		return nil
	}

	// general case:
	newRecords := make([]string, 0, db.Length()+1)
	newRecords = append(newRecords, db.records[0:number-1]...)
	newRecords = append(newRecords, record)
	newRecords = append(newRecords, db.records[number-1:]...)

	err = saveToFile(db.filePath, newRecords)
	if err != nil {
		return err
	}

	db.records = newRecords

	return nil
}

// Update replaces the text of the number-th record, and returns the old text.
func (db *Database) Update(number int, record string) (string, error) {
	old, err := db.Record(number)
	if err != nil {
		return "", err
	}

	db.records[number-1] = record

	err = saveToFile(db.filePath, db.records)
	if err != nil {
		db.records[number-1] = old
		return "", err
	}

	return old, nil

}

// Delete removes the number-th record from the database.
func (db *Database) Delete(number int) (string, error) {
	old, err := db.Record(number)
	if err != nil {
		return "", err
	}

	newRecords := make([]string, 0, db.Length()-1)
	for i := range db.records {
		if i == number-1 {
			continue
		}

		newRecords = append(newRecords, db.records[i])
	}

	err = saveToFile(db.filePath, newRecords)
	if err != nil {
		return "", err
	}

	db.records = newRecords

	return old, nil
}

func (db *Database) checkEmpty() error {
	if db.Length() == 0 {
		return fmt.Errorf("no records in database")
	}

	return nil
}

func (db *Database) checkBounds(lower, number int) error {
	numberOfRecords := db.Length()

	if number < lower || number > numberOfRecords {
		return fmt.Errorf("record number (%d) out of bounds [%d, %d]",
			number, lower, numberOfRecords)
	}

	return nil
}

func loadFromFile(filePath string) ([]string, error) {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open linedb backing file %s: %w", filePath, err)
	}

	if len(buf) == 0 {
		return []string{}, nil
	}

	records := strings.Split(string(buf), "\n")

	return records, nil
}

func saveToFile(filePath string, records []string) error {
	err := os.WriteFile(filePath, []byte(strings.Join(records, "\n")), 0)
	if err != nil {
		return fmt.Errorf("cannot save linedb backing file %s: %w", filePath, err)
	}

	return nil
}
