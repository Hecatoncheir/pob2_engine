package discovery

import (
	"errors"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Читатель не блокирует атомарную замену connection.json владельцем профиля.
func openRecord(path string) (*os.File, error) {
	name, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	handle, err := windows.CreateFile(name, windows.GENERIC_READ, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return nil, &os.PathError{Op: "open", Path: path, Err: err}
	}
	return os.NewFile(uintptr(handle), path), nil
}

type fileRenameInfo struct {
	Flags          uint32
	RootDirectory  windows.Handle
	FileNameLength uint32
	FileName       [1]uint16
}

func replaceRecord(source, target string) error {
	from, err := windows.UTF16PtrFromString(source)
	if err != nil {
		return err
	}
	name, err := windows.UTF16FromString(target)
	if err != nil {
		return err
	}
	handle, err := windows.CreateFile(from, windows.DELETE, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE|windows.FILE_SHARE_DELETE, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return &os.PathError{Op: "open", Path: source, Err: err}
	}
	var layout fileRenameInfo
	offset := int(unsafe.Offsetof(layout.FileName))
	buffer := make([]byte, offset+len(name)*2)
	info := (*fileRenameInfo)(unsafe.Pointer(&buffer[0]))
	info.Flags = windows.FILE_RENAME_REPLACE_IF_EXISTS | windows.FILE_RENAME_POSIX_SEMANTICS
	info.FileNameLength = uint32((len(name) - 1) * 2)
	copy(unsafe.Slice((*uint16)(unsafe.Pointer(&buffer[offset])), len(name)), name)
	// FileRenameInfoEx сохраняет открытые дескрипторы прежнего файла:
	// https://learn.microsoft.com/windows/win32/api/winbase/ns-winbase-file_rename_info
	err = func() error {
		defer windows.CloseHandle(handle)
		return windows.SetFileInformationByHandle(handle, windows.FileRenameInfoEx, &buffer[0], uint32(len(buffer)))
	}()
	if errors.Is(err, windows.ERROR_INVALID_PARAMETER) || errors.Is(err, windows.ERROR_NOT_SUPPORTED) {
		return os.Rename(source, target)
	}
	if err != nil {
		return &os.LinkError{Op: "rename", Old: source, New: target, Err: err}
	}
	return nil
}

func lockFile(file *os.File) error {
	err := windows.LockFileEx(windows.Handle(file.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return ErrLocked
	}
	return err
}
