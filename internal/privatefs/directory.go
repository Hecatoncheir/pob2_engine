package privatefs

import (
	"fmt"
	"os"
	"regexp"
)

var profileName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]{0,63}$`)

func ValidateProfile(name string) error {
	if !profileName.MatchString(name) {
		return fmt.Errorf("имя профиля: 1–64 буквы ASCII, цифры, дефис или подчёркивание")
	}
	return nil
}

// EnsureDir защищает собственный каталог приложения, не меняя его родителя.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0700); err != nil {
		return err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("каталог профиля не должен быть ссылкой: %s", path)
	}
	return protectDir(path)
}
