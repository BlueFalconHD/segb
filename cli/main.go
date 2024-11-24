package main

// Just a simple CLI tool to demonstrate the use of the segb-go library.
// All it does is take in a SEGB file and print out the contents.

import (
	"flag"
	"fmt"
	"github.com/bluefalconhd/segb"
	"os"
)

func PrettyHexdump(data []byte) {
	for i := 0; i < len(data); i += 16 {
		fmt.Printf("%08x: ", i)
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				fmt.Printf("%02x ", data[i+j])
			} else {
				fmt.Print("   ")
			}
		}
		fmt.Print(" ")
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				if data[i+j] >= 32 && data[i+j] <= 126 {
					fmt.Printf("%c", data[i+j])
				} else {
					fmt.Print(".")
				}
			}
		}
		fmt.Println()
	}
}

func main() {
	// Parse the command line arguments
	flag.Parse()

	// Get the filename from the command line arguments
	filename := flag.Arg(0)

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v\n", err)
		}
	}(file)

	// Decode the SEGB file
	segbData, err := segb.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding SEGB file: %v\n", err)
		return
	}

	fmt.Printf("Version: %v\n", segbData.Version)
	fmt.Printf("Created: %v\n", segbData.Created.String())
	fmt.Println("Entries:")
	for i, entry := range segbData.Entries {
		fmt.Printf("Entry %d:\n", i)
		fmt.Printf("  State: %v\n", entry.State)
		fmt.Printf("  Created: %s\n", entry.Created.String())
		PrettyHexdump(entry.Data)

		fmt.Println("--------------------")
	}
}
