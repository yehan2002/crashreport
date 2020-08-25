# CrashReport

A Simple package for writing crash reports.

## Usage

### Creating crash reports

```golang
func LoadTexture(path string){
    err := TextureManager.Load(path)
    if err == nil{
      report := crashreport.Crash("Invalid texture loaded").Include(path)
      report.WriteTo("./crashreport.crash")
    }
}

```

### Viewing crash reports

`$crashreport -browser ./path/to/crash/file.zip`
