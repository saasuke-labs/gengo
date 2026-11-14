package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripYAMLFrontMatter(t *testing.T) {
	input := []byte(`---
title: Discover Barcelona
tags: ["Barcelona"]
---
# Barcelona
City body.
`)

	result := stripYAMLFrontMatter(input)

	assert.Equal(t, "# Barcelona\nCity body.\n", string(result))
}
