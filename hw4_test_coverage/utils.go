package main

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
)

func filterByQuery(query string, users []User) []User {
	if query != "" {
		var resultUsers = make([]User, 0, 0)
		for _, user := range users {
			contains := false
			if match, err := regexp.MatchString(query, user.Name); err == nil && match {
				contains = true
			}
			if match, err := regexp.MatchString(query, user.About); err == nil && match {
				contains = true
			}

			if contains {
				resultUsers = append(resultUsers, user)
			}
		}
		return resultUsers
	}
	return users
}

func sortUsers(orderField string, users []User, orderBy int) {
	switch orderField {
	case OrderFieldId:
		sort.Slice(users, func(i, j int) bool {
			if orderBy == OrderByDesc {
				return users[i].Id < users[j].Id
			} else {
				return users[i].Id > users[j].Id
			}
		})
	case OrderFieldName:
		fallthrough
	case "":
		sort.Slice(users, func(i, j int) bool {
			if orderBy == OrderByDesc {
				return users[i].Name < users[j].Name
			} else {
				return users[i].Name > users[j].Name
			}
		})
	case OrderFieldAge:
		sort.Slice(users, func(i, j int) bool {
			if orderBy == OrderByDesc {
				return users[i].Age < users[j].Age
			} else {
				return users[i].Age > users[j].Age
			}
		})
	}
}

func readDataset(path string) []User {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	var dataset Dataset
	err = xml.Unmarshal(byteValue, &dataset)
	if err != nil {
		panic(err)
	}

	users := make([]User, 0)
	for _, row := range dataset.Row {
		user := &User{
			Id:     row.Id,
			Name:   row.FirstName + row.LastName,
			Age:    row.Age,
			Gender: row.Gender,
			About:  row.About,
		}
		users = append(users, *user)
	}
	return users
}
