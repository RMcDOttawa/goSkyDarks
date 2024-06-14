package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBiasStringParser(t *testing.T) {

	t.Run("parse valid string", func(t *testing.T) {
		count, binning, err := ParseBiasSet("16,1")
		require.Nil(t, err, "Valid string should not return an error")
		require.Equal(t, 16, count, "Count should be 16")
		require.Equal(t, 1, binning, "Binning should be 1")
	})

	t.Run("fail on wrong number of tokens", func(t *testing.T) {
		_, _, err := ParseBiasSet("16,1,9")
		require.NotNil(t, err, "Wrong number of tokens should return an error")
		require.ErrorContains(t, err, "must have 2 parts")
	})

	t.Run("fail on empty string", func(t *testing.T) {
		_, _, err := ParseBiasSet("")
		require.NotNil(t, err, "Empty string should return an error")
		require.ErrorContains(t, err, "must have 2 parts")
	})

	t.Run("fail on garbage frame count", func(t *testing.T) {
		_, _, err := ParseBiasSet("junk,1")
		require.NotNil(t, err, "Invalid frame count should return error")
		require.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("fail on frame count out of range", func(t *testing.T) {
		_, _, err := ParseBiasSet("0,1")
		require.NotNil(t, err, "frame count < 1 should return error")
		require.ErrorContains(t, err, "must be > 0")
	})

	t.Run("fail on garbage binning", func(t *testing.T) {
		_, _, err := ParseBiasSet("10,junk")
		require.NotNil(t, err, "Invalid binning should return error")
		require.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("fail on binning out of range", func(t *testing.T) {
		_, _, err := ParseBiasSet("10,0")
		require.NotNil(t, err, "binning < 1 should return error")
		require.ErrorContains(t, err, "must be > 0")
	})

}

func TestDarkStringParser(t *testing.T) {

	t.Run("parse valid string", func(t *testing.T) {
		count, exposure, binning, err := ParseDarkSet("16,5.0,1")
		require.Nil(t, err, "Valid string should not return an error")
		require.Equal(t, 16, count, "Count should be 16")
		require.Equal(t, 5.0, exposure, "Exposure should be 5")
		require.Equal(t, 1, binning, "Binning should be 1")
	})

	t.Run("fail on wrong number of tokens", func(t *testing.T) {
		_, _, _, err := ParseDarkSet("16,1")
		require.NotNil(t, err, "Wrong number of tokens should return an error")
		require.ErrorContains(t, err, "must have 3 parts")
	})

	t.Run("fail on empty string", func(t *testing.T) {
		_, _, _, err := ParseDarkSet("")
		require.NotNil(t, err, "Empty string should return an error")
		require.ErrorContains(t, err, "must have 3 parts")
	})

	t.Run("fail on garbage frame count", func(t *testing.T) {
		_, _, _, err := ParseDarkSet("junk,10,1")
		require.NotNil(t, err, "Invalid frame count should return error")
		require.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("fail on frame count out of range", func(t *testing.T) {
		_, _, _, err := ParseDarkSet("0,10,1")
		require.NotNil(t, err, "frame count < 1 should return error")
		require.ErrorContains(t, err, "must be > 0")
	})

	t.Run("fail on garbage exposure", func(t *testing.T) {
		_, _, _, err := ParseDarkSet("10,junk,1")
		require.NotNil(t, err, "Invalid binning should return error")
		require.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("fail on exposure out of range", func(t *testing.T) {
		_, _, _, err := ParseDarkSet("10,0,1")
		require.NotNil(t, err, "binning < 1 should return error")
		require.ErrorContains(t, err, "must be > 0")
	})

	t.Run("fail on garbage binning", func(t *testing.T) {
		_, _, _, err := ParseDarkSet("10,10,junk")
		require.NotNil(t, err, "Invalid binning should return error")
		require.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("fail on binning out of range", func(t *testing.T) {
		_, _, _, err := ParseDarkSet("10,10,0")
		require.NotNil(t, err, "binning < 1 should return error")
		require.ErrorContains(t, err, "must be > 0")
	})

}
