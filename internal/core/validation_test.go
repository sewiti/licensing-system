package core

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidUsername(t *testing.T) {
	tests := []struct {
		username string
		want     bool
	}{
		{"mindaugas", true},
		{"dev", true},
		{"dev ", false},
		{"as", false},
		{"maxlengthmaxlengthmaxlengthmaxlengthmaxlengthmaxlengthmaxlengthmaxlengthmaxlength", false},
	}
	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			assert.Equal(t, tt.want, ValidUsername(tt.username))
		})
	}
}

func TestValidLicenseTags(t *testing.T) {
	tests := []struct {
		tags []string
		want bool
	}{
		{[]string{"tag-1", "tag-2"}, true},
		{[]string{"", "tag"}, false},
		{[]string{"maxlengthmaxlengthmaxlengthmaxlengthmaxlengthmaxlengthmaxlengthmaxlengthmaxlength"}, false},
		{[]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20", "21"}, false},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.tags, ";"), func(t *testing.T) {
			assert.Equal(t, tt.want, ValidLicenseTags(tt.tags))
		})
	}
}

func TestValidEmail(t *testing.T) {
	tests := []struct {
		email string
		want  bool
	}{
		{"email@test.com", true},
		{"email@test", true},
		{"email", false},
		{"", false},
		{"email@test@com", false},
		{"Email Test <email@test.com>", false},
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			assert.Equal(t, tt.want, ValidEmail(tt.email))
		})
	}
}

func TestValidPhoneNumber(t *testing.T) {
	tests := []struct {
		phoneNumber string
		want        bool
	}{
		{"+370 123 56", true},
		{"+370(123)123", true},
		{"+370(1(2)3)234", false},
		{"+370-(123)-231", true},
	}
	for _, tt := range tests {
		t.Run(tt.phoneNumber, func(t *testing.T) {
			assert.Equal(t, tt.want, ValidPhoneNumber(tt.phoneNumber))
		})
	}
}
