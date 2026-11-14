package al

// DNP3 Function Codes
// Reference: IEEE 1815-2012 Section 7.2
const (
	// Response codes (outstation to master)
	FuncConfirm             = 0x00 // Application layer confirmation
	FuncResponse            = 0x81 // Response with data
	FuncUnsolicitedResponse = 0x82 // Unsolicited response
	FuncNoAck               = 0x7F // No acknowledgment required

	// Master to outstation codes
	FuncRead                = 0x01 // Read data
	FuncWrite               = 0x02 // Write data
	FuncSelect              = 0x03 // Select for operate
	FuncOperate             = 0x04 // Execute selected operation
	FuncDirectOperate       = 0x05 // Direct operate (select + operate)
	FuncDirectOperateNoResp = 0x06 // Direct operate, no response
	FuncFreeze              = 10   // Freeze counter values
	FuncFileOpen            = 13   // Open file for transfer
	FuncFileClose           = 14   // Close file
	FuncFileRead            = 15   // Read file data
	FuncFileWrite           = 16   // Write file data
	// The legacy decimal constant 21 ("get identifier") was removed: IEEE 1815
	// assigns no request function to 21, and 21 (0x15) is DISABLE_UNSOLICITED.
	// Keeping both meanings on one wire value was ambiguous, so the misleading
	// symbol is gone rather than reassigned.
	FuncGetLabel          = 22 // Get device label
	FuncGetDescription    = 23 // Get device description
	FuncChangeFilename    = 24 // Change file name
	FuncStartUpload       = 25 // Start file upload
	FuncStartDownload     = 26 // Start file download
	FuncAuthenticate      = 27 // Authenticate request
	FuncAuthenticateConf  = 28 // Authentication confirmation
	FuncAbort             = 29 // Abort file transfer
	FuncTimeSync          = 32 // Time synchronization
	FuncRecordCurrentTime = 33 // Record current time
	FuncFreezeClear       = 37 // Freeze and clear counters
	FuncFreezeAtTime      = 38 // Freeze at specified time
	// Enable and disable unsolicited match IEEE 1815 and pinned OpenDNP3
	// 3.1.2 FunctionCode.h (0x14 / 0x15). The previous decimal values 41 and
	// 42 were not those function codes.
	FuncEnableUnsolicited     = 0x14 // Enable unsolicited responses
	FuncDisableUnsolicited    = 0x15 // Disable unsolicited responses
	FuncAssignClass           = 48   // Assign data to classes
	FuncDelayMeasurement      = 51   // Measure communication delay
	FuncRecordBatteryVoltage  = 52   // Record battery voltage
	FuncStartRestart          = 53   // Initiate device restart
	FuncInitializeApplication = 54   // Initialize application
	FuncStartSynchronization  = 57   // Start time synchronization
	FuncStopSynchronization   = 58   // Stop time synchronization
	FuncClockSyncBroadcast    = 59   // Clock sync via broadcast
)

// FunctionCodeName returns the name of a function code.
func FunctionCodeName(code uint8) string {
	names := map[uint8]string{
		0x00: "CONFIRM",
		0x01: "READ",
		0x02: "WRITE",
		0x03: "SELECT",
		0x04: "OPERATE",
		0x05: "DIRECT_OPERATE",
		0x06: "DIRECT_OPERATE_NO_RESPONSE",
		10:   "FREEZE",
		13:   "FILE_OPEN",
		14:   "FILE_CLOSE",
		15:   "FILE_READ",
		16:   "FILE_WRITE",
		22:   "GET_LABEL",
		23:   "GET_DESCRIPTION",
		24:   "CHANGE_FILENAME",
		25:   "START_UPLOAD",
		26:   "START_DOWNLOAD",
		27:   "AUTHENTICATE",
		28:   "AUTHENTICATE_CONFIRM",
		29:   "ABORT",
		32:   "TIME_SYNC",
		33:   "RECORD_CURRENT_TIME",
		37:   "FREEZE_CLEAR",
		38:   "FREEZE_AT_TIME",
		0x14: "ENABLE_UNSOLICITED",
		0x15: "DISABLE_UNSOLICITED",
		48:   "ASSIGN_CLASS",
		51:   "DELAY_MEASUREMENT",
		52:   "RECORD_BATTERY_VOLTAGE",
		53:   "START_RESTART",
		54:   "INITIALIZE_APPLICATION",
		57:   "START_SYNCHRONIZATION",
		58:   "STOP_SYNCHRONIZATION",
		59:   "CLOCK_SYNC_BROADCAST",
		0x7F: "NO_ACK",
		0x81: "RESPONSE",
		0x82: "UNSOLICITED_RESPONSE",
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
	if code == FuncConfirm || code == FuncResponse || code == FuncUnsolicitedResponse || code == FuncNoAck {
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
