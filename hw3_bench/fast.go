package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

// вам надо написать более быструю оптимальную функцию
func FastSearch(out io.Writer) {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	r := regexp.MustCompile("@")
	androidPattern := regexp.MustCompile("Android")
	msiePattern := regexp.MustCompile("MSIE")
	seenBrowsers := make(map[string]bool)

	users := make([]User, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		user := &User{}
		err := user.UnmarshalJSON([]byte(scanner.Text()))
		if err != nil {
			panic(err)
		}
		users = append(users, *user)
	}

	var isAndroid bool
	var isMSIE bool
	_, _ = fmt.Fprintln(out, "found users:")
	for i, user := range users {

		isAndroid = false
		isMSIE = false

		browsers := user.Browsers

		for _, browserRaw := range browsers {
			if ok := androidPattern.MatchString(browserRaw); ok {
				isAndroid = true
				seenBrowsers[browserRaw] = true
			}

			if ok := msiePattern.MatchString(browserRaw); ok {
				isMSIE = true
				seenBrowsers[browserRaw] = true
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := r.ReplaceAllString(user.Email, " [at] ")
		foundUser := fmt.Sprintf("[%d] %s <%s>", i, user.Name, email)
		_, _ = fmt.Fprintln(out, foundUser)
	}

	_, _ = fmt.Fprintln(out, "\nTotal unique browsers", len(seenBrowsers))
}
