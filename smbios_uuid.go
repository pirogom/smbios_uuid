package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/digitalocean/go-smbios/smbios"
	"github.com/google/uuid"
)

/***	toBigEndian
		from https://github.com/siderolabs/go-smbios
**/
func toBigEndian(formatted []byte) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, 16))

	if err := binary.Write(buf, binary.BigEndian, formatted[4:20]); err != nil {
		return nil, err
	}

	b = buf.Bytes()

	return b, nil
}

/**
*	toMiddleEndian
    from https://github.com/siderolabs/go-smbios
**/
func toMiddleEndian(formatted []byte) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, 16))

	timeLow := binary.BigEndian.Uint32(formatted[4:8])
	if err := binary.Write(buf, binary.LittleEndian, timeLow); err != nil {
		return nil, err
	}

	timeMid := binary.BigEndian.Uint16(formatted[8:10])
	if err := binary.Write(buf, binary.LittleEndian, timeMid); err != nil {
		return nil, err
	}

	timeHigh := binary.BigEndian.Uint16(formatted[10:12])
	if err := binary.Write(buf, binary.LittleEndian, timeHigh); err != nil {
		return nil, err
	}

	clockSeqHi := formatted[12:13][0]
	if err := binary.Write(buf, binary.BigEndian, clockSeqHi); err != nil {
		return nil, err
	}

	clockSeqLow := formatted[13:14][0]
	if err := binary.Write(buf, binary.BigEndian, clockSeqLow); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, binary.BigEndian, formatted[14:20]); err != nil {
		return nil, err
	}

	b = buf.Bytes()

	return b, nil
}

/**
*	getSMBiosUUID
    from https://github.com/siderolabs/go-smbios
**/
func getSMBiosUUID(majorVer int, minorVer int, fm []byte) (uid uuid.UUID, err error) {
	var b []byte
	if majorVer >= 3 || (majorVer == 2 && minorVer >= 6) {
		b, err = toMiddleEndian(fm)
		if err != nil {
			return uid, fmt.Errorf("failed to convert to middle endian: %w", err)
		}
	} else {
		b, err = toBigEndian(fm)
		if err != nil {
			return uid, fmt.Errorf("failed to convert to big endian: %w", err)
		}
	}

	uid, err = uuid.FromBytes(b)
	if err != nil {
		return uid, fmt.Errorf("invalid GetUUID: %w", err)
	}
	return uid, nil
}

/**
*	GetSystemUUID
**/
func GetSystemUUID() (string, error) {
	// Find SMBIOS data in operating system-specific location.
	rc, ep, err := smbios.Stream()

	if err != nil {
		return "", err
	}
	defer rc.Close()

	// Decode SMBIOS structures from the stream.
	d := smbios.NewDecoder(rc)
	ss, err := d.Decode()
	if err != nil {
		return "", err
	}

	// Determine SMBIOS version and table location from entry point.
	major, minor, _ := ep.Version()

	for _, s := range ss {
		if s.Header.Type == 1 { // System Information
			srcBuf := s.Formatted[0:16]
			uid, uidErr := getSMBiosUUID(major, minor, srcBuf)
			if uidErr != nil {
				return "", uidErr
			}
			return strings.ToUpper(uid.String()), nil
		}
	}
	return "", errors.New("SMBIOS System Infomation's UUID not found")
}
