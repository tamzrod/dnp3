---
title: "330 - Glossary"
owner: glossary
---

# DNP3 Protocol Glossary

## A

### ACK (Acknowledgment)
Data link layer response indicating successful frame receipt.

### APDU (Application Protocol Data Unit)
Complete application layer message including function code and objects.

### Application Layer
Protocol layer handling user data, function codes, and operations.

### Assignment
Configuration of data point class assignment.

## B

### Balanced Mode
Data link mode where either device can initiate communication.

### Binary Input
Two-state (ON/OFF) measurement point.

### Binary Output
Controllable two-state output point.

### Broadcast
Message sent to all devices (address 0xFFFF).

### Buffer Overflow
Condition when event buffer capacity is exceeded.

## C

### CFCB (Critical Frame Count Bit)
See FCB (Data Link confirmation mechanism).

### Challenge
Random value sent by outstation for authentication verification.

### Class
Priority grouping for data (0-3) used in class-based polling.

### CON (Confirmation)
Application layer flag requiring explicit acknowledgment.

### CRC (Cyclic Redundancy Check)
Error detection code for data link frames.

### CTE (Change Time Estimate)
See CTO Cohort.

### CTO Cohort
Time value with unsynchronization indicator.

## D

### DNP3 (Distributed Network Protocol 3)
SCADA protocol for utility/industrial communication.

### Database
Logical organization of all data points in an outstation.

### Deadband
Threshold for analog event generation.

### Direct Operate
Control operation combining select and operate in one request.

### DIR (Direction)
Data link bit indicating message direction (master/outstation).

### Double-Bit Binary
Two-bit binary status with transitional state indication.

## E

### Event
Record of data change including value, flags, and timestamp.

### Event Buffer
Storage for generated events pending transmission.

## F

### FCB (Frame Count Bit)
Data link confirmation mechanism bit that toggles on each confirmed frame.

### FCV (Frame Count Bit Valid)
Data link bit indicating FCB is meaningful.

### FIR (First Fragment)
Transport layer bit indicating first fragment of multi-fragment message.

### FIN (Final Fragment)
Transport layer bit indicating final fragment of multi-fragment message.

### Fragment
Unit of transport layer containing portion of application message.

### Frame
Unit of data link layer transmission including header, data, and CRC.

### Freeze
Operation to capture counter value at a point in time.

## G

### Group
Category of data type (e.g., Group 1 = Binary Input).

## I

### IIN (Internal Indication)
Status flags in every response indicating device condition.

### Index
Point number within a group.

### Integrity Poll
Poll of all Class 0 (static) data.

## M

### MAC (Message Authentication Code)
Cryptographic code for message authentication.

### Master
Central control station that initiates communication.

## N

### NACK (Negative Acknowledgment)
Data link response indicating rejection or busy condition.

## O

### Operate
Execute previously selected control operation.

### Outstation
Remote device (RTU/IED) that responds to master requests.

## P

### PDU (Protocol Data Unit)
Complete protocol message at any layer.

### Polled Mode
Communication pattern where master requests data.

### Primary
Device that initiates data link communication.

### PRM (Primary)
Data link bit indicating primary station.

## Q

### Qualifier
Object addressing specification indicating range format.

## R

### Response
Outstation reply to master request.

### Rollover
Counter reaching maximum value and wrapping to zero.

### RTU (Remote Terminal Unit)
Type of outstation device.

## S

### SBO (Select-Before-Operate)
Control pattern requiring SELECT before OPERATE.

### SCADA (Supervisory Control and Data Acquisition)
System architecture for remote monitoring and control.

### Secondary
Device that responds to data link primary.

### Sequence Number
Number tracking message order at various layers.

### Session
Communication session between master and outstation.

### Static Data
Current point values without events.

## T

### Transport Layer
Protocol layer handling message fragmentation and reassembly.

## U

### UNS (Unsolicited)
Application layer bit indicating unprompted response.

### Unsolicited Response
Response sent by outstation without prior master request.

### User Data
Data link frame data field containing transport/application data.

## V

### Variation
Specific encoding format within an object group.

## References

- IEEE 1815-2012 Section 3: Definitions and Acronyms
- DNP3 Users Group Technical Guidelines
