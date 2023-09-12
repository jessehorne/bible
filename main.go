package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/exp/slices"
	"strconv"
)

// helpMessage contains the text printed when using the --help flag
// remember to update README.md when this changes
var helpMessage = `Usage: bible [OPTION]...
Access the Holy Bible in your terminal.

  --b=...         Book
                  DEFAULT: "Gen"

  --v=...         Verse(s) (Examples: "1:10-11", "5", "3:16")
                  DEFAULT: Random verse(s)
                  
  --t=...         Version (Examples: "KJV")
                  DEFAULT: "KJV"
                  
  --l=...         Language (Examples: "EN")
                  DEFAULT: "EN"
  
  -lt             List supported versions.
  --lb		  	  List all books in a version.
  
  -n              Include the number of the verse when printed. (Example: "1 In the beginning..." vs "In the beginning...")

  --help          Show this information.

Examples:
> bible --b="Gen" --v=1:1-2
Genesis 1:1-2
In the beginning God created the heaven and the earth.
And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters.

> bible --b="Gen" --v=1:1-2 -n
Genesis 1:1-2
1:1 In the beginning God created the heaven and the earth.
1:2 And the earth was without form, and void; and darkness was upon the face of the deep. And the Spirit of God moved upon the face of the waters.

For more information, please visit https://github.com/jessehorne/bible`

var connections = map[string]*sql.DB{
	"kjv": nil,
}

type Verse struct {
	Chapter int
	Number  int
	Content string
}

func loadDatabases() {
	// connect to local database
	kjvEnglishConn, err := sql.Open("sqlite3", "data/kjv.db")
	if err != nil {
		panic(err)
	}
	connections["kjv-en"] = kjvEnglishConn
}

func versionExists(v string) bool {
	_, ok := connections[v]
	return ok
}

// bookExists returns true or false depending on if the given book exists for the version
func bookExists(v string, book string) bool {
	if !versionExists(v) {
		return false
	}

	books, err := getBooks(v)
	if err != nil {
		return false
	}

	return slices.Contains(books, book)

	return false
}

// getBooks returns an array of books supported by the version if it exists, otherwise an error
func getBooks(v string) ([]string, error) {
	var books []string
	if !versionExists(v) {
		return books, errors.New(fmt.Sprintf("'%s' isn't a supported version", v))
	}

	query := "SELECT DISTINCT book FROM bible"
	q, err := connections[v].Query(query)
	if err != nil {
		return books, err
	}
	defer q.Close()

	for q.Next() {
		var b string
		q.Scan(&b)
		books = append(books, b)
	}

	return books, nil
}

// getVerses returns an array of verses for the specified version+book, if the book and/or verses exist, otherwise an error
func getVerses(v string, book string, chapter int, start int, end int) ([]Verse, error) {
	var verses []Verse

	var query string
	if chapter == 0 {
		query = fmt.Sprintf("SELECT chapter, verse as number, content FROM bible WHERE book='%s'", book)
	} else if start == 0 || end == 0 {
		query = fmt.Sprintf("SELECT chapter, verse as number, content FROM bible WHERE book='%s' AND chapter='%d'", book, chapter)
	} else {
		query = fmt.Sprintf("SELECT chapter, verse as number, content FROM bible WHERE book='%s' AND chapter='%d' AND verse BETWEEN %d AND %d", book, chapter, start, end)
	}
	q, err := connections[v].Query(query)
	if err != nil {
		return verses, err
	}
	defer q.Close()

	for q.Next() {
		var v Verse
		q.Scan(&v.Chapter, &v.Number, &v.Content)
		verses = append(verses, v)
	}

	return verses, nil
}

// stripVerse strips verse of the <verse> tags included in the ebible db
func stripVerse(v string) string {
	var started bool
	var final string
	var charFoundYet bool

	for i := 0; i < len(v); i++ {
		c := string(v[i])

		if c == "\n" {
			continue
		}

		if c == "<" {
			started = true
		} else if c == ">" {
			started = false
		}

		// strip of <, >, and first space
		if !started && c != "<" && c != ">" {
			if c == " " && !charFoundYet {
				continue
			}

			if c != "" {
				charFoundYet = true
			}

			final += c
		}
	}

	return final
}

