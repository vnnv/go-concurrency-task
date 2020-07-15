package service

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// code goes here

func worker(jobToExecute job, in, out chan interface{}, wg *sync.WaitGroup, no int) {
	defer wg.Done()
	defer close(out)
	defer log.Println("Finish job #", no)

	log.Println("Starting job #", no)
	jobToExecute(in, out)
}

// ExecutePipeline accept slice of jobs and execute every job
func ExecutePipeline(jobs ...job) {
	log.Printf("*** Starting pipeline of %d jobs...\n", len(jobs))

	wg := sync.WaitGroup{}

	var input chan interface{}
	var output chan interface{}

	for i, job := range jobs {
		log.Printf("Executing job %d... \n", i)
		output = make(chan interface{})

		wg.Add(1)
		go worker(job, input, output, &wg, i)

		input = output
	}

	if output != nil {
		res := <-output
		log.Println("Last job result is: ", res)
	}
	log.Println("Waiting to finish pipeline...")
	wg.Wait()

	log.Println("*** Pipeline is over...")
}

// Convert only from int, float and string. For other types return error.
// Can be used if the test contains different type input values
func convertToString(data interface{}) (string, error) {
	var res string
	switch data := data.(type) {
	case int:
		res = strconv.Itoa(data)
	case float32:
		res = strconv.FormatFloat(float64(data), 'f', -1, 32)
	case float64:
		res = strconv.FormatFloat(data, 'f', -1, 64)
	case string:
		res = data
	default:
		return "", fmt.Errorf("Can not parse the incoming data. Not supported variable type.\n")
	}

	return res, nil
}

// SingleHash - to be described
func SingleHash(in, out chan interface{}) {
	md5Lock := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	log.Println("*** Start SingleHash()...")
	for inDataRaw := range in {

		data, err := convertToString(inDataRaw)
		if err != nil {
			fmt.Println(err)
			return
		}
		wg.Add(1)
		go func(inData string, wg *sync.WaitGroup) {
			defer wg.Done()

			// Lock calling md5 - to prevent overheating "_
			md5Lock.Lock()
			md5 := DataSignerMd5(data)
			md5Lock.Unlock()

			leftChannel := make(chan string)
			rightChannel := make(chan string)

			// Calc left part
			go func(data string, out chan string) {
				res := DataSignerCrc32(data)
				out <- res
			}(inData, leftChannel)

			// Calc right part
			go func(data string, out chan string) {
				res := DataSignerCrc32(data)
				out <- res
			}(md5, rightChannel)

			left := <-leftChannel
			right := <-rightChannel

			res := left + "~" + right

			log.Println("Single Hash input data ", inData)
			log.Println("Single Hash crc32(data) ", left)
			log.Println("Single Hash md5(data) ", md5)
			log.Println("Single Hash crc32(md5(data)) ", right)
			log.Println("Single Hash final result: ", res)

			out <- res

		}(data, wg)

	}
	log.Println("Waiting to finish SingleHash()")
	wg.Wait()
	log.Println("*** SingleHash() finished")
}

// MultiHash - to be described
func MultiHash(in, out chan interface{}) {
	log.Println("*** Start MultiHash()...")
	wg := &sync.WaitGroup{}

	for inDataRaw := range in {
		data, err := convertToString(inDataRaw)

		if err != nil {
			log.Panic(err)
		}
		wg.Add(1)
		go func(data string, outerWg *sync.WaitGroup) {
			defer outerWg.Done()

			log.Println("MultiHash() input data: ", data)
			var result string

			// keep the results is a map
			res := make(map[int]string)

			mutex := &sync.Mutex{}
			wg := &sync.WaitGroup{}

			for i := 0; i < 6; i++ {
				wg.Add(1)
				go func(index int, input string) {
					defer mutex.Unlock()
					defer wg.Done()

					result := DataSignerCrc32(input)

					mutex.Lock()
					res[index] = result
				}(i, strconv.Itoa(i)+data)

			}
			wg.Wait()

			for i := 0; i < 6; i++ {
				s := res[i]
				log.Println("Input: ", data, "MultiHash() crc32(th+data) ", i, s)
				result += s
			}
			out <- result
		}(data, wg)
	}
	log.Println("Waiting MultiHash() to finish...")
	wg.Wait()
	log.Println("*** End MultiHash()")

}

// CombineResults combines and sort incoming values from MultiHash function
func CombineResults(in, out chan interface{}) {
	var tempSlice []string

	var res string
	for inDataRaw := range in {
		data, err := convertToString(inDataRaw)

		if err != nil {
			fmt.Println(err)
			return
		}
		tempSlice = append(tempSlice, data)
		res += data
	}
	sort.Strings(tempSlice)
	endResult := strings.Join(tempSlice[:], "_")
	log.Println("Combine result final: ", endResult)
	out <- endResult
}
