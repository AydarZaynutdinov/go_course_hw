package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"testing"
	"time"
)

// код писать тут
func SearchServer(writer http.ResponseWriter, request *http.Request) {
	// Check for token
	accessToken := request.Header.Get("AccessToken")
	if accessToken == TimeoutToken {
		time.Sleep(time.Second * 2)
		return
	}
	if accessToken == UnknownErrorToken {
		panic("unknown error")
	}
	if accessToken == UnauthorizedToken {
		writer.WriteHeader(http.StatusUnauthorized)
		return
	}
	if accessToken == InternalServerErrorToken {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if accessToken == BadRequestToken {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	if accessToken == UnknownBadRequestErrorToken {
		returnedError := SearchErrorResponse{
			Error: "someError",
		}
		writer.WriteHeader(http.StatusBadRequest)
		byteError, err := json.Marshal(returnedError)
		if err != nil {
			panic(err)
		}
		writer.Write(byteError)
		return
	}
	if accessToken == DifferentStructToken {
		user := &User{}
		byteUser, err := json.Marshal(user)
		if err != nil {
			panic(err)
		}
		writer.Write(byteUser)
		return
	}

	// Read all users
	users := readDataset(path)

	// Get params
	limit, err := strconv.Atoi(request.FormValue(RequestLimit))
	if err != nil {
		panic(err)
	}
	offset, err := strconv.Atoi(request.FormValue(RequestOffset))
	if err != nil {
		panic(err)
	}
	query := request.FormValue(RequestQuery)
	orderField := request.FormValue(RequestOrderField)
	orderBy, err := strconv.Atoi(request.FormValue(RequestOrderBy))
	if err != nil {
		panic(err)
	}

	// Check for correct order_field value
	if orderField != OrderFieldId && orderField != OrderFieldName && orderField != OrderFieldAge {
		returnedError := SearchErrorResponse{
			Error: "ErrorBadOrderField",
		}
		writer.WriteHeader(http.StatusBadRequest)
		byteError, err := json.Marshal(returnedError)
		if err != nil {
			panic(err)
		}
		writer.Write(byteError)
		return
	}

	// Sort all users
	if orderBy != OrderByAsIs {
		sortUsers(orderField, users, orderBy)
	}

	// Filter all users
	resultUsers := filterByQuery(query, users)

	// Offset and limit
	var finish int
	var start int
	if len(resultUsers) < offset {
		resultUsers = make([]User, 0, 0)
	} else {
		start = offset
		if len(resultUsers) < start+limit {
			finish = len(resultUsers)
		} else {
			finish = start + limit
		}
	}
	byteResponse, err := json.Marshal(resultUsers[start:finish])
	if err != nil {
		panic(err)
	}
	writer.Write(byteResponse)

}

////////////////////////////////

func TestWithIncorrectLimitValue(t *testing.T) {
	client := &SearchClient{
		AccessToken: "",
		URL:         "",
	}

	request := &SearchRequest{Limit: -1}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if err.Error() != "limit must be > 0" {
		t.Errorf("unexpected error message. Expected %v, but got %v", "limit must be > 0", err.Error())
	}
}

func TestWithIncorrectOffsetValue(t *testing.T) {
	client := &SearchClient{
		AccessToken: "",
		URL:         "",
	}

	request := &SearchRequest{Offset: -1}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if err.Error() != "offset must be > 0" {
		t.Errorf("unexpected error message. Expected %v, but got %v", "offset must be > 0", err.Error())
	}
}

func TestWithOffsetValue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      1,
		Offset:     10,
		Query:      "",
		OrderField: "Id",
		OrderBy:    OrderByAsIs,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if !result.NextPage {
		t.Error("next page is `false`, but should be `true`")
	}
	if len(result.Users) != 1 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 1, len(result.Users))
	}
	if result.Users[0].Id != 10 {
		t.Error("incorrect result")
	}
	ts.Close()
}

func TestWithQueryValue(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      100,
		Offset:     0,
		Query:      "eu",
		OrderField: "Id",
		OrderBy:    OrderByAsIs,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if result.NextPage {
		t.Error("next page is `true`, but should be `false`")
	}
	if len(result.Users) != 24 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 2, len(result.Users))
	}
	for _, user := range result.Users {
		if match, _ := regexp.MatchString("eu", user.Name); match {
			continue
		}
		if match, _ := regexp.MatchString("eu", user.About); match {
			continue
		}
		t.Errorf("incorrect result")
	}
	ts.Close()
}

func TestWithCorrectParametersSortById(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      100,
		Offset:     0,
		Query:      "",
		OrderField: "Id",
		OrderBy:    OrderByAsIs,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if !result.NextPage {
		t.Error("next page is `false`, but should be `true`")
	}
	if len(result.Users) != 25 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 25, len(result.Users))
	}
	ts.Close()
}

func TestWithTimeoutError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := TimeoutToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if match, _ := regexp.MatchString("timeout for", err.Error()); !match {
		t.Errorf("unexpected error message. Should contains %v, but got %v", "timeout for", err.Error())
	}
	ts.Close()
}

