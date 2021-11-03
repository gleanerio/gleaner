These should be used so that enviroment variables, and other defaults
can be read.

*Config.yaml are the config files 
 
others can be submitted as subtress.
Subtress cannot be extracted when they are arrays, aka v1.Sub('sources'),
so passing in the root may be a better solution, but env substitute is 
more reliable without the dot path