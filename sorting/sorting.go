package sorting

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/bitrise-io/go-utils/pathutil"
	log "github.com/bitrise-io/steps-cocoapods-install/logger"
)

// PathDepth ...
func PathDepth(pth string) (int, error) {
	abs, err := pathutil.AbsPath(pth)
	if err != nil {
		return 0, err
	}
	comp := strings.Split(abs, string(os.PathSeparator))

	fixedComp := []string{}
	for _, c := range comp {
		if c != "" {
			fixedComp = append(fixedComp, c)
		}
	}

	return len(fixedComp), nil
}

// ByComponents ..
type ByComponents []string

func (s ByComponents) Len() int {
	return len(s)
}
func (s ByComponents) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByComponents) Less(i, j int) bool {
	path1 := s[i]
	path2 := s[j]

	d1, err := PathDepth(path1)
	if err != nil {
		log.Warn("failed to calculate path depth (%s), error: %s", path1, err)
		return false
	}

	d2, err := PathDepth(path2)
	if err != nil {
		log.Warn("failed to calculate path depth (%s), error: %s", path1, err)
		return false
	}

	if d1 < d2 {
		return true
	} else if d1 > d2 {
		return false
	}

	// if same component size,
	// do alphabetic sort based on the last component
	base1 := filepath.Base(path1)
	base2 := filepath.Base(path2)

	return base1 < base2
}
