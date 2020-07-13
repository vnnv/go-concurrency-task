package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	_ "sync"
	"time"
)

// code goes here

func Log(msg string) {
	fmt.Println(msg)
}

func worker(jobToExecute job, in, out chan interface{}, wg *sync.WaitGroup, no int) {
	//defer fmt.Println("Finish job #", no)
	defer wg.Done()
	defer close(out)

	//fmt.Println("Starting job #", no)
	jobToExecute(in, out)
}

func ExecutePipeline(jobs ...job) {

	wg := sync.WaitGroup{}

	noJobs := len(jobs)
	fmt.Printf("Starting pipeline of %d jobs...\n", noJobs)

	//input := make(chan interface{})
	var input chan interface{}
	var output chan interface{}

	for i, job := range jobs {
		fmt.Printf("Executing job %d type: %T\n", i, job)
		output = make(chan interface{})

		wg.Add(1)
		go worker(job, input, output, &wg, i)

		input = output
	}

	if output != nil {
		res := <- output
		fmt.Println("result: ", res)
	}
	fmt.Println("*** Waiting to finish pipeline...")
	wg.Wait()


	fmt.Println("Pipeline is over...")
}

func convertToString(data interface{}) (string, error) {
	var res string
	switch data.(type) {
	case int: res = strconv.Itoa(data.(int))
	case float32 : res = strconv.FormatFloat(float64(data.(float32)), 'f', -1, 32)
	case float64 : res = strconv.FormatFloat(data.(float64), 'f', -1, 64)
	case string : res = data.(string)
	default:
		return "", fmt.Errorf("Can not parse the incoming data\n")
	}

	return res, nil
}

func privateCrc32(/*wg *sync.WaitGroup,*/ data string, out chan string ) {
	//defer wg.Done()

	res := DataSignerCrc32(data)
	out <- res
}

func SingleHash(in, out chan interface{}) {
	md5Lock := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	fmt.Println("Starting SingleHash()")
	for inDataRaw := range in {

		//fmt.Println("* SingleHash start: ", time.Now())
		data, err := convertToString(inDataRaw)
		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("Single Hash input data string: ", data)
		wg.Add(1)
		go func(inData string, wg *sync.WaitGroup) {
			defer wg.Done()

			md5Lock.Lock()
			md5 := DataSignerMd5(data)
			md5Lock.Unlock()

			leftChannel := make(chan string)
			rightChannel := make(chan string)
			//wgInt := &sync.WaitGroup{}

			//wgInt.Add(1)
			go privateCrc32(/*wgInt,*/ data, leftChannel)
			//wgInt.Add(1)
			go privateCrc32(/*wgInt,*/ md5, rightChannel)


			left := <- leftChannel
			right := <- rightChannel

			//wgInt.Wait()
			res := left + "~" + right

			fmt.Println("Single Hash crc32(data) ", left)
			fmt.Println("Single Hash md5(data) ", md5)
			fmt.Println("Single Hash crc32(md5(data)) ", right)
			fmt.Println("Single Hash final res: ", res)
			fmt.Println("* SingleHash end: ", time.Now())

			out <- res

		}(data, wg)

	}
	fmt.Println("Waiting to finish SingleHash()")
	wg.Wait()
	fmt.Println("****** SingleHash() end")
}

func MultiHash(in, out chan interface{}) {
	//for {
	wg := &sync.WaitGroup{}
		for inDataRaw := range in {
			data, err := convertToString(inDataRaw)

			if err != nil {
				fmt.Println(err)
				return
			}
			//fmt.Println("Multi hash input res: ", data, time.Now())
			wg.Add(1)
			go func(data string, outer_wg *sync.WaitGroup) {
				defer outer_wg.Done()
				var result string
				// go crc32 -> res goes to map[i] = res
				res := make(map[int]string)
				mutex := &sync.Mutex{}

				wg := &sync.WaitGroup{}

				for i := 0 ; i < 6; i++ {
					//res[i] = make(chan string)
					_data := strconv.Itoa(i) + data
					wg.Add(1)
					go func(index int, input string) {
						defer mutex.Unlock()
						defer wg.Done()

						result := DataSignerCrc32(input)
						mutex.Lock()
						res[index] = result

					}(i, _data)

					//go privteCrc32(wg, _data, res[i])
					//crc32 := DataSignerCrc32(_data)
					//fmt.Println(data, "Multi hash crc32(th+data) ", i, crc32)
					//result += crc32
				}
				wg.Wait()

				for i := 0; i < 6; i++ {
					s := res[i]
					fmt.Println(data, "Multi hash crc32(th+data) ", i, s)
					result += s
				}
				//fmt.Println(data, "Multihash result: ", result, time.Now())
				out <- result
			}(data, wg)
			//return
		}
	//}
	wg.Wait()

}

func CombineResults(in, out chan interface{}) {
	var tempSlice []string

	var res string
	for inDataRaw := range in {
		data, err := convertToString(inDataRaw)

		if err != nil {
			fmt.Println(err)
			return
		}
		//fmt.Println("Combine result res: ", data)
		tempSlice = append(tempSlice, data)
		res += data
	}
	sort.Strings(tempSlice)
	endResult := fmt.Sprintf(strings.Join(tempSlice[:], "_"))
	fmt.Println("Combine result final: ", endResult)
	out <- endResult

}


//func ExecutePipelineNotWork(jobs ...job) {
//
//	noJobs := len(jobs)
//	fmt.Printf("Starting pipelineof %d jobs...\n", noJobs)
//
//	input := make(chan interface{})
//	output := make(chan interface{})
//
//	var res interface{}
//
//	for i, job := range jobs {
//		fmt.Printf("Executing job %d type: %T\n", i, job)
//		go job(input, output)
//
//		out:
//
//			for {
//			select {
//			case res = <- output:
//				if i != 0 {
//					input <- res
//					res = nil
//				}
//				break
//			default:
//				if i == noJobs - 1 {
//					time.Sleep(10*time.Millisecond)
//					break out
//				}
//			}
//
//		} // end for
//
//
//
//	}
//	//res := <- output
//	fmt.Println("Pipeline is over...")
//}

//func ExecutePipelineWork(jobs ...job) {
//	//wg := new(sync.WaitGroup)
//
//	noJobs := len(jobs)
//	fmt.Printf("Starting pipelineof %d jobs...\n", noJobs)
//
//	input := make(chan interface{})
//	output := make(chan interface{})
//
//	var res interface{}
//
//	for i, job := range jobs {
//		fmt.Printf("Executing job %d type: %T\n", i, job)
//		// wg.Add(1)
//		go job(input, output)
//
//		if i != 0 && res != nil {
//			input <- res
//		}
//
//		if i < noJobs - 1 {
//			res = <- output
//			//r = res
//			fmt.Println("intermediate result is: ", res)
//			//input = output
//			//output = make(chan interface{})
//		}else{
//			fmt.Printf("Last job is started\n")
//		}
//
//	}
//	//res := <- output
//	fmt.Println("Pipeline is over...")
//}