func TestWithUnknownError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := UnknownErrorToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if match, _ := regexp.MatchString("unknown error", err.Error()); !match {
		t.Errorf("unexpected error message. Should contains %v, but got %v", "unknown error", err.Error())
	}
	ts.Close()
}

func TestWithUnauthorizedError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := UnauthorizedToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if err.Error() != "Bad AccessToken" {
		t.Errorf("unexpected error message. Expected %v, but got %v", "Bad AccessToken", err.Error())
	}
	ts.Close()
}

func TestWithInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := InternalServerErrorToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if err.Error() != "SearchServer fatal error" {
		t.Errorf("unexpected error message. Expected %v, but got %v", "SearchServer fatal error", err.Error())
	}
	ts.Close()
}

func TestWithBadRequestWithEmptyResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := BadRequestToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if match, _ := regexp.MatchString("cant unpack error json", err.Error()); !match {
		t.Errorf("unexpected error message. Should contains %v, but got %v", "cant unpack error json", err.Error())
	}
	ts.Close()
}

func TestWithIncorrectOrderField(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		OrderField: "test",
	}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if err.Error() != "OrderFeld test invalid" {
		t.Errorf("unexpected error message. Expected %v, but got %v", "OrderFeld test invalid", err.Error())
	}
	ts.Close()
}

func TestWithUnknownBadRequestErrorError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := UnknownBadRequestErrorToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if err.Error() != "unknown bad request error: someError" {
		t.Errorf("unexpected error message. Expected %v, but got %v", "unknown bad request error: someError", err.Error())
	}
	ts.Close()
}

func TestWithReturningDifferentStruct(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := DifferentStructToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{}

	result, err := client.FindUsers(*request)
	if result != nil {
		t.Errorf("result should be nil, but it was %v", result)
	}
	if err == nil {
		t.Error("expected error, but it didn't happen")
	}
	if match, _ := regexp.MatchString("cant unpack result json", err.Error()); !match {
		t.Errorf("unexpected error message. Should contains %v, but got %v", "cant unpack result json", err.Error())
	}
	ts.Close()
}

func TestWithCorrectParametersSortByIdDesc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "Id",
		OrderBy:    OrderByDesc,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if !result.NextPage {
		t.Error("next page is `false`, but should be `true`")
	}
	if len(result.Users) != 2 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 25, len(result.Users))
	}
	if result.Users[0].Id > result.Users[1].Id {
		t.Error("incorrect order of users")
	}
	ts.Close()
}

func TestWithCorrectParametersSortByIdAsc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "Id",
		OrderBy:    OrderByAsc,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if !result.NextPage {
		t.Error("next page is `false`, but should be `true`")
	}
	if len(result.Users) != 2 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 25, len(result.Users))
	}
	if result.Users[0].Id < result.Users[1].Id {
		t.Error("incorrect order of users")
	}
	ts.Close()
}

func TestWithCorrectParametersSortByNameDesc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "Name",
		OrderBy:    OrderByDesc,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if !result.NextPage {
		t.Error("next page is `false`, but should be `true`")
	}
	if len(result.Users) != 2 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 25, len(result.Users))
	}
	if result.Users[0].Name > result.Users[1].Name {
		t.Error("incorrect order of users")
	}
	ts.Close()
}

func TestWithCorrectParametersSortByNameAsc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "Name",
		OrderBy:    OrderByAsc,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if !result.NextPage {
		t.Error("next page is `false`, but should be `true`")
	}
	if len(result.Users) != 2 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 25, len(result.Users))
	}
	if result.Users[0].Name < result.Users[1].Name {
		t.Error("incorrect order of users")
	}
	ts.Close()
}

func TestWithCorrectParametersSortByAgeDesc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "Age",
		OrderBy:    OrderByDesc,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if !result.NextPage {
		t.Error("next page is `false`, but should be `true`")
	}
	if len(result.Users) != 2 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 25, len(result.Users))
	}
	if result.Users[0].Age > result.Users[1].Age {
		t.Error("incorrect order of users")
	}
	ts.Close()
}

func TestWithCorrectParametersSortByAgeAsc(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	accessToken := SomeGoodToken
	client := &SearchClient{
		AccessToken: accessToken,
		URL:         ts.URL,
	}

	request := &SearchRequest{
		Limit:      2,
		Offset:     0,
		Query:      "",
		OrderField: "Age",
		OrderBy:    OrderByAsc,
	}

	result, err := client.FindUsers(*request)
	if err != nil {
		t.Errorf("unexpected error: %#v", err)
	}
	if !result.NextPage {
		t.Error("next page is `false`, but should be `true`")
	}
	if len(result.Users) != 2 {
		t.Errorf("incorrect count of returned users. Should be %v, but result has %v", 25, len(result.Users))
	}
	if result.Users[0].Age < result.Users[1].Age {
		t.Error("incorrect order of users")
	}
	ts.Close()
}