// versesToInts returns the chapter, start verse and end verse from a X:Y-Z formatted string
func versesToInts(v string) (int, int, int, error) {
	var chapter string
	var start string
	var end string

	var seenColonYet bool
	var seenDashYet bool

	stage := "chapter"

	// loop through the verses string
	for i := 0; i < len(v); i++ {
		c := string(v[i])

		if c == ":" {
			if seenColonYet {
				return 0, 0, 0, errors.New("invalid: you only need one ':'")
			} else {
				if stage != "chapter" {
					return 0, 0, 0, errors.New("invalid: can't get to start before finding chapter")
				}

				seenColonYet = true
				stage = "start"
			}
		} else if c == "-" {
			if seenDashYet {
				return 0, 0, 0, errors.New("invalid: you only need one '-'")
			} else {
				if stage != "start" {
					return 0, 0, 0, errors.New("invalid: can't get to end before finding start")
				}

				seenDashYet = true
				stage = "end"
			}
		} else {
			if stage == "chapter" {
				chapter += c
			} else if stage == "start" {
				start += c
			} else if stage == "end" {
				end += c
			}
		}
	}

	var chapterInt int
	var startInt int
	var endInt int

	chapterInt, err := strconv.Atoi(chapter)
	if err != nil {
		return 0, 0, 0, errors.New("invalid: chapter is formatted incorrectly")
	}

	if len(start) == 0 {
		startInt = 1
	} else {
		tempStartInt, err := strconv.Atoi(start)
		if err != nil {
			return 0, 0, 0, errors.New("invalid: start verse is formatted incorrectly")
		}
		startInt = tempStartInt
	}

	if len(end) == 0 {
		endInt = 0
	} else {
		tempEndInt, err := strconv.Atoi(end)
		if err != nil {
			return 0, 0, 0, errors.New("invalid: end verse is formatted incorrectly")
		}
		endInt = tempEndInt
	}

	return chapterInt, startInt, endInt, nil
}

func main() {
	loadDatabases()

	var book = flag.String("b", "Gen", "Book")
	var verses = flag.String("v", "0", "Verses")
	var version = flag.String("t", "kjv-en", "Version")

	var listVersions = flag.Bool("lt", false, "List all versions")
	var listBooks = flag.Bool("lb", false, "List all books in the chosen version.")

	var showNumbers = flag.Bool("n", false, "Show verse numbers.")

	flag.Usage = func() {
		fmt.Println(helpMessage)
	}

	flag.Parse()

	if *listBooks {
		books, err := getBooks(*version)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		for i, b := range books {
			fmt.Print(b)

			if i < len(books)-1 {
				fmt.Print(", ")
			} else {
				fmt.Println()
			}
		}

		return
	} else if *listVersions {
		// this will hopefully not be hardcoded soon <3
		fmt.Println("KJV in English (kjv-en)")
		return
	} else {
		// Nothing else to do besides get verses...

		// check that version exists
		if !versionExists(*version) {
			fmt.Println("invalid: that version doesn't exist...try 'kjv-en' or '--help")
			return
		}

		// check that book exists
		if !bookExists(*version, *book) {
			fmt.Println("invalid: book doesn't exist for that version")
			return
		}

		// get verses from verse string
		chapter, start, end, err := versesToInts(*verses)
		if err != nil {
			fmt.Println(err.Error())
		}

		vs, err := getVerses(*version, *book, chapter, start, end)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if len(vs) == 0 {
			fmt.Println("invalid: unknown verse range")
			return
		}

		for _, v := range vs {
			if *showNumbers {
				fmt.Print(v.Chapter, ":", v.Number, " ")
			}

			fmt.Print(stripVerse(v.Content) + "\n")
		}

		return
	}

	fmt.Println("Try 'bible --help'.")
}
