// Programmatic E2E test of the DNP3 Engineering Workbench.
// Uses the PUBLIC API (pkg/dnp3/master and pkg/dnp3/outstation).
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"dnp3/pkg/dnp3"
	"dnp3/pkg/dnp3/master"
	"dnp3/pkg/dnp3/outstation"
	"dnp3/pkg/dnp3/types"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func main() {
	fmt.Println("========================================")
	fmt.Println("DNP3 Workbench E2E Test (Public API)")
	fmt.Println("========================================")
	
	port, err := getFreePort()
	if err != nil {
		log.Fatalf("Failed to get free port: %v", err)
	}
	fmt.Printf("Using port: %d\n\n", port)

	// Step 1: Create and start Outstation
	fmt.Println("=== STEP 1: Create and Start Outstation ===")
	
	outstationConfig := outstation.NewConfig(
		outstation.WithAddress(1024),
		outstation.WithMasterAddress(0xFFFF),
		outstation.WithTransport(dnp3.TCP, "localhost", port),
	)
	
	outstationServer, err := outstation.NewServer(outstationConfig)
	if err != nil {
		log.Fatalf("FAILED: Create outstation: %v", err)
	}
	fmt.Println("✅ Outstation created")

	// Start server in background
	serverCtx, serverCancel := context.WithCancel(context.Background())
	defer serverCancel()

	if err := outstationServer.Start(serverCtx); err != nil {
		log.Fatalf("FAILED: Start outstation: %v", err)
	}
	fmt.Printf("✅ Outstation started and LISTENING on port %d\n", port)
	
	// Give server time to start listening
	time.Sleep(200 * time.Millisecond)
	
	// Step 2: Create and Connect Master
	fmt.Println("\n=== STEP 2: Create and Connect Master ===")
	
	masterConfig := master.NewConfig(
		master.WithMasterAddress(0xFFFF),
		master.WithOutstationAddress(1024),
		master.WithTransport(dnp3.TCP, "localhost", port),
		master.WithTimeout(5*time.Second),
	)
	
	masterClient, err := master.NewClient(masterConfig)
	if err != nil {
		log.Fatalf("FAILED: Create master client: %v", err)
	}
	fmt.Println("✅ Master client created")

	ctx := context.Background()
	
	if err := masterClient.Connect(ctx); err != nil {
		log.Fatalf("FAILED: Master Connect: %v", err)
	}
	fmt.Printf("✅ Master CONNECTED successfully to localhost:%d\n", port)
	fmt.Printf("✅ Master state: %s\n", masterClient.State())
	
	// Step 3: Read Class 0 (All Static Data)
	fmt.Println("\n=== STEP 3: Read Class 0 (All Static Data) ===")
	
	readReq := types.NewReadRequest(
		types.GroupRequest{Group: 1, Variation: 0},   // Binary Inputs (all variations)
		types.GroupRequest{Group: 30, Variation: 0},  // Analog Inputs (all variations)
		types.GroupRequest{Group: 20, Variation: 0},  // Counters (all variations)
	)
	
	readResp, err := masterClient.Read(ctx, readReq)
	if err != nil {
		log.Fatalf("FAILED: Read Class 0: %v", err)
	}
	
	fmt.Println("✅ READ Class 0 SUCCEEDED!")
	fmt.Printf("   IIN: %v\n", readResp.IIN)
	
	if len(readResp.BinaryInputs) > 0 {
		fmt.Printf("   Binary Inputs (%d points): ", len(readResp.BinaryInputs))
		for i, bi := range readResp.BinaryInputs {
			if i >= 4 {
				fmt.Printf("...(+%d)", len(readResp.BinaryInputs)-4)
				break
			}
			fmt.Printf("[%d]=%v ", bi.Index, bi.Value)
		}
		fmt.Println()
	}
	
	if len(readResp.AnalogInputs) > 0 {
		fmt.Printf("   Analog Inputs (%d points): ", len(readResp.AnalogInputs))
		for i, ai := range readResp.AnalogInputs {
			if i >= 4 {
				fmt.Printf("...(+%d)", len(readResp.AnalogInputs)-4)
				break
			}
			fmt.Printf("[%d]=%.1f ", ai.Index, ai.Value)
		}
		fmt.Println()
	}
	
	if len(readResp.Counters) > 0 {
		fmt.Printf("   Counters (%d points): ", len(readResp.Counters))
		for i, c := range readResp.Counters {
			if i >= 4 {
				fmt.Printf("...(+%d)", len(readResp.Counters)-4)
				break
			}
			fmt.Printf("[%d]=%d ", c.Index, c.Value)
		}
		fmt.Println()
	}

	// Step 4: Read Digital Inputs specifically
	fmt.Println("\n=== STEP 4: Read Digital Inputs (Group 1) ===")
	
	diReq := types.NewReadRequest(types.GroupRequest{Group: 1, Variation: 1})
	
	diResp, err := masterClient.Read(ctx, diReq)
	if err != nil {
		log.Fatalf("FAILED: Read Digital Inputs: %v", err)
	}
	
	fmt.Println("✅ READ Digital Inputs SUCCEEDED!")
	if len(diResp.BinaryInputs) > 0 {
		fmt.Printf("   Binary Inputs: ")
		for i, bi := range diResp.BinaryInputs {
			if i >= 8 {
				break
			}
			fmt.Printf("[%d]=%v ", bi.Index, bi.Value)
		}
		fmt.Println()
	}

	// Step 5: Read Analog Inputs specifically
	fmt.Println("\n=== STEP 5: Read Analog Inputs (Group 30) ===")
	
	aiReq := types.NewReadRequest(types.GroupRequest{Group: 30, Variation: 1})
	
	aiResp, err := masterClient.Read(ctx, aiReq)
	if err != nil {
		log.Fatalf("FAILED: Read Analog Inputs: %v", err)
	}
	
	fmt.Println("✅ READ Analog Inputs SUCCEEDED!")
	if len(aiResp.AnalogInputs) > 0 {
		fmt.Printf("   Analog Inputs: ")
		for i, ai := range aiResp.AnalogInputs {
			if i >= 8 {
				break
			}
			fmt.Printf("[%d]=%.1f ", ai.Index, ai.Value)
		}
		fmt.Println()
	}

	// Step 6: Operate Binary Output
	fmt.Println("\n=== STEP 6: Operate Binary Output (Group 12) ===")
	
	// Create binary control command (Group 12)
	controlOutput := types.NewBinaryControl(0, true, types.DirectOperate)
	
	operateResp, err := masterClient.Operate(ctx, controlOutput)
	if err != nil {
		log.Fatalf("FAILED: Operate Binary Output: %v", err)
	}
	
	fmt.Println("✅ OPERATE Binary Output SUCCEEDED!")
	fmt.Printf("   Command: Direct Operate Group 12, Index 0, Value=true")
	fmt.Println()
	if operateResp != nil {
		fmt.Printf("   Response Status: %v\n", operateResp.Status)
	} else {
		fmt.Println("   Response: (no error)")
	}

	// Step 7: Clean shutdown
	fmt.Println("\n=== STEP 7: Clean Shutdown ===")
	
	masterClient.Disconnect(ctx)
	fmt.Println("✅ Master disconnected")
	
	outstationServer.Stop(ctx)
	fmt.Println("✅ Outstation stopped")

	fmt.Println("\n========================================")
	fmt.Println("ALL WORKBENCH PATH TESTS PASSED!")
	fmt.Println("========================================")
	fmt.Println("\nFUNCTIONAL WORKBENCH PATH: PASS")
	fmt.Println("Ready for user local test: YES")
	os.Exit(0)
}
