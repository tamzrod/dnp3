package al

// DNP3 Function Codes
// Reference: IEEE 1815-2012 Section 7.2
const (
	// Response codes (outstation to master)
	FuncConfirm               = 0  // Deprecated, use separate confirmation
	FuncResponse              = 0  // Response with data
	FuncUnsolicitedResponse   = 1  // Unsolicited response
	FuncNoAck                 = 127 // No acknowledgment required

	// Master to outstation codes
	FuncRead                  = 2  // Read data
	FuncWrite                 = 3  // Write data
	FuncSelect                = 4  // Select for operate
	FuncOperate               = 5  // Execute selected operation
	FuncDirectOperate         = 6  // Direct operate (select + operate)
	FuncDirectOperateNoResp   = 7  // Direct operate, no response
	FuncFreeze                = 10 // Freeze counter values
	FuncFileOpen              = 13 // Open file for transfer
	FuncFileClose             = 14 // Close file
	FuncFileRead              = 15 // Read file data
	FuncFileWrite             = 16 // Write file data
	FuncGetIdentifier         = 21 // Get device identification
	FuncGetLabel              = 22 // Get device label
	FuncGetDescription        = 23 // Get device description
	FuncChangeFilename        = 24 // Change file name
	FuncStartUpload           = 25 // Start file upload
	FuncStartDownload         = 26 // Start file download
	FuncAuthenticate          = 27 // Authenticate request
	FuncAuthenticateConf      = 28 // Authentication confirmation
	FuncAbort                 = 29 // Abort file transfer
	FuncTimeSync              = 32 // Time synchronization
	FuncRecordCurrentTime     = 33 // Record current time
	FuncFreezeClear           = 37 // Freeze and clear counters
	FuncFreezeAtTime          = 38 // Freeze at specified time
	FuncEnableUnsolicited     = 41 // Enable unsolicited responses
	FuncDisableUnsolicited    = 42 // Disable unsolicited responses
	FuncAssignClass           = 48 // Assign data to classes
	FuncDelayMeasurement      = 51 // Measure communication delay
	FuncRecordBatteryVoltage  = 52 // Record battery voltage
	FuncStartRestart          = 53 // Initiate device restart
	FuncInitializeApplication = 54 // Initialize application
	FuncStartSynchronization  = 57 // Start time synchronization
	FuncStopSynchronization   = 58 // Stop time synchronization
	FuncClockSyncBroadcast    = 59 // Clock sync via broadcast
)

// FunctionCodeName returns the name of a function code.
func FunctionCodeName(code uint8) string {
	names := map[uint8]string{
		0:   "RESPONSE",
		1:   "UNSOLICITED_RESPONSE",
		2:   "READ",
		3:   "WRITE",
		4:   "SELECT",
		5:   "OPERATE",
		6:   "DIRECT_OPERATE",
		7:   "DIRECT_OPERATE_NO_RESPONSE",
		10:  "FREEZE",
		13:  "FILE_OPEN",
		14:  "FILE_CLOSE",
		15:  "FILE_READ",
		16:  "FILE_WRITE",
		21:  "GET_IDENTIFIER",
		22:  "GET_LABEL",
		23:  "GET_DESCRIPTION",
		24:  "CHANGE_FILENAME",
		25:  "START_UPLOAD",
		26:  "START_DOWNLOAD",
		27:  "AUTHENTICATE",
		28:  "AUTHENTICATE_CONFIRM",
		29:  "ABORT",
		32:  "TIME_SYNC",
		33:  "RECORD_CURRENT_TIME",
		37:  "FREEZE_CLEAR",
		38:  "FREEZE_AT_TIME",
		41:  "ENABLE_UNSOLICITED",
		42:  "DISABLE_UNSOLICITED",
		48:  "ASSIGN_CLASS",
		51:  "DELAY_MEASUREMENT",
		52:  "RECORD_BATTERY_VOLTAGE",
		53:  "START_RESTART",
		54:  "INITIALIZE_APPLICATION",
		57:  "START_SYNCHRONIZATION",
		58:  "STOP_SYNCHRONIZATION",
		59:  "CLOCK_SYNC_BROADCAST",
		127: "NO_ACK",
	}

	if name, ok := names[code]; ok {
		return name
	}
	
	if code >= 64 && code <= 127 {
		return "MANUFACTURER_SPECIFIC"
	}
	
	return "UNKNOWN"
}

