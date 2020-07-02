package service

import (
	"fmt"
	"sync"
	_ "sync"
	"time"
)

// code goes here

func Log(msg string) {
	fmt.Println(msg)
}

func worker(jobToExecute job, in, out chan interface{}, wg *sync.WaitGroup, no int) {
	defer wg.Done()
	defer close(out)
	defer fmt.Println("Finish job #", no)

	fmt.Println("Starting job #", no)
	jobToExecute(in, out)
}

func ExecutePipeline(jobs ...job) {

	wg := sync.WaitGroup{}

	noJobs := len(jobs)
	fmt.Printf("Starting pipelineof %d jobs...\n", noJobs)

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

func SingleHash(in, out chan interface{}) {

}

func MultiHash(in, out chan interface{}) {

}

func CombineResults(in, out chan interface{}) {

}