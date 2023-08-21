package linedb_test

import (
	"os"
	"strings"
	"testing"

	. "cdop.pt/go/open/assertive"
	"cdop.pt/go/open/linedb"
)

func TestLoadFailure(t *testing.T) {
	filePath := "/tmp/no-such-file-or-directory"

	db, err := linedb.Open(filePath)
	Want(t, err != nil)
	Want(t, db == nil)
}

func TestRecord(t *testing.T) {
	tests := []struct {
		name   string
		dbCont string
		number int
		text   string
		err    string
	}{
		{"empty", "", 1, "", "no records in database"},
		{"oneline[2]", "ln1", 2, "", "record number (2) out of bounds [1, 1]"},
		{"twolines[1]", "ln1\nln2", 1, "ln1", ""},
		{"twolines[2]", "ln1\nln2", 2, "ln2", ""},
		{"twolines[3]", "ln1\nln2", 3, "", "record number (3) out of bounds [1, 2]"},
	}

	for _, x := range tests {
		t.Run(x.name, func(t *testing.T) {

			filePath := mkTestFile(t, x.dbCont)
			defer os.Remove(filePath)

			db, err := linedb.Open(filePath)
			Need(t, err == nil)

			text, err := db.Record(x.number)

			if x.err == "" {
				Want(t, err == nil)
			} else {
				Need(t, err != nil)
				Want(t, err.Error() == x.err)
			}
			Want(t, text == x.text)

		})
	}
}

func TestSelect(t *testing.T) {
	tests := []struct {
		name    string
		dbCont  string
		pattern string
		records []linedb.Rec
	}{
		{"empty", "", "pattern", []linedb.Rec{}},
		{"test1", "ln1", "pattern", []linedb.Rec{}},
		{"test2", "ln1", "ln1", []linedb.Rec{
			{Number: 1, Text: "ln1"},
		}},
		{"test3", "ln1\nln2\nln3", "ln2", []linedb.Rec{
			{Number: 2, Text: "ln2"},
		}},
		{"test4", "ln1\nln2\nln3", "ln7", []linedb.Rec{}},
		{"test5", "ln1\nline21\nln3\nline22\nln5", "ine", []linedb.Rec{
			{Number: 2, Text: "line21"},
			{Number: 4, Text: "line22"},
		}},
	}

	for _, x := range tests {

		pattern := x.pattern

		t.Run(x.name, func(t *testing.T) {

			filePath := mkTestFile(t, x.dbCont)
			defer os.Remove(filePath)

			db, err := linedb.Open(filePath)
			Need(t, err == nil)

			records := db.Select(func(r linedb.Rec) bool {
				return strings.Contains(r.Text, pattern)
			})

			Want(t, sliceEq(records, x.records))

		})
	}
}

func TestInsert(t *testing.T) {
	tests := []struct {
		name    string
		dbContS string
		number  int
		text    string
		err     string
		length  int
		dbContE string
	}{
		{"empty2empty", "", 0, "", "", 0, ""},
		{"outofbounds1", "", -1, "", "record number (-1) out of bounds [0, 0]", 0, ""},
		{"outofbounds2", "ln1", 2, "", "record number (2) out of bounds [0, 1]", 1, "ln1"},
		{"insertempty", "", 0, "text", "", 1, "text"},
		{"insert1empty0", " ", 0, "text", "", 2, " \ntext"},
		{"insert1empty1", " ", 1, "text", "", 2, "text\n "},
		{"insert1", "ln1\nln2\nln3", 1, "text", "", 4, "text\nln1\nln2\nln3"},
		{"insert2", "ln1\nln2\nln3", 2, "text", "", 4, "ln1\ntext\nln2\nln3"},
		{"insert3", "ln1\nln2\nln3", 3, "text", "", 4, "ln1\nln2\ntext\nln3"},
		{"insert4", "ln1\nln2\nln3", 0, "text", "", 4, "ln1\nln2\nln3\ntext"},
	}

	for _, x := range tests {
		t.Run(x.name, func(t *testing.T) {

			filePath := mkTestFile(t, x.dbContS)
			defer os.Remove(filePath)

			db, err := linedb.Open(filePath)
			Need(t, err == nil)

			err = db.Insert(x.number, x.text)

			if x.err == "" {
				Want(t, err == nil)
			} else {
				Need(t, err != nil)
				Want(t, err.Error() == x.err)
			}

			buf, err := os.ReadFile(filePath)
			Need(t, err == nil)

			Want(t, db.Length() == x.length)
			Want(t, string(buf) == x.dbContE)

		})
	}
}

func TestInsertFail(t *testing.T) {
	tests := []struct {
		name    string
		dbContS string
		number  int
		recs    []linedb.Rec
	}{
		{"empty", "", 1, []linedb.Rec{}},
		{"end", "ln1", 0, []linedb.Rec{{Number: 1, Text: "ln1"}}},
		{"general", "ln1", 1, []linedb.Rec{{Number: 1, Text: "ln1"}}},
	}

	for _, x := range tests {
		dbContS := x.dbContS
		number := x.number
		recs := x.recs

		t.Run(x.name, func(t *testing.T) {
			filePath := mkTestFile(t, dbContS)
			defer os.Remove(filePath)

			db, err := linedb.Open(filePath)
			Need(t, err == nil)
			Want(t, db != nil)

			err = os.Chmod(filePath, 0400)
			Need(t, err == nil)
			defer os.Chmod(filePath, 0600)

			err = db.Insert(number, "text")
			Need(t, err != nil)

			Want(t, sliceEq(db.All(), recs))
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name    string
		dbContS string
		number  int
		rec     string
		old     string
		err     string
		dbContE string
	}{
		{"empty2empty", "", 1, "", "", "no records in database", ""},
		{"empty2one", "", 1, "ln1", "", "no records in database", ""},
		{"outofbounds1", "ln1", 0, "", "", "record number (0) out of bounds [1, 1]", "ln1"},
		{"outofbounds2", "ln1", 2, "", "", "record number (2) out of bounds [1, 1]", "ln1"},
		{"update1", "ln1\nln2\nln3", 1, "new", "ln1", "", "new\nln2\nln3"},
		{"update2", "ln1\nln2\nln3", 2, "new", "ln2", "", "ln1\nnew\nln3"},
		{"update3", "ln1\nln2\nln3", 3, "new", "ln3", "", "ln1\nln2\nnew"},
	}

	for _, x := range tests {
		t.Run(x.name, func(t *testing.T) {

			filePath := mkTestFile(t, x.dbContS)
			defer os.Remove(filePath)

			db, err := linedb.Open(filePath)
			Need(t, err == nil)

			old, err := db.Update(x.number, x.rec)

			if x.err == "" {
				Want(t, err == nil)
			} else {
				Need(t, err != nil)
				Want(t, err.Error() == x.err)
			}
			Want(t, old == x.old)

			buf, err := os.ReadFile(filePath)
			Need(t, err == nil)

			Want(t, string(buf) == x.dbContE)

		})
	}
}

func TestUpdateFail(t *testing.T) {
	filePath := mkTestFile(t, "initial")
	defer os.Remove(filePath)

	db, err := linedb.Open(filePath)
	Need(t, err == nil)
	Want(t, db != nil)

	err = os.Chmod(filePath, 0400)
	Need(t, err == nil)
	defer os.Chmod(filePath, 0600)

	_, err = db.Update(1, "new")
	Need(t, err != nil)

	buf, err := os.ReadFile(filePath)
	Need(t, err == nil)

	Want(t, string(buf) == "initial")
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name    string
		dbContS string
		number  int
		text    string
		err     string
		length  int
		dbContE string
	}{
		{"empty2empty", "", 1, "", "no records in database", 0, ""},
		{"one2empty", "ln1", 1, "ln1", "", 0, ""},
		{"outofbounds1", "ln1", 0, "", "record number (0) out of bounds [1, 1]", 1, "ln1"},
		{"outofbounds2", "ln1", 2, "", "record number (2) out of bounds [1, 1]", 1, "ln1"},
		{"delete1", "ln1\nln2\nln3", 1, "ln1", "", 2, "ln2\nln3"},
		{"delete2", "ln1\nln2\nln3", 2, "ln2", "", 2, "ln1\nln3"},
		{"delete3", "ln1\nln2\nln3", 3, "ln3", "", 2, "ln1\nln2"},
	}

	for _, x := range tests {
		t.Run(x.name, func(t *testing.T) {

			filePath := mkTestFile(t, x.dbContS)
			defer os.Remove(filePath)

			db, err := linedb.Open(filePath)
			Need(t, err == nil)

			text, err := db.Delete(x.number)

			if x.err == "" {
				Want(t, err == nil)
			} else {
				Need(t, err != nil)
				Want(t, err.Error() == x.err)
			}
			Want(t, text == x.text)

			buf, err := os.ReadFile(filePath)
			Need(t, err == nil)

			Want(t, db.Length() == x.length)
			Want(t, string(buf) == x.dbContE)

		})
	}
}

func TestDeleteFail(t *testing.T) {
	filePath := mkTestFile(t, "initial")
	defer os.Remove(filePath)

	db, err := linedb.Open(filePath)
	Need(t, err == nil)
	Want(t, db != nil)

	err = os.Chmod(filePath, 0400)
	Need(t, err == nil)
	defer os.Chmod(filePath, 0600)

	_, err = db.Delete(1)
	Need(t, err != nil)

	buf, err := os.ReadFile(filePath)
	Need(t, err == nil)

	Want(t, string(buf) == "initial")
}

func mkTestFile(t *testing.T, content string) string {
	f, err := os.CreateTemp("", "linedb")
	Need(t, err == nil)

	name := f.Name()

	_, err = f.WriteString(content)
	Need(t, err == nil)

	err = f.Sync()
	Need(t, err == nil)

	err = f.Close()
	Need(t, err == nil)

	return name
}

func sliceEq[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
