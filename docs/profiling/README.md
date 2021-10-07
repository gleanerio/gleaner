# Profiling

## About 

From time to time I do some profiling to see where the larger
memory activity is.  


## Code

In the code there is a profile section in the main that can 
be commented in/out.  You will need to make sure the 
imports include the following.

```
runtime/trace
github.com/pkg/profile
```


## Commands

With the code uncommented (it could be hidden in a flag too) just
run the code as normal.  The profiles will be generated.  

```
 go tool pprof -png mem.pprof > out.png

```
