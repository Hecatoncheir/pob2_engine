package discovery

import (
	"strings"
	"testing"

	"golang.org/x/sys/windows"
)

func TestWindowsProfileAndRecordAllowOnlyCurrentUser(t *testing.T) {
	p, err := Open(t.TempDir(), "private")
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Publish(Record{InstanceID: "test", Address: "127.0.0.1:1", Token: "secret"}); err != nil {
		t.Fatal(err)
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{p.Dir, p.RecordPath()} {
		sd, err := windows.GetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		acl, _, err := sd.DACL()
		if err != nil {
			t.Fatal(err)
		}
		if acl.AceCount != 1 || !strings.Contains(sd.String(), "(A;") || !strings.Contains(sd.String(), user.User.Sid.String()) {
			t.Fatalf("неверная DACL: %s", sd.String())
		}
	}
}

func TestWindowsReaderAllowsAtomicRecordReplacement(t *testing.T) {
	profile, err := Open(t.TempDir(), "replace")
	if err != nil {
		t.Fatal(err)
	}
	if err := profile.Publish(Record{InstanceID: "old", Address: "127.0.0.1:1", Token: "old"}); err != nil {
		t.Fatal(err)
	}
	reader, err := openRecord(profile.RecordPath())
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	if err := profile.Publish(Record{InstanceID: "new", Address: "127.0.0.1:2", Token: "new"}); err != nil {
		t.Fatal(err)
	}
	record, err := profile.Read()
	if err != nil || record.InstanceID != "new" {
		t.Fatal("запись не заменена при открытом читателе")
	}
}
