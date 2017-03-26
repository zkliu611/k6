/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2016 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package js

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCryptoMD5(t *testing.T) {
	if testing.Short() {
		return
	}

	snippet := `
	import { _assert } from "k6"
	import crypto from "k6/crypto"

	export default function() {
		let hash = crypto.md5("hello world")
		_assert(hash === "5eb63bbbe01eeed093cb22bb8f5acdc3")
	}
	`

	assert.NoError(t, runSnippet(snippet))
}

func TestCryptoSHA1(t *testing.T) {
	if testing.Short() {
		return
	}

	snippet := `
	import { _assert } from "k6"
	import crypto from "k6/crypto"

	export default function() {
		let hash = crypto.sha1("hello world")
		_assert(hash === "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed")
	}
	`

	assert.NoError(t, runSnippet(snippet))
}
