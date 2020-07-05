package service

import (
	"fmt"
	"strconv"
	"sync"
	_ "sync"
	"time"
)

// code goes here

func Log(msg string) {
	fmt.Println(msg)
}

func worker(jobToExecute job, in, out chan interface{}, wg *sync.WaitGroup, no int) {
	defer fmt.Println("Finish job #", no)
	defer wg.Done()
	defer close(out)

	fmt.Println("Starting job #", no)
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

	//if output != nil {
	//	res := <- output
	//	fmt.Println("* FIN result: ", res)
	//}
	wg.Wait()

	//select {
	//case res := <- output:
	//	fmt.Println("The result of the last function is ", res)
	//	time.Sleep(10*time.Millisecond)
	//	break
	//default:
	//	time.Sleep(100*time.Millisecond)
	//	break
	//}
	//time.Sleep(1*time.Second)
	//res := <- output
	fmt.Println("Pipeline is over...")
}

func ExecutePipelineNotWork(jobs ...job) {

	noJobs := len(jobs)
	fmt.Printf("Starting pipelineof %d jobs...\n", noJobs)

	input := make(chan interface{})
	output := make(chan interface{})

	var res interface{}

	for i, job := range jobs {
		fmt.Printf("Executing job %d type: %T\n", i, job)
		go job(input, output)

		out:

			for {
			select {
			case res = <- output:
				if i != 0 {
					input <- res
					res = nil
				}
				break
			default:
				if i == noJobs - 1 {
					time.Sleep(10*time.Millisecond)
					break out
				}
			}

		} // end for



	}
	//res := <- output
	fmt.Println("Pipeline is over...")
}

func ExecutePipelineWork(jobs ...job) {
	//wg := new(sync.WaitGroup)

	noJobs := len(jobs)
	fmt.Printf("Starting pipelineof %d jobs...\n", noJobs)

	input := make(chan interface{})
	output := make(chan interface{})

	var res interface{}

	for i, job := range jobs {
		fmt.Printf("Executing job %d type: %T\n", i, job)
		// wg.Add(1)
		go job(input, output)

		if i != 0 && res != nil {
			input <- res
		}

		if i < noJobs - 1 {
			res = <- output
			//r = res
			fmt.Println("intermediate result is: ", res)
			//input = output
			//output = make(chan interface{})
		}else{
			fmt.Printf("Last job is started\n")
		}

	}
	//res := <- output
	fmt.Println("Pipeline is over...")
}

func convertToString(data interface{}) (string, error) {
	var res string
	switch data.(type) {
	case int: res = strconv.Itoa(data.(int))
	case float32 : res = fmt.Sprintf("%f", data.(float32))
	case float64 : res = fmt.Sprintf("%f", data.(float64))
	case string : res = data.(string)
	default:
		return "", fmt.Errorf("Can not parse the incoming data\n")
	}

	return res, nil
}

func SingleHash(in, out chan interface{}) {
	for inDataRaw := range in {
		data, err := convertToString(inDataRaw)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("In data string: ", data)

		left := DataSignerCrc32(data)
		right := DataSignerCrc32(DataSignerMd5(data))
		res := left + "~" + right
		fmt.Println("SingleHash res: ", res)
		out <- res
	}


}

func MultiHash(in, out chan interface{}) {
	for inDataRaw := range in {
		data, err := convertToString(inDataRaw)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Multihash res: ", data)
		out <- data
	}
}

func CombineResults(in, out chan interface{}) {
	var res string
	for inDataRaw := range in {
		data, err := convertToString(inDataRaw)

		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Combine result res: ", data)
		res += data
	}
	out <- res

}