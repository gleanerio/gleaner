Demonstration to run against samples.earth

### demo.yaml is a hand generated file
run demo using gleaner
```
gleaner -cfg configs/demo/demo
```
Note: no .yaml extension  at the end of the file

run demo using glcon
```
glcon gleaner batch -cfgFile  configs/demo/demo
```

### gleaner configuration directory mode
After running glcon config

```
glcon config generate  -cfgName  myDemo
```

gleaner:
```
gleaner -cfg configs/demo/gleaner
```
glcon:
```
glcon gleaner batch -cfgName  demo
```
Note: No filename needed. assumed to be gleaner

