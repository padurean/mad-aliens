package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/padurean/mad-aliens/internal/invasion"
	"github.com/padurean/mad-aliens/internal/world"
)

var (
	ColorReset = "\033[0m"
	ColorGreen = "\033[32m"
	ColorRed   = "\033[31m"
)

func init() {
	// do not use colors on Windows as they might not work
	if runtime.GOOS == "windows" {
		ColorReset = ""
		ColorGreen = ""
		ColorRed = ""
	}
}

type flags struct {
	help        bool
	interactive bool
}

func (f *flags) parse() {
	flag.BoolVar(&f.help, "help", f.help, "show help")
	flag.BoolVar(&f.interactive, "i", f.interactive, "run the invasion interactively (step by step)")
	flag.Parse()
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: invasion [OPTIONS] <number of aliens>\nOPTIONS:\n")
		flag.PrintDefaults()
	}

	var f flags
	f.parse()

	if f.help {
		flag.Usage()
		return
	}

	var (
		n   int
		err error
	)

	args := flag.Args()
	if len(args) < 1 {
		fmt.Printf("Please specify the number of 👽 aliens: ")
		_, err = fmt.Scanf("%d", &n)
	} else {
		n, err = strconv.Atoi(args[0])
	}

	if err != nil {
		fmt.Printf("Failed to read the number of 👽 aliens: %v\n", err)
		os.Exit(1)
	}

	if n == 0 {
		fmt.Println("Zero 👽 aliens => no invasion! 🎉")
		os.Exit(0)
	}

	interactiveMode := ColorRed + "🔴OFF" + ColorReset
	if f.interactive {
		interactiveMode = ColorGreen + "🟢ON" + ColorReset
	}
	fmt.Printf("Interactive mode: %s\n", interactiveMode)

	worldIn := "world.txt"
	var w world.World
	if err := w.Read(worldIn); err != nil {
		fmt.Printf("Failed to read world from file '%s': %v", worldIn, err)
		os.Exit(2)
	}

	onEvent := func(event string) {
		fmt.Println(event)
		if f.interactive {
			fmt.Print(ColorGreen + "Press 'Enter' to continue ..." + ColorReset)
			bufio.NewReader(os.Stdin).ReadBytes('\n')
		}
	}

	inv := invasion.New(w, n, onEvent)
	fmt.Println(inv.Run())

	worldOut := "world_after_invasion.txt"
	if err := inv.World.Write(worldOut); err != nil {
		fmt.Printf("Failed to write world to file '%s': %v", worldOut, err)
		os.Exit(3)
	}

	fmt.Println("The End. 🏁")
}
