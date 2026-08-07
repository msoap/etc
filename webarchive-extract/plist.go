package main

// Minimal property-list reader: enough of the format to walk a .webarchive.
// Binary plists are parsed natively; XML plists are parsed with encoding/xml.
// If either fails we fall back to shelling out to plutil.

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
)

// loadPlist reads path and returns the root value as nested Go types:
// map[string]any, []any, string, []byte, int64, float64, bool.
func loadPlist(path string, forcePlutil bool) (any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("file is empty")
	}
	if !forcePlutil {
		if bytes.HasPrefix(raw, []byte("bplist0")) {
			v, err := parseBinaryPlist(raw)
			if err == nil {
				return v, nil
			}
			warnf("native binary-plist parse failed (%v), retrying with plutil", err)
		} else {
			v, err := parseXMLPlist(raw)
			if err == nil {
				return v, nil
			}
			warnf("XML plist parse failed (%v), retrying with plutil", err)
		}
	}
	return plutilPlist(raw)
}

// plutilPlist converts any plist flavour to XML via plutil, then parses it.
func plutilPlist(raw []byte) (any, error) {
	cmd := exec.Command("plutil", "-convert", "xml1", "-o", "-", "-")
	cmd.Stdin = bytes.NewReader(raw)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("plutil: %s", msg)
	}
	return parseXMLPlist(out)
}

// ---------------------------------------------------------------- XML plists

func parseXMLPlist(raw []byte) (any, error) {
	dec := xml.NewDecoder(bytes.NewReader(raw))
	dec.Strict = false
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("no <plist> element found")
		}
		se, ok := tok.(xml.StartElement)
		if !ok || se.Name.Local != "plist" {
			continue
		}
		for {
			tok, err := dec.Token()
			if err != nil {
				return nil, fmt.Errorf("<plist> has no root value")
			}
			switch t := tok.(type) {
			case xml.StartElement:
				return decodeXMLValue(dec, t)
			case xml.EndElement:
				return nil, fmt.Errorf("<plist> has no root value")
			}
		}
	}
}

func decodeXMLValue(dec *xml.Decoder, start xml.StartElement) (any, error) {
	switch start.Name.Local {
	case "dict":
		m := map[string]any{}
		key, haveKey := "", false
		for {
			tok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "key" {
					var s string
					if err := dec.DecodeElement(&s, &t); err != nil {
						return nil, err
					}
					key, haveKey = s, true
					continue
				}
				v, err := decodeXMLValue(dec, t)
				if err != nil {
					return nil, err
				}
				if haveKey {
					m[key] = v
					haveKey = false
				}
			case xml.EndElement:
				return m, nil
			}
		}
	case "array":
		var a []any
		for {
			tok, err := dec.Token()
			if err != nil {
				return nil, err
			}
			switch t := tok.(type) {
			case xml.StartElement:
				v, err := decodeXMLValue(dec, t)
				if err != nil {
					return nil, err
				}
				a = append(a, v)
			case xml.EndElement:
				return a, nil
			}
		}
	case "string", "date", "ustring":
		var s string
		if err := dec.DecodeElement(&s, &start); err != nil {
			return nil, err
		}
		return s, nil
	case "data":
		var s string
		if err := dec.DecodeElement(&s, &start); err != nil {
			return nil, err
		}
		b, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(s), ""))
		if err != nil {
			return nil, fmt.Errorf("bad base64 in <data>: %w", err)
		}
		return b, nil
	case "integer":
		var s string
		if err := dec.DecodeElement(&s, &start); err != nil {
			return nil, err
		}
		n, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
		if err != nil {
			return nil, err
		}
		return n, nil
	case "real":
		var s string
		if err := dec.DecodeElement(&s, &start); err != nil {
			return nil, err
		}
		f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return nil, err
		}
		return f, nil
	case "true":
		return true, dec.Skip()
	case "false":
		return false, dec.Skip()
	default:
		return nil, dec.Skip()
	}
}

// ------------------------------------------------------------- binary plists

type bplist struct {
	data    []byte
	offsets []uint64
	refSize int
}

func parseBinaryPlist(data []byte) (any, error) {
	if len(data) < 40 {
		return nil, fmt.Errorf("too short for a binary plist")
	}
	if !bytes.HasPrefix(data, []byte("bplist0")) {
		return nil, fmt.Errorf("missing bplist magic")
	}
	trailer := data[len(data)-32:]
	offSize, refSize := int(trailer[6]), int(trailer[7])
	numObjects := binary.BigEndian.Uint64(trailer[8:16])
	topObject := binary.BigEndian.Uint64(trailer[16:24])
	tableOff := binary.BigEndian.Uint64(trailer[24:32])

	if offSize < 1 || offSize > 8 || refSize < 1 || refSize > 8 {
		return nil, fmt.Errorf("bad trailer sizes (offset %d, ref %d)", offSize, refSize)
	}
	body := uint64(len(data) - 32)
	if numObjects == 0 || tableOff < 8 || tableOff > body ||
		numObjects > (body-tableOff)/uint64(offSize) {
		return nil, fmt.Errorf("offset table out of range")
	}
	p := &bplist{data: data, refSize: refSize, offsets: make([]uint64, numObjects)}
	for i := range p.offsets {
		start := tableOff + uint64(i*offSize)
		p.offsets[i] = beUint(data[start : start+uint64(offSize)])
	}
	return p.object(topObject, 0)
}

func beUint(b []byte) uint64 {
	var n uint64
	for _, c := range b {
		n = n<<8 | uint64(c)
	}
	return n
}

