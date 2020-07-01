package service

import (
	"fmt"
	_ "sync"
	"time"
)

// code goes here

func Log(msg string) {
	fmt.Println(msg)
}

func ExecutePipeline(jobs ...job) {
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
	time.Sleep(1 * time.Second)
}

func SingleHash(in, out chan interface{}) {

}

func MultiHash(in, out chan interface{}) {

}

func CombineResults(in, out chan interface{}) {

}