package main

// stripBOM removes a leading UTF-8 byte order mark (EF BB BF), if present.
// It is not required for valid UTF-8 and can appear on any file saved by
// editors like Visual Studio, so every file we merge needs it stripped
// before we scan its lines — otherwise it survives as part of the first
// line's content when that file isn't first in the merged output.
func stripBOM(data []byte) []byte {
	const bom = "\xef\xbb\xbf"
	if len(data) >= 3 && string(data[:3]) == bom {
		return data[3:]
	}
	return data
}