func (p *bplist) slice(off uint64, n uint64) ([]byte, error) {
	end := off + n
	if off > uint64(len(p.data)) || end > uint64(len(p.data)) || end < off {
		return nil, fmt.Errorf("object at %d runs past end of file", off)
	}
	return p.data[off:end], nil
}

// count reads the length of a variable-sized object, returning the length and
// the size of the marker+length header.
func (p *bplist) count(off uint64) (uint64, uint64, error) {
	marker := p.data[off]
	if marker&0x0F != 0x0F {
		return uint64(marker & 0x0F), 1, nil
	}
	next, err := p.slice(off+1, 1)
	if err != nil {
		return 0, 0, err
	}
	if next[0]&0xF0 != 0x10 {
		return 0, 0, fmt.Errorf("expected int length marker at %d", off+1)
	}
	size := uint64(1) << (next[0] & 0x0F)
	b, err := p.slice(off+2, size)
	if err != nil {
		return 0, 0, err
	}
	return beUint(b), 2 + size, nil
}

func (p *bplist) object(ref uint64, depth int) (any, error) {
	if depth > 64 {
		return nil, fmt.Errorf("plist nested more than 64 levels deep")
	}
	if ref >= uint64(len(p.offsets)) {
		return nil, fmt.Errorf("object reference %d out of range", ref)
	}
	off := p.offsets[ref]
	if off >= uint64(len(p.data)) {
		return nil, fmt.Errorf("object %d offset out of range", ref)
	}
	marker := p.data[off]

	switch marker & 0xF0 {
	case 0x00:
		switch marker {
		case 0x00, 0x0F:
			return nil, nil
		case 0x08:
			return false, nil
		case 0x09:
			return true, nil
		}
		return nil, fmt.Errorf("unknown marker 0x%02x", marker)

	case 0x10: // integer, 2^n bytes big-endian
		n := uint64(1) << (marker & 0x0F)
		b, err := p.slice(off+1, n)
		if err != nil {
			return nil, err
		}
		if n >= 8 { // 8- and 16-byte ints are signed; keep the low 64 bits
			return int64(binary.BigEndian.Uint64(b[n-8:])), nil
		}
		return int64(beUint(b)), nil

	case 0x20: // real
		n := uint64(1) << (marker & 0x0F)
		b, err := p.slice(off+1, n)
		if err != nil {
			return nil, err
		}
		switch n {
		case 4:
			return float64(math.Float32frombits(binary.BigEndian.Uint32(b))), nil
		case 8:
			return math.Float64frombits(binary.BigEndian.Uint64(b)), nil
		}
		return nil, fmt.Errorf("unsupported real width %d", n)

	case 0x30: // date, seconds since 2001-01-01
		b, err := p.slice(off+1, 8)
		if err != nil {
			return nil, err
		}
		secs := math.Float64frombits(binary.BigEndian.Uint64(b))
		epoch := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		return epoch.Add(time.Duration(secs * float64(time.Second))).Format(time.RFC3339), nil

	case 0x40: // data
		n, hdr, err := p.count(off)
		if err != nil {
			return nil, err
		}
		b, err := p.slice(off+hdr, n)
		if err != nil {
			return nil, err
		}
		return bytes.Clone(b), nil

	case 0x50: // ASCII string
		n, hdr, err := p.count(off)
		if err != nil {
			return nil, err
		}
		b, err := p.slice(off+hdr, n)
		if err != nil {
			return nil, err
		}
		return string(b), nil

	case 0x60: // UTF-16BE string
		n, hdr, err := p.count(off)
		if err != nil {
			return nil, err
		}
		b, err := p.slice(off+hdr, n*2)
		if err != nil {
			return nil, err
		}
		u := make([]uint16, n)
		for i := range u {
			u[i] = binary.BigEndian.Uint16(b[i*2:])
		}
		return string(utf16.Decode(u)), nil

	case 0x80: // UID
		n := uint64(marker&0x0F) + 1
		b, err := p.slice(off+1, n)
		if err != nil {
			return nil, err
		}
		return int64(beUint(b)), nil

	case 0xA0, 0xC0: // array, set
		n, hdr, err := p.count(off)
		if err != nil {
			return nil, err
		}
		refs, err := p.refs(off+hdr, n)
		if err != nil {
			return nil, err
		}
		out := make([]any, 0, n)
		for _, r := range refs {
			v, err := p.object(r, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil

	case 0xD0: // dictionary
		n, hdr, err := p.count(off)
		if err != nil {
			return nil, err
		}
		keyRefs, err := p.refs(off+hdr, n)
		if err != nil {
			return nil, err
		}
		valRefs, err := p.refs(off+hdr+n*uint64(p.refSize), n)
		if err != nil {
			return nil, err
		}
		out := make(map[string]any, n)
		for i := range keyRefs {
			k, err := p.object(keyRefs[i], depth+1)
			if err != nil {
				return nil, err
			}
			v, err := p.object(valRefs[i], depth+1)
			if err != nil {
				return nil, err
			}
			ks, ok := k.(string)
			if !ok {
				ks = fmt.Sprint(k)
			}
			out[ks] = v
		}
		return out, nil
	}
	return nil, fmt.Errorf("unknown marker 0x%02x at offset %d", marker, off)
}

func (p *bplist) refs(off, n uint64) ([]uint64, error) {
	b, err := p.slice(off, n*uint64(p.refSize))
	if err != nil {
		return nil, err
	}
	out := make([]uint64, n)
	for i := range out {
		out[i] = beUint(b[i*p.refSize : (i+1)*p.refSize])
	}
	return out, nil
}
