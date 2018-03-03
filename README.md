# Edgar Getter
Edgar Getter is a tool written in Go for downloading company periodic reports, 
filings and forms from Securities and Exchange Commission (SEC) EDGAR site.

# Installation
`go get github.com/RelateTheNews/go-edgar-getter`
# Usage
```go
import getter

func main(){
  var g Getter
  var getURI string
  
  // Note this is a sample URI. Must verify correct URIs on www.sec.gov
  getURI = "https://www.sec.gov/Archives/edgar/Feed/2013/QTR1/"
  
  g.NewGetter()
  
  // files is a list of successfully retrieved files
  files := g.RetrieveURIs(getURI, 0)
}
```

# Contribution Guidelines
Contributions are greatly appreciated. The maintainers actively manage the issues list. 
For a list of priary mainters see [./MAINTAINERS.md](./MAINTAINERS.md). The project follows the typical GitHub pull request model.
Before starting any work, please either comment on an existing issue, or file a new one.
