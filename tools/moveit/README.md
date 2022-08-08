# moveit

## About

A simple testing code to see what the process of moving 
objects from 

```
.../summoned/sourcex
to 
.../archive/sourcex/[DATE]
```

might look like.

* do nothing in prov
* copy summoned to the archive
* empty summoned and milled
* do next index

Q1: What if the index fails, I already have things in the archive.
    - prov could be used to recover from the archive (but that code is not written)
    - If you have already loaded to the graph, just don't on index failure
    - check the sitemap first to ensure site is on-line and active first
