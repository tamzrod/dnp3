// Command-driven memory-only outstation for the EVT-001 OpenDNP3 lab.
// It listens on 127.0.0.1 and accepts publish commands on stdin.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"dnp3/internal/al"
	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/outstation"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go-outstation <port>")
		os.Exit(2)
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil || port <= 0 || port > 65535 {
		fmt.Fprintln(os.Stderr, "bad port")
		os.Exit(2)
	}
	cfg := outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(1),
		outstation.WithTransport(dnp3.TCP, "127.0.0.1", port),
	)
	if v := os.Getenv("EVT_MAX_EVENTS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 3 {
			fmt.Fprintln(os.Stderr, "bad EVT_MAX_EVENTS")
			os.Exit(2)
		}
		cfg.MaxEventBuffers = n
	}
	server, err := outstation.NewServer(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Start(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer server.Stop(context.Background())
	fmt.Println("READY")

	in := bufio.NewScanner(os.Stdin)
	for in.Scan() {
		fields := strings.Fields(in.Text())
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "quit":
			fmt.Println("BYE")
			return
		case "bin":
			if err := publishBin(server, fields[1:]); err != nil {
				fmt.Println("ERR", err)
			} else {
				fmt.Println("OK")
			}
		case "ana":
			if err := publishAna(server, fields[1:]); err != nil {
				fmt.Println("ERR", err)
			} else {
				fmt.Println("OK")
			}
		case "ctr":
			if err := publishCtr(server, fields[1:]); err != nil {
				fmt.Println("ERR", err)
			} else {
				fmt.Println("OK")
			}
		case "fill":
			n, err := strconv.Atoi(fields[1])
			if err != nil {
				fmt.Println("ERR", err)
				continue
			}
			full := 0
			for i := 0; i < n; i++ {
				err := server.PublishBinaryEvent(uint16(i), true, al.BinEventOnline, outstation.EventClass1, time.Date(2021, 6, 7, 8, 9, 10, 0, time.UTC))
				if err != nil {
					full++
				}
			}
			fmt.Printf("FILL accepted=%d rejected=%d\n", n-full, full)
		default:
			fmt.Println("ERR unknown")
		}
	}
}

func publishBin(server outstation.Server, args []string) error {
	// bin <index> <0|1> <class 1-3> <rfc3339|none>
	if len(args) != 4 {
		return fmt.Errorf("bin args")
	}
	index, value, class, at, err := pointArgs(args)
	if err != nil {
		return err
	}
	return server.PublishBinaryEvent(index, value != 0, al.BinEventOnline, class, at)
}

func publishAna(server outstation.Server, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("ana args")
	}
	index64, err := strconv.ParseUint(args[0], 10, 16)
	if err != nil {
		return err
	}
	value, err := strconv.ParseInt(args[1], 10, 32)
	if err != nil {
		return err
	}
	class, err := classOf(args[2])
	if err != nil {
		return err
	}
	at, err := timeOf(args[3])
	if err != nil {
		return err
	}
	return server.PublishAnalogEvent(uint16(index64), int32(value), al.AnalogEventOnline, class, at)
}

func publishCtr(server outstation.Server, args []string) error {
	if len(args) != 4 {
		return fmt.Errorf("ctr args")
	}
	index64, err := strconv.ParseUint(args[0], 10, 16)
	if err != nil {
		return err
	}
	value, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return err
	}
	class, err := classOf(args[2])
	if err != nil {
		return err
	}
	at, err := timeOf(args[3])
	if err != nil {
		return err
	}
	return server.PublishCounterEvent(uint16(index64), uint32(value), al.CounterEventOnline|al.CounterEventRollover, class, at)
}

func pointArgs(args []string) (uint16, int, outstation.EventClass, time.Time, error) {
	index64, err := strconv.ParseUint(args[0], 10, 16)
	if err != nil {
		return 0, 0, 0, time.Time{}, err
	}
	value, err := strconv.Atoi(args[1])
	if err != nil {
		return 0, 0, 0, time.Time{}, err
	}
	class, err := classOf(args[2])
	if err != nil {
		return 0, 0, 0, time.Time{}, err
	}
	at, err := timeOf(args[3])
	if err != nil {
		return 0, 0, 0, time.Time{}, err
	}
	return uint16(index64), value, class, at, nil
}

func classOf(s string) (outstation.EventClass, error) {
	switch s {
	case "1":
		return outstation.EventClass1, nil
	case "2":
		return outstation.EventClass2, nil
	case "3":
		return outstation.EventClass3, nil
	default:
		return 0, fmt.Errorf("class %s", s)
	}
}

func timeOf(s string) (time.Time, error) {
	if s == "none" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, s)
}
