package settings

import (
	"fmt"
	"time"
)

// RegionalProfile is the persisted store presentation and time policy.
type RegionalProfile struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	Timezone    string `json:"timezone"`
	DateFormat  string `json:"date_format"`
	TimeFormat  string `json:"time_format"`
	Currency    string `json:"currency"`
	Locale      string `json:"locale"`
}

var regionalProfiles = []RegionalProfile{
	{CountryCode: "PS", CountryName: "فلسطين", Timezone: "Asia/Jerusalem", DateFormat: "DD/MM/YYYY", TimeFormat: "24h", Currency: "ILS", Locale: "ar"},
	{CountryCode: "IL", CountryName: "إسرائيل", Timezone: "Asia/Jerusalem", DateFormat: "DD/MM/YYYY", TimeFormat: "12h", Currency: "ILS", Locale: "ar"},
	{CountryCode: "SA", CountryName: "السعودية", Timezone: "Asia/Jerusalem", DateFormat: "DD/MM/YYYY", TimeFormat: "12h", Currency: "SAR", Locale: "ar"},
	{CountryCode: "AE", CountryName: "الإمارات", Timezone: "Asia/Jerusalem", DateFormat: "DD/MM/YYYY", TimeFormat: "12h", Currency: "AED", Locale: "ar"},
	{CountryCode: "JO", CountryName: "الأردن", Timezone: "Asia/Jerusalem", DateFormat: "DD/MM/YYYY", TimeFormat: "24h", Currency: "JOD", Locale: "ar"},
	{CountryCode: "EG", CountryName: "مصر", Timezone: "Asia/Jerusalem", DateFormat: "DD/MM/YYYY", TimeFormat: "24h", Currency: "EGP", Locale: "ar"},
}

func DefaultRegionalProfile() RegionalProfile {
	return regionalProfiles[1]
}

func RegionalProfiles() []RegionalProfile {
	profiles := make([]RegionalProfile, len(regionalProfiles))
	copy(profiles, regionalProfiles)
	return profiles
}

func RegionalProfileForCountry(code string) (RegionalProfile, error) {
	for _, profile := range regionalProfiles {
		if profile.CountryCode == code {
			if _, err := time.LoadLocation(profile.Timezone); err != nil {
				return RegionalProfile{}, fmt.Errorf("load timezone %s: %w", profile.Timezone, err)
			}
			return profile, nil
		}
	}
	return RegionalProfile{}, fmt.Errorf("unsupported country code %q", code)
}
