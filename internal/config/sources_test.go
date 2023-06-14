package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var sources = []Sources{
	{
		Name:           "test1",
		Headless:       true,
		Active:         true,
		SourceType:     "sitemap",
		IdentifierType: IdentifierSha,
	},
	{
		Name:       "test2",
		Headless:   false,
		Active:     true,
		SourceType: "robots",
	},
	{
		Name:           "test3",
		Headless:       false,
		Active:         false,
		SourceType:     "sitemap",
		IdentifierType: JsonSha,
	},
	{
		Name:       "test4",
		Headless:   true,
		Active:     false,
		SourceType: "sitemap",
	},
}

var empty = []Sources{}

func TestGetSourceByType(t *testing.T) {
	t.Run("It gets sources of the given type", func(t *testing.T) {
		expected := []Sources{
			{
				Name:           "test1",
				Headless:       true,
				Active:         true,
				SourceType:     "sitemap",
				IdentifierType: IdentifierSha,
			},
			{
				Name:           "test3",
				Headless:       false,
				Active:         false,
				SourceType:     "sitemap",
				IdentifierType: JsonSha,
			},
			{
				Name:       "test4",
				Headless:   true,
				Active:     false,
				SourceType: "sitemap",
			},
		}
		results := GetSourceByType(sources, "sitemap")
		assert.ElementsMatch(t, expected, results)
	})

	t.Run("It returns an empty slice if there are no such sources", func(t *testing.T) {
		results := GetSourceByType(sources, "csv")
		assert.ElementsMatch(t, empty, results)
	})

	t.Run("It handles an empty source slice correctly", func(t *testing.T) {
		results := GetSourceByType(empty, "sitemap")
		assert.ElementsMatch(t, empty, results)
	})
}

func TestGetActiveSourceByType(t *testing.T) {
	t.Run("It gets active sources of the given type", func(t *testing.T) {
		expected := []Sources{
			{
				Name:       "test1",
				Headless:   true,
				Active:     true,
				SourceType: "sitemap",
				Logo: "",
				URL: "",
				PID: "",
				ProperName: "",
				Domain: "",
				CredentialsFile: "",
				Other: nil,
				HeadlessWait: 0,
				Delay: 0,
				IdentifierPath: "",
				ApiPageLimit: 0,
				IdentifierType: IdentifierSha,
				FixContextOption: 0,
			},
		}
		results := GetActiveSourceByType(sources, "sitemap")
		assert.ElementsMatch(t, expected, results)
	})

	t.Run("It returns an empty slice if there are no such sources", func(t *testing.T) {
		results := GetActiveSourceByType(sources, "csv")
		assert.ElementsMatch(t, empty, results)
	})

	t.Run("It handles an empty source slice correctly", func(t *testing.T) {
		results := GetActiveSourceByType(empty, "sitemap")
		assert.ElementsMatch(t, empty, results)
	})
}

func TestGetActiveSourceByHeadless(t *testing.T) {
	t.Run("It gets active sources of the given type", func(t *testing.T) {
		expectedTrue := []Sources{
			{
				Name:       "test1",
				Headless:   true,
				Active:     true,
				SourceType: "sitemap",
				Logo: "",
				URL: "",
				PID: "",
				ProperName: "",
				Domain: "",
				CredentialsFile: "",
				Other: nil,
				HeadlessWait: 0,
				Delay: 0,
				IdentifierPath: "",
				ApiPageLimit: 0,
				IdentifierType: IdentifierSha,
				FixContextOption: 0,
			},
		}
		results := GetActiveSourceByHeadless(sources, true)
		assert.ElementsMatch(t, expectedTrue, results)

		expectedFalse := []Sources{
			{
				Name:       "test2",
				Headless:   false,
				Active:     true,
				SourceType: "robots",
				Logo: "",
				URL: "",
				PID: "",
				ProperName: "",
				Domain: "",
				CredentialsFile: "",
				Other: nil,
				HeadlessWait: 0,
				Delay: 0,
				IdentifierPath: "",
				ApiPageLimit: 0,
				IdentifierType: "",
				FixContextOption: 0,
			},
		}
		results = GetActiveSourceByHeadless(sources, false)
		assert.ElementsMatch(t, expectedFalse, results)
	})

	t.Run("It returns an empty slice if there are no such sources", func(t *testing.T) {
		test := []Sources{
			{
				Name:       "test1",
				Headless:   true,
				Active:     true,
				SourceType: "sitemap",
			},
			{
				Name:       "test3",
				Headless:   false,
				Active:     false,
				SourceType: "sitemap",
			},
			{
				Name:       "test4",
				Headless:   true,
				Active:     false,
				SourceType: "sitemap",
			},
		}
		results := GetActiveSourceByHeadless(test, false)
		assert.ElementsMatch(t, empty, results)
	})

	t.Run("It handles an empty source slice correctly", func(t *testing.T) {
		results := GetActiveSourceByHeadless(empty, true)
		assert.ElementsMatch(t, empty, results)
	})
}

func TestGetSourceByName(t *testing.T) {
	t.Run("It gets sources of the given name", func(t *testing.T) {
		expected := Sources{
			Name:       "test1",
			Headless:   true,
			Active:     true,
			SourceType: "sitemap",
			Logo: "",
			URL: "",
			PID: "",
			ProperName: "",
			Domain: "",
			CredentialsFile: "",
			Other: nil,
			HeadlessWait: 0,
			Delay: 0,
			IdentifierPath: "",
			ApiPageLimit: 0,
			IdentifierType: IdentifierSha,
			FixContextOption: 0,
		}

		results, err := GetSourceByName(sources, "test1")
		assert.EqualValues(t, &expected, results)
		assert.Nil(t, err)

	})

	t.Run("It returns an empty slice if there are no such sources", func(t *testing.T) {
		results, err := GetSourceByName(sources, "test99")
		assert.ElementsMatch(t, empty, results)
		assert.NotNil(t, err)
	})

	t.Run("It handles an empty source slice correctly", func(t *testing.T) {
		results, err := GetSourceByName(empty, "test1")
		assert.ElementsMatch(t, empty, results)
		assert.NotNil(t, err)
	})
}
