package git

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

// A ParseError is a parse error.
type ParseError string

// An OrdinaryStatus is a status of a modified file.
type OrdinaryStatus struct {
	X    byte
	Y    byte
	Sub  string
	MH   int
	MI   int
	MW   int
	HH   string
	HI   string
	Path string
}

// A RenamedOrCopiedStatus is a status of a renamed or copied file.
type RenamedOrCopiedStatus struct {
	X        byte
	Y        byte
	Sub      string
	MH       int
	MI       int
	MW       int
	HH       string
	HI       string
	RC       byte
	Score    int
	Path     string
	OrigPath string
}

// An UnmergedStatus is the status of an unmerged file.
type UnmergedStatus struct {
	X    byte
	Y    byte
	Sub  string
	M1   int
	M2   int
	M3   int
	MW   int
	H1   string
	H2   string
	H3   string
	Path string
}

// An UntrackedStatus is a status of an untracked file.
type UntrackedStatus struct {
	Path string
}

// An IgnoredStatus is a status of an ignored file.
type IgnoredStatus struct {
	Path string
}

// A Status is a status.
type Status struct {
	Ordinary        []OrdinaryStatus
	RenamedOrCopied []RenamedOrCopiedStatus
	Unmerged        []UnmergedStatus
	Untracked       []UntrackedStatus
	Ignored         []IgnoredStatus
}

var (
	statusPorcelainV2ZOrdinaryRegexp = regexp.MustCompile(`` +
		`^1 ` +
		`([!\.\?ACDMRU])([!\.\?ACDMRU]) ` +
		`(N\.\.\.|S[\.C][\.M][\.U]) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-9a-f]+) ` +
		`([0-9a-f]+) ` +
		`(.*)` +
		`$`,
	)
	statusPorcelainV2ZRenamedOrCopiedRegexp = regexp.MustCompile(`` +
		`^2 ` +
		`([!\.\?ACDMRU])([!\.\?ACDMRU]) ` +
		`(N\.\.\.|S[\.C][\.M][\.U]) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-9a-f]+) ` +
		`([0-9a-f]+) ` +
		`([CR])([0-9]+) ` +
		`(.*?)\t(.*)` +
		`$`,
	)
	statusPorcelainV2ZUnmergedRegexp = regexp.MustCompile(`` +
		`^u ` +
		`([!\.\?ACDMRU])([!\.\?ACDMRU]) ` +
		`(N\.\.\.|S[\.C][\.M][\.U]) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-7]+) ` +
		`([0-9a-f]+) ` +
		`([0-9a-f]+) ` +
		`([0-9a-f]+) ` +
		`(.*)` +
		`$`,
	)
	statusPorcelainV2ZUntrackedRegexp = regexp.MustCompile(`` +
		`^\? ` +
		`(.*)` +
		`$`,
	)
	statusPorcelainV2ZIgnoredRegexp = regexp.MustCompile(`` +
		`^! ` +
		`(.*)` +
		`$`,
	)
)

func (e ParseError) Error() string {
	return fmt.Sprintf("cannot parse %q", string(e))
}

// ParseStatusPorcelainV2 parses the output of
//   git status --ignored --porcelain=v2
// See https://git-scm.com/docs/git-status.
func ParseStatusPorcelainV2(output []byte) (*Status, error) {
	status := &Status{}
	s := bufio.NewScanner(bytes.NewReader(output))
	for s.Scan() {
		text := s.Text()
		switch text[0] {
		case '1':
			m := statusPorcelainV2ZOrdinaryRegexp.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			var (
				mH, _ = strconv.ParseInt(text[m[8]:m[9]], 8, 64)
				mI, _ = strconv.ParseInt(text[m[10]:m[11]], 8, 64)
				mW, _ = strconv.ParseInt(text[m[12]:m[13]], 8, 64)
			)
			os := OrdinaryStatus{
				X:    text[m[2]],
				Y:    text[m[4]],
				Sub:  text[m[6]:m[7]],
				MH:   int(mH),
				MI:   int(mI),
				MW:   int(mW),
				HH:   text[m[14]:m[15]],
				HI:   text[m[16]:m[17]],
				Path: text[m[18]:m[19]],
			}
			status.Ordinary = append(status.Ordinary, os)
		case '2':
			m := statusPorcelainV2ZRenamedOrCopiedRegexp.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			var (
				mH, _    = strconv.ParseInt(text[m[8]:m[9]], 8, 64)
				mI, _    = strconv.ParseInt(text[m[10]:m[11]], 8, 64)
				mW, _    = strconv.ParseInt(text[m[12]:m[13]], 8, 64)
				score, _ = strconv.ParseInt(text[m[20]:m[21]], 10, 64)
			)
			rocs := RenamedOrCopiedStatus{
				X:        text[m[2]],
				Y:        text[m[4]],
				Sub:      text[m[6]:m[7]],
				MH:       int(mH),
				MI:       int(mI),
				MW:       int(mW),
				HH:       text[m[14]:m[15]],
				HI:       text[m[16]:m[17]],
				RC:       text[m[18]],
				Score:    int(score),
				Path:     text[m[22]:m[23]],
				OrigPath: text[m[24]:m[25]],
			}
			status.RenamedOrCopied = append(status.RenamedOrCopied, rocs)
		case 'u':
			m := statusPorcelainV2ZUnmergedRegexp.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			var (
				m1, _ = strconv.ParseInt(text[m[6]:m[7]], 8, 64)
				m2, _ = strconv.ParseInt(text[m[8]:m[9]], 8, 64)
				m3, _ = strconv.ParseInt(text[m[10]:m[11]], 8, 64)
				mW, _ = strconv.ParseInt(text[m[12]:m[13]], 8, 64)
			)
			us := UnmergedStatus{
				X:   text[m[2]],
				Y:   text[m[4]],
				Sub: text[m[6]:m[7]],
				M1:  int(m1),
				M2:  int(m2),
				M3:  int(m3),
				MW:  int(mW),
			}
			status.Unmerged = append(status.Unmerged, us)
		case '?':
			m := statusPorcelainV2ZUntrackedRegexp.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			us := UntrackedStatus{
				Path: text[m[2]:m[3]],
			}
			status.Untracked = append(status.Untracked, us)
		case '!':
			m := statusPorcelainV2ZIgnoredRegexp.FindStringSubmatchIndex(text)
			if m == nil {
				return nil, ParseError(text)
			}
			us := IgnoredStatus{
				Path: text[m[2]:m[3]],
			}
			status.Ignored = append(status.Ignored, us)
		case '#':
			continue
		default:
			return nil, ParseError(text)
		}
	}
	return status, s.Err()
}
