# MD5 All

Blazingly fast cli application for checking md5 sum of a specific file or directory

## How to run it?

Print md5 sums to the terminal by running the binary file with the following command (you can replace `./go.mod` with any other filename or path):

```shell
./cmd/main ./go.mod
```

Output:

```
./go.mod -> 451a40ce008a3493e35efdd45b224352
```

Or save the output onto the file:

```shell
./cmd/main ./go.mod > output.txt
```

## Architecture

Inspiration for this small cli tool project is coming from this [article](https://go.dev/blog/pipelines) about "Pipeline" concurrency pattern in Golang. In this article, the [author](https://github.com/Sajmani) implements a similar tool in a <i>slightly</i> different way. After reading the article, the decision to try building it myself was made. And here it is.

I decided to use fan-out + fan-in techniques to split the work among `GOMAXPROCS` workers (goroutines), and while it might be an overkill for a folder with just a couple of files, it certainly speeds up the process of checksumming large directories.

![Concurrency model](https://i.imgur.com/XVUlU5z.png)