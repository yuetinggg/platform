// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"strings"
	"testing"
)

func TestPasswordHash(t *testing.T) {
	hash := HashPassword("Test")

	if !ComparePassword(hash, "Test") {
		t.Fatal("Passwords don't match")
	}

	if ComparePassword(hash, "Test2") {
		t.Fatal("Passwords should not have matched")
	}
}

func TestUserJson(t *testing.T) {
	user := User{Id: NewId(), Username: NewId()}
	json := user.ToJson()
	ruser := UserFromJson(strings.NewReader(json))

	if user.Id != ruser.Id {
		t.Fatal("Ids do not match")
	}
}

func TestUserPreSave(t *testing.T) {
	user := User{Password: "test"}
	user.PreSave()
	user.Etag(true, true)
}

func TestUserPreUpdate(t *testing.T) {
	user := User{Password: "test"}
	user.PreUpdate()

	user.ThemeProps = StringMap{
		"codeTheme":     "github",
		"awayIndicator": "#cdbd4e",
		"buttonColor":   "invalid",
	}
	user.PreUpdate()

	if user.ThemeProps["codeTheme"] != "github" || user.ThemeProps["awayIndicator"] != "#cdbd4e" {
		t.Fatal("shouldn't have changed valid theme props")
	} else if user.ThemeProps["buttonColor"] != "#ffffff" {
		t.Fatal("should've changed invalid theme prop")
	}
}

func TestUserUpdateMentionKeysFromUsername(t *testing.T) {
	user := User{Username: "user"}
	user.SetDefaultNotifications()

	if user.NotifyProps["mention_keys"] != "user,@user" {
		t.Fatal("default mention keys are invalid: %v", user.NotifyProps["mention_keys"])
	}

	user.Username = "person"
	user.UpdateMentionKeysFromUsername("user")
	if user.NotifyProps["mention_keys"] != "person,@person" {
		t.Fatal("mention keys are invalid after changing username: %v", user.NotifyProps["mention_keys"])
	}

	user.NotifyProps["mention_keys"] += ",mention"
	user.UpdateMentionKeysFromUsername("person")
	if user.NotifyProps["mention_keys"] != "person,@person,mention" {
		t.Fatal("mention keys are invalid after adding extra mention keyword: %v", user.NotifyProps["mention_keys"])
	}

	user.Username = "user"
	user.UpdateMentionKeysFromUsername("person")
	if user.NotifyProps["mention_keys"] != "user,@user,mention" {
		t.Fatal("mention keys are invalid after changing username with extra mention keyword: %v", user.NotifyProps["mention_keys"])
	}
}

func TestUserIsValid(t *testing.T) {
	user := User{}

	if err := user.IsValid(); err == nil {
		t.Fatal()
	}

	user.Id = NewId()
	if err := user.IsValid(); err == nil {
		t.Fatal()
	}

	user.CreateAt = GetMillis()
	if err := user.IsValid(); err == nil {
		t.Fatal()
	}

	user.UpdateAt = GetMillis()
	if err := user.IsValid(); err == nil {
		t.Fatal()
	}

	user.Username = NewId() + "^hello#"
	if err := user.IsValid(); err == nil {
		t.Fatal()
	}

	user.Username = NewId()
	user.Email = strings.Repeat("01234567890", 20)
	if err := user.IsValid(); err == nil {
		t.Fatal()
	}

	user.Email = "test@nowhere.com"
	user.Nickname = strings.Repeat("01234567890", 20)
	if err := user.IsValid(); err == nil {
		t.Fatal()
	}

	user.Nickname = ""
	if err := user.IsValid(); err != nil {
		t.Fatal(err)
	}

	user.FirstName = ""
	user.LastName = ""
	if err := user.IsValid(); err != nil {
		t.Fatal(err)
	}

	user.FirstName = strings.Repeat("01234567890", 20)
	if err := user.IsValid(); err == nil {
		t.Fatal(err)
	}

	user.FirstName = ""
	user.LastName = strings.Repeat("01234567890", 20)
	if err := user.IsValid(); err == nil {
		t.Fatal(err)
	}
}

func TestUserGetFullName(t *testing.T) {
	user := User{}

	if fullName := user.GetFullName(); fullName != "" {
		t.Fatal("Full name should be blank")
	}

	user.FirstName = "first"
	if fullName := user.GetFullName(); fullName != "first" {
		t.Fatal("Full name should be first name")
	}

	user.FirstName = ""
	user.LastName = "last"
	if fullName := user.GetFullName(); fullName != "last" {
		t.Fatal("Full name should be last name")
	}

	user.FirstName = "first"
	if fullName := user.GetFullName(); fullName != "first last" {
		t.Fatal("Full name should be first name and last name")
	}
}

func TestUserGetDisplayName(t *testing.T) {
	user := User{Username: "user"}

	if displayName := user.GetDisplayName(); displayName != "user" {
		t.Fatal("Display name should be username")
	}

	user.FirstName = "first"
	user.LastName = "last"
	if displayName := user.GetDisplayName(); displayName != "first last" {
		t.Fatal("Display name should be full name")
	}

	user.Nickname = "nickname"
	if displayName := user.GetDisplayName(); displayName != "nickname" {
		t.Fatal("Display name should be nickname")
	}
}

var usernames = []struct {
	value    string
	expected bool
}{
	{"spin-punch", true},
	{"Spin-punch", false},
	{"spin punch-", false},
	{"spin_punch", true},
	{"spin", true},
	{"PUNCH", false},
	{"spin.punch", true},
	{"spin'punch", false},
	{"spin*punch", false},
	{"all", false},
}

func TestValidUsername(t *testing.T) {
	for _, v := range usernames {
		if IsValidUsername(v.value) != v.expected {
			t.Errorf("expect %v as %v", v.value, v.expected)
		}
	}
}

func TestCleanUsername(t *testing.T) {
	if CleanUsername("Spin-punch") != "spin-punch" {
		t.Fatal("didn't clean name properly")
	}
	if CleanUsername("PUNCH") != "punch" {
		t.Fatal("didn't clean name properly")
	}
	if CleanUsername("spin'punch") != "spin-punch" {
		t.Fatal("didn't clean name properly")
	}
	if CleanUsername("spin") != "spin" {
		t.Fatal("didn't clean name properly")
	}
	if len(CleanUsername("all")) != 27 {
		t.Fatal("didn't clean name properly")
	}
}

func TestRoles(t *testing.T) {

	if IsValidUserRoles("admin") {
		t.Fatal()
	}

	if IsValidUserRoles("junk") {
		t.Fatal()
	}

	if IsInRole("system_admin junk", "admin") {
		t.Fatal()
	}

	if !IsInRole("system_admin junk", "system_admin") {
		t.Fatal()
	}

	if IsInRole("admin", "system_admin") {
		t.Fatal()
	}
}
