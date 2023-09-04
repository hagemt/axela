package brains

import "io"

func CloseAllQuietly(cs ...io.Closer) {
	for _, c := range cs {
		if c != nil {
			_ = c.Close()
		}
	}
}