// IsValidFunctionCode checks if a function code is valid.
func IsValidFunctionCode(code uint8) bool {
	// Response codes
	if code == FuncResponse || code == FuncUnsolicitedResponse || code == FuncNoAck {
		return true
	}
	
	// Master to outstation codes
	if code >= FuncCodeMinMaster && code <= FuncCodeMaxMaster {
		return true
	}
	
	// Manufacturer-specific codes
	if code >= FuncCodeMinMfr && code <= FuncCodeMaxMfr {
		return true
	}
	
	return false
}

// IsConfirmationFunction returns true for CONFIRM function codes.
func IsConfirmationFunction(code uint8) bool {
	return code == FuncConfirm
}

// IsReadFunction returns true for READ function codes.
func IsReadFunction(code uint8) bool {
	return code == FuncRead
}

// IsWriteFunction returns true for WRITE function codes.
func IsWriteFunction(code uint8) bool {
	return code == FuncWrite
}

// IsControlFunction returns true for control operation codes.
func IsControlFunction(code uint8) bool {
	return code == FuncSelect ||
		code == FuncOperate ||
		code == FuncDirectOperate ||
		code == FuncDirectOperateNoResp
}

// IsTimeFunction returns true for time-related function codes.
func IsTimeFunction(code uint8) bool {
	return code == FuncTimeSync ||
		code == FuncRecordCurrentTime ||
		code == FuncStartSynchronization ||
		code == FuncStopSynchronization ||
		code == FuncClockSyncBroadcast
}

// IsFileFunction returns true for file transfer function codes.
func IsFileFunction(code uint8) bool {
	return code == FuncFileOpen ||
		code == FuncFileClose ||
		code == FuncFileRead ||
		code == FuncFileWrite ||
		code == FuncChangeFilename ||
		code == FuncStartUpload ||
		code == FuncStartDownload ||
		code == FuncAbort
}

// FunctionCodeCategory describes the category of a function code.
type FunctionCodeCategory int

const (
	CategoryResponse FunctionCodeCategory = iota
	CategoryRead
	CategoryWrite
	CategoryControl
	CategoryTime
	CategoryFile
	CategoryConfiguration
	CategoryOther
)

// Category returns the category of a function code.
func Category(code uint8) FunctionCodeCategory {
	if code == FuncResponse || code == FuncUnsolicitedResponse {
		return CategoryResponse
	}
	if IsReadFunction(code) {
		return CategoryRead
	}
	if IsWriteFunction(code) {
		return CategoryWrite
	}
	if IsControlFunction(code) {
		return CategoryControl
	}
	if IsTimeFunction(code) {
		return CategoryTime
	}
	if IsFileFunction(code) {
		return CategoryFile
	}
	if code == FuncEnableUnsolicited || 
	   code == FuncDisableUnsolicited ||
	   code == FuncAssignClass ||
	   code == FuncGetIdentifier ||
	   code == FuncGetLabel ||
	   code == FuncGetDescription {
		return CategoryConfiguration
	}
	return CategoryOther
}

// CategoryName returns the name of a function code category.
func CategoryName(cat FunctionCodeCategory) string {
	names := []string{
		"Response",
		"Read",
		"Write",
		"Control",
		"Time",
		"File",
		"Configuration",
		"Other",
	}
	if cat >= 0 && int(cat) < len(names) {
		return names[cat]
	}
	return "Unknown"
}
