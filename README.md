# CrashReport

A Simple package for writing crash reports.

## Usage

### Creating crash reports

```golang
f,err := os.Create(filename)
if err == nil{
    crashreport.Crash("invalid texture loaded", f)
    f.Close()
}
```

### Viewing crash reports

`$crashreport -browser ./path/to/crash/file.zip`
