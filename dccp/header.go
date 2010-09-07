// Copyright 2010 GoDCCP Authors. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package dccp

import "os"

type GenericHeader struct {
	SourcePort    uint16
	DestPort      uint16
	CCVal, CsCov  byte
	Type          int
	X             bool
	SeqNo         uint64
	AckNo         uint64
	ServiceCode   uint64
	Reset         []byte
	Options       []Option
	Data          []byte
}

type Option struct {
	Type      int
	Data      []byte
	Mandatory bool
}

var (
	ErrAlign         = os.NewError("align")
	ErrSize          = os.NewError("size")
	ErrSemantic      = os.NewError("semantic")
	ErrNumeric       = os.NewError("numeric")
	ErrOption        = os.NewError("option")
)

// Packet types. Stored in the Type field of the generic header.
// Receivers MUST ignore any packets with reserved type.  That is,
// packets with reserved type MUST NOT be processed, and they MUST
// NOT be acknowledged as received.
const (
	Request  = 0
	Response = 1
	Data     = 2
	Ack      = 3
	DataAck  = 4
	CloseReq = 5
	Close    = 6
	Reset    = 7
	Sync     = 8
	SyncAck  = 9
)

func isTypeReserved(Type int) bool {
	return Type >= 10 && Type <= 15
}

// Reset codes
const (
	ResetUnspecified       = 0
	ResetClosed            = 1
	ResetAborted           = 2
	ResetNoConnection      = 3
	ResetPacketError       = 4
	ResetOptionError       = 5
	ResetMandatoryError    = 6
	ResetConnectionRefused = 7
	ResetBadServiceCode    = 8
	ResetTooBusy           = 9
	ResetBadInitCookie     = 10
	ResetAgressionPenalty  = 11
)

func isResetCodeReserved(code int) bool {
	return code >= 12 && code <= 127
}

func isResetCodeCCIDSpecific(code int) bool {
	return code >= 128 && code <= 255
}

func areTypeAndXCompatible(Type int, X bool, AllowShortSeqNoFeature bool) bool {
	switch Type {
	case Request, Response:
		return X
	case Data, Ack, DataAck:
		if AllowShortSeqNoFeature {
			return true
		}
		return X
	case CloseReq, Close:
		return X
	case Reset:
		return X
	case Sync, SyncAck:
		return X
	}
	panic("unreach")
}

// Any DCCP header has a subset of the following subheaders, in this order:
// (1) Generic header
// (2) Acknowledgement Number Subheader
// (3) Code Subheader: Service Code, or Reset Code and Reset Data fields
// (4) Options and Padding
// (5) Application Data

// See RFC 4340, Page 21
func calcAckNoSubheaderSize(Type int, X bool) int {
	if Type == Request || Type == Data {
		return 0
	}
	if X {
		return 8
	}
	return 4
}

func calcCodeSubheaderSize(Type int) int {
	switch Type {
	case Request, Response:
		return 4
	case Data, Ack, DataAck:
		return 0
	case CloseReq, Close:
		return 0
	case Reset:
		return 4
	case Sync, SyncAck:
		return 0
	}
	panic("unreach")
}

// calcFixedHeaderSize() returns the size of the fixed portion of
// the generic header in bytes, based on its @Type and @X. This
// includes (1), (2) and (3).
func calcFixedHeaderSize(Type int, X bool) int {
	var r int
	switch X {
	case false:
		r = 12
	case true:
		r = 16
	}
	r += calcAckNoSubheaderSize(Type, X)
	r += calcCodeSubheaderSize(Type)
	return r
}

func mayHaveAppData(Type int) bool {
	switch Type {
	case Request, Response:
		return true
	case Data:
		return true
	case Ack:
		return true // may have App Data (essentially for padding) but must be ignored
	case DataAck:
		return true
	case CloseReq, Close:
		return true // may have App Data (essentially for padding) but must be ignored
	case Reset:
		return true // used for UTF-8 encoded error text
	case Sync, SyncAck:
		return true // may have App Data (essentially for padding) but must be ignored
	}
	panic("unreach")
}

// Options

const (
	OptionPadding         = 0
	OptionMandatory       = 1
	OptionSlowReceiver    = 2
	OptionChangeL         = 32
	OptionConfirmL        = 33
	OptionChangeR         = 34
	OptionConfirmR        = 35
	OptionInitCookie      = 36
	OptionNDPCount        = 37
	OptionAckVectorNonce0 = 38
	OptionAckVectorNonce1 = 39
	OptionDataDropped     = 40
	OptionTimestamp       = 41
	OptionTimestampEcho   = 42
	OptionElapsedTime     = 43
	OptionDataChecksum    = 44
)

func isOptionReserved(optionType int) bool {
	return (optionType >= 3 && optionType <= 31) || 
		(optionType >= 45 && optionType <= 127)
}

func isOptionCCIDSpecific(optionType int) bool {
	return optionType >= 128 && optionType <= 255
}

func isOptionSingleByte(optionType int) bool {
	return optionType >= 0 && optionType <= 31
}

func isOptionValidForType(optionType, Type int) bool {
	if Type != Data {
		return true
	}
	switch optionType {
	case OptionPadding,
		OptionSlowReceiver,
		OptionNDPCount,
		OptionTimestamp,
		OptionTimestampEcho,
		OptionDataChecksum:
		return true
	default:
		return false
	}
	panic("unreach")
}
