package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"sync"
)

// сюда писать код
func ExecutePipeline(jobs ...job) {
	out := make(chan interface{})
	wg := sync.WaitGroup{}

	for _, currentJob := range jobs {
		currentJob := currentJob

		// make input channel for this job
		// from last job's output channel
		currentIn := out
		// make output channel for this job
		currentOut := make(chan interface{})

		// increase counter before call job
		wg.Add(1)

		// should set WaitGroup as a variable to this func
		go func(wg *sync.WaitGroup) {
			// decrease counter after this func finishes
			defer wg.Done()

			// call job
			currentJob(currentIn, currentOut)

			// also need close current output channel
			close(currentOut)
		}(&wg)
		// set current output channel like a global output channel
		// on the next iteration next job will use it like a input channel
		out = currentOut
	}
	// wait all jobs will finish
	wg.Wait()
}

var mu sync.Mutex

func SingleHash(in chan interface{}, out chan interface{}) {
	wg := sync.WaitGroup{}
	for nextInputVar := range in {
		wg.Add(1)
		go func(nextInputVar interface{}, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()

			str := getStringFromInterface(nextInputVar)

			res1 := make(chan string, 1)
			go func(result chan string, input string) {
				defer close(result)
				result <- DataSignerCrc32(input)
			}(res1, str)
			res2 := make(chan string, 1)
			go func(result chan string, input string) {
				defer close(result)
				mu.Lock()
				md5 := DataSignerMd5(input)
				mu.Unlock()
				result <- DataSignerCrc32(md5)
			}(res2, str)

			result := fmt.Sprint(<-res1, "~", <-res2)
			out <- result
		}(nextInputVar, out, &wg)
		runtime.Gosched()
	}
	wg.Wait()
}

func MultiHash(in chan interface{}, out chan interface{}) {
	wg := sync.WaitGroup{}
	for nextInputVar := range in {
		wg.Add(1)
		go func(nextInputVar interface{}, out chan interface{}, wg *sync.WaitGroup) {
			defer wg.Done()

			str := getStringFromInterface(nextInputVar)

			wg1 := sync.WaitGroup{}
			resSlice := make([]string, 6, 6)
			for index := range resSlice {
				wg1.Add(1)
				go func(input string, index int, res []string, wg *sync.WaitGroup) {
					defer wg.Done()
					res[index] = DataSignerCrc32(fmt.Sprint(strconv.FormatInt(int64(index), 10), str))
				}(str, index, resSlice, &wg1)
			}
			wg1.Wait()
			var res string
			for _, currentString := range resSlice {
				res = fmt.Sprintf("%v%v", res, currentString)
			}
			out <- res
		}(nextInputVar, out, &wg)
		runtime.Gosched()
	}
	wg.Wait()
}

func CombineResults(in chan interface{}, out chan interface{}) {
	sl := make([]string, 0, 0)
	for nextInputVar := range in {
		str := getStringFromInterface(nextInputVar)
		sl = append(sl, str)
		runtime.Gosched()
	}
	sort.Strings(sl)
	result := ""
	for _, currentString := range sl {
		if result == "" {
			result = currentString
		} else {
			result = fmt.Sprintf("%v_%v", result, currentString)
		}
	}
	out <- result
}

func getStringFromInterface(in interface{}) string {
	var str string

	switch val := in.(type) {
	case int:
		str = strconv.FormatInt(int64(val), 10)
	case string:
		str = val
	default:
		panic(fmt.Sprintf("Invalid type of input data. Expected int or string vut got %T\n", val))
	}
	return str
}
