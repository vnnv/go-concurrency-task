In this task we are going to build something like unix pipeline:
```
grep 127.0.0.1 | awk '{print $2}' | sort | uniq -c | sort -nr
```

STDOUT of one program is passing as STDIN of next program

But in our case for this role go chanels are being used. So we are passing channels from one function to another.

The task consists of two parts:
* To write function ExecutePipeline which provides pipeline of worker-functions, which performs some calculations
* To implement that workers. They performing some hash calculations on input data with help of functions from common.go

The final hashsum is result of next functions call chain (inside ExecutePipeline):
* SingleHash calculates crc32(data)+"~"+crc32(md5(data)) hash ( concat of two strings with ~ symbol), where data is that what comes in "in" chanel (its numbers from first function in pipeline)
* MultiHash calculates crc32(th+data)) hash (concat of number asserted to string and a string), where th=0..5 ( i.e. 6 hashes on every input number), then concats the results in order as they come (0..5), where the data is that what comes in "in" chanel
* CombineResults receives all results, sorts (https://golang.org/pkg/sort/), concats results with _ symbol to one string
* crc32 calculates via DataSignerCrc32 (in common.go)
* md5 calculates via DataSignerMd5  (in common.go)

What the trick?
* DataSignerMd5 can only be called just 1 time and calculating for 10 ms. If running simultaneously - there will be overheat to 1 second
* DataSignerCrc32, calculating for 1 second
* We have 3 seconds totals to perform all calculations
* In bold approach (linear) it would take 57 seconds for 7 elements, so process should be paralleled somehow

Here are the logs of program with 2 input values (input is commented out in a test, you should implement logs yourself):

```
0 SingleHash data 0
0 SingleHash md5(data) cfcd208495d565ef66e7dff9f98764da
0 SingleHash crc32(md5(data)) 502633748
0 SingleHash crc32(data) 4108050209
0 SingleHash result 4108050209~502633748
4108050209~502633748 MultiHash: crc32(th+step1)) 0 2956866606
4108050209~502633748 MultiHash: crc32(th+step1)) 1 803518384
4108050209~502633748 MultiHash: crc32(th+step1)) 2 1425683795
4108050209~502633748 MultiHash: crc32(th+step1)) 3 3407918797
4108050209~502633748 MultiHash: crc32(th+step1)) 4 2730963093
4108050209~502633748 MultiHash: crc32(th+step1)) 5 1025356555
4108050209~502633748 MultiHash result: 29568666068035183841425683795340791879727309630931025356555

1 SingleHash data 1
1 SingleHash md5(data) c4ca4238a0b923820dcc509a6f75849b
1 SingleHash crc32(md5(data)) 709660146
1 SingleHash crc32(data) 2212294583
1 SingleHash result 2212294583~709660146
2212294583~709660146 MultiHash: crc32(th+step1)) 0 495804419
2212294583~709660146 MultiHash: crc32(th+step1)) 1 2186797981
2212294583~709660146 MultiHash: crc32(th+step1)) 2 4182335870
2212294583~709660146 MultiHash: crc32(th+step1)) 3 1720967904
2212294583~709660146 MultiHash: crc32(th+step1)) 4 259286200
2212294583~709660146 MultiHash: crc32(th+step1)) 5 2427381542
2212294583~709660146 MultiHash result: 4958044192186797981418233587017209679042592862002427381542

CombineResults 29568666068035183841425683795340791879727309630931025356555_4958044192186797981418233587017209679042592862002427381542
```

Code should be written in signer.go 

To test program run u can use `go test -v -race`

Tips:
* task is made in a way to test different approaches on paralelling, all covered by most basic tutorials for go parallelism tasks
* You should not "hoard" data inside workers - pass it as it comes right to the next function. It should work like awk in topmost example of linux pipeline
* Think on how is organized proper function end if the data flow is finite. What should be done to achieve that?
* If you faced race condition (build/test with -race flag) - examine cli output - there must be useful info to debug the race error
* Before paralelize all the things you should write linear bold code that performs correct calculations. Then parallel it
* you can expect that there might be not more than 100 input elements.
* Answer to question "when is closing chanel loop?" helps to implement ExecutePipeline function
* Answer to question "am I need results of previous calculations?" helps to parallelize SingleHash & MultiHash
* It may be useful to make some chart of calculations
* It may be useful to take a look at test code to understand how's test will be performed