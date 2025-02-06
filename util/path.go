package util

import (
	"os/user"
	"path/filepath"
	"strings"
)

// HomePath enables support for ~/home/paths.
type HomePath struct {
	Path string
}

func NewHomePath(in string) (*HomePath, error) {
	h := &HomePath{}
	err := h.UnmarshalText([]byte(in))
	return h, err
}

func (h *HomePath) UnmarshalText(text []byte) error {
	h.Path = string(text)
	if strings.HasPrefix(h.Path, "~/") {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		h.Path = filepath.Join(usr.HomeDir, string(text[2:]))
		return nil
	}
	return nil
}
