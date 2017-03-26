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
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"golang.org/x/crypto/ripemd160"
)

func (a JSAPI) CryptoMD5(input string) string {
	hasher := md5.New()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoRIPEMD160(input string) string {
	hasher := ripemd160.New()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoSHA1(input string) string {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoSHA224(input string) string {
	hasher := sha256.New224()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoSHA256(input string) string {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoSHA384(input string) string {
	hasher := sha512.New384()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoSHA512(input string) string {
	hasher := sha512.New()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoSHA512_224(input string) string {
	hasher := sha512.New512_224()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoSHA512_256(input string) string {
	hasher := sha512.New512_256()
	_, err := hasher.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func (a JSAPI) CryptoHMAC_MD5(input, key string) string {
	sig := hmac.New(md5.New, []byte(key))
	_, err := sig.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(sig.Sum(nil))
}

func (a JSAPI) CryptoHMAC_SHA1(input, key string) string {
	sig := hmac.New(sha1.New, []byte(key))
	_, err := sig.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(sig.Sum(nil))
}

func (a JSAPI) CryptoHMAC_SHA224(input, key string) string {
	sig := hmac.New(sha256.New224, []byte(key))
	_, err := sig.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(sig.Sum(nil))
}

func (a JSAPI) CryptoHMAC_SHA256(input, key string) string {
	sig := hmac.New(sha256.New, []byte(key))
	_, err := sig.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(sig.Sum(nil))
}

func (a JSAPI) CryptoHMAC_SHA384(input, key string) string {
	sig := hmac.New(sha512.New384, []byte(key))
	_, err := sig.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(sig.Sum(nil))
}

func (a JSAPI) CryptoHMAC_SHA512(input, key string) string {
	sig := hmac.New(sha512.New, []byte(key))
	_, err := sig.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(sig.Sum(nil))
}

func (a JSAPI) CryptoHMAC_RIPEMD160(input, key string) string {
	sig := hmac.New(ripemd160.New, []byte(key))
	_, err := sig.Write([]byte(input))
	if err != nil {
		throw(a.vu.vm, err)
	}
	return hex.EncodeToString(sig.Sum(nil))
}
