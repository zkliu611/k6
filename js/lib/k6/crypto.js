/**
 * @module k6/crypto
 */

/**
 * Calculate MD5 hash of input.
 * @see    module:k6/crypto.md5
 * @param  {string} input         Input to calculate hash on
 * @return {string}
 */
export function md5(input) {
    return __jsapi__.CryptoMD5(input);
};

/**
 * Calculate SHA1 hash of input.
 * @see    module:k6/crypto.sha1
 * @param  {string} input         Input to calculate hash on
 * @return {string}
 */
export function sha1(input) {
    return __jsapi__.CryptoSHA1(input);
};

export default {
	md5,
	sha1,
};