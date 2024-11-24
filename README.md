# segb
## a golang package for decoding apple's proprietary segb format

### Usage
Currently, to use this project as a command line tool, you can clone the repository and run the CLI module with a path to the SEGB file to decode.
```bash
git clone https://github.com/bluefalconhd/segb
cd segb
go run ./cli /path/to/your/file.segb
```

Otherwise, you can use the package in your own project by importing it and calling the `Decode` function with a streaam of the SEGB data.
```go
package main

import (
    "fmt"
    "os"
    "github.com/bluefalconhd/segb"
)

func main() {
    file, err := os.Open("/path/to/your/file.segb")
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    data, err := segb.Decode(file)
    if err != nil {
        fmt.Println("Error decoding file:", err)
        return
    }

    fmt.Println(data.Version)
}
```

### Reference
As a resource for any curious people looking to learn more about the SEGB file format, I have created a document that outlines the format and how it is structured. You can find it [here](segb.md).

### License
This project is licensed under the MIT License. See the [LICENSE.md](LICENSE.md) file for more information.