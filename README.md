---

# GPUtil

GPUtil: Golang Implementation based on https://github.com/anderskm/gputil (Python Version)

## Overview

This project is a Golang implementation of the functionality provided by https://github.com/anderskm/gputil in Python. It offers some useful features that can help you accomplish specific tasks.

## Features

- Feature 1: List GPU informations
- Feature 2: List processes having compute context on the device

## Installation

To install and run this project, follow these steps:

```bash
$ go get -u github.com/lichunqiang/gputil
```

## Usage Example

Here is an example code snippet demonstrating the usage of this project:

```go
package main

import (
   "context"
   "fmt"
   "github.com/lichunqiang/gputil"
)

func main() {
   ctx := context.Background()
   gpus, err := gputil.GetGPUs(ctx)
   if err != nil {
      panic(err)
   }
   for _, item := range gpus {
      fmt.Println(item.String())
   }
}
```

## Contribution Guidelines

If you would like to contribute to this project, please follow these steps:

1. Fork the project to your GitHub account.
2. Clone the project to your local machine:
   ```
   git clone https://github.com/lichunqiang/gputil.git
   ```
3. Create a new branch:
   ```
   git checkout -b feature/your-feature
   ```
4. Make your modifications and improvements.
5. Commit your changes:
   ```
   git commit -m "Add your commit message"
   ```
6. Push your changes to the remote repository:
   ```
   git push origin feature/your-feature
   ```
7. Create a Pull Request and wait for review and merge.

## License

This project is licensed under the [MIT License](https://opensource.org/licenses/MIT).
