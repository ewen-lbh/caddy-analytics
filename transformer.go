package caddy_analytics

import (
	"bytes"

	"golang.org/x/text/transform"
)

// Just like github.com/icholy/replace, but only replaces the first occurrence of the old string with the new string.

// firstOccurenceReplacer replaces text in a stream
// See: http://golang.org/x/text/transform
type firstOccurenceReplacer struct {
	transform.NopResetter

	old, new []byte
	oldlen   int
}

var _ transform.Transformer = (*firstOccurenceReplacer)(nil)

// firstStringReplacer returns a transformer that replaces all instances of old with new.
// Unlike strings.Replace, empty old values don't match anything.
func firstStringReplacer(old, new string) firstOccurenceReplacer {
	return firstOccurenceReplacer{old: []byte(old), new: []byte(new), oldlen: len([]byte(old))}
}

// Transform implements golang.org/x/text/transform#Transformer
func (t firstOccurenceReplacer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	var n int
	// don't do anything for empty old string. We're forced to do this because an optimization in
	// transform.String prevents us from generating any output when the src is empty.
	// see: https://github.com/golang/text/blob/master/transform/transform.go#L570-L576
	if t.oldlen == 0 {
		n, err = fullcopy(dst, src)
		return n, n, err
	}
	// replace first instance of old with new
	i := bytes.Index(src[nSrc:], t.old)
	if i != -1 {
		// copy everything up to the match
		n, err = fullcopy(dst[nDst:], src[nSrc:nSrc+i])
		nSrc += n
		nDst += n
		if err != nil {
			return
		}
		// copy the new value
		n, err = fullcopy(dst[nDst:], t.new)
		if err != nil {
			return
		}
		nDst += n
		nSrc += t.oldlen
	}
	// if we're at the end, tack on any remaining bytes
	if atEOF {
		n, err = fullcopy(dst[nDst:], src[nSrc:])
		nDst += n
		nSrc += n
		return
	}
	// skip everything except the trailing len(r.old) - 1
	// we do this becasue there could be a match straddling
	// the boundary
	if skip := len(src[nSrc:]) - t.oldlen + 1; skip > 0 {
		n, err = fullcopy(dst[nDst:], src[nSrc:nSrc+skip])
		nSrc += n
		nDst += n
		if err != nil {
			return
		}
	}
	err = transform.ErrShortSrc
	return
}

func fullcopy(dst, src []byte) (n int, err error) {
	n = copy(dst, src)
	if n < len(src) {
		err = transform.ErrShortDst
	}
	return
}
