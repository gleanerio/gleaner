# Flags

```go
flag.BoolVar(&setupVal, "setup", false, "Run Gleaner configuration check and exit")
flag.StringVar(&viperVal, "cfg", "config", "Configuration file (can be YAML, JSON) Do NOT provide the extension in the command line. -cfg file not -cfg file.yml")
flag.StringVar(&sourceVal, "source", "", "Override config file source(s) to specify an index target")
flag.BoolVar(&stdoutVal, "stdout", false, "Send log output to stdout and file (default only file)")
flag.StringVar(&modeVal, "mode", "full", "Set the mode (full | diff) to index all or just diffs")
```
