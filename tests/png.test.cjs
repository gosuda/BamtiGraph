'use strict';
const { test } = require('node:test'), assert = require('node:assert/strict'), zlib = require('node:zlib');
const T = require('bamtigraph');
function crc(bytes) { let c = 0xffffffff; for (const b of bytes) {
    c ^= b;
    for (let k = 0; k < 8; k++)
        c = (c & 1) ? (c >>> 1) ^ 0xedb88320 : c >>> 1;
} return (c ^ 0xffffffff) >>> 0; }
function decode(bytes) {
    const b = Buffer.from(bytes);
    assert.deepEqual([...b.subarray(0, 8)], [137, 80, 78, 71, 13, 10, 26, 10]);
    let at = 8, w, h, ended = false, metadata = null;
    const idat = [], types = [];
    while (at < b.length) {
        const n = b.readUInt32BE(at), type = b.toString('ascii', at + 4, at + 8), data = b.subarray(at + 8, at + 8 + n);
        types.push(type);
        assert.equal(crc(b.subarray(at + 4, at + 8 + n)), b.readUInt32BE(at + 8 + n));
        if (type === 'IHDR') {
            w = data.readUInt32BE(0);
            h = data.readUInt32BE(4);
            assert.deepEqual([...data.subarray(8)], [8, 6, 0, 0, 0]);
        }
        if (type === 'IDAT')
            idat.push(data);
        if (type === 'iTXt') {
            const separator = data.indexOf(0);
            metadata = JSON.parse(data.subarray(separator + 5).toString('utf8'));
        }
        at += 12 + n;
        if (type === 'IEND') {
            ended = true;
            break;
        }
    }
    assert.ok(ended);
    assert.equal(at, b.length);
    const raw = zlib.inflateSync(Buffer.concat(idat)), stride = w * 4, out = new Uint8Array(w * h * 4);
    assert.equal(raw.length, h * (stride + 1));
    function paeth(a, b, c) { const p = a + b - c, da = Math.abs(p - a), db = Math.abs(p - b), dc = Math.abs(p - c); return da <= db && da <= dc ? a : db <= dc ? b : c; }
    for (let y = 0; y < h; y++) {
        const f = raw[y * (stride + 1)];
        for (let x = 0; x < stride; x++) {
            const a = x >= 4 ? out[y * stride + x - 4] : 0, b = y ? out[(y - 1) * stride + x] : 0, c = y && x >= 4 ? out[(y - 1) * stride + x - 4] : 0, p = [0, a, b, Math.floor((a + b) / 2), paeth(a, b, c)][f];
            assert.notEqual(p, undefined);
            out[y * stride + x] = (raw[y * (stride + 1) + 1 + x] + p) & 255;
        }
    }
    return { width: w, height: h, data: out, metadata, types };
}
for (const [w, h] of [[1, 1], [7, 5], [257, 131], [400, 200]])
    test(`PNG ${w}x${h}: standard zlib inflate, CRC and RGBA round trip`, () => { let state = 31337; const data = Uint8ClampedArray.from({ length: w * h * 4 }, () => { state = (Math.imul(state, 1664525) + 1013904223) >>> 0; return state >>> 24; }), image = { width: w, height: h, data }, title = 'Traffic / \uD55C\uAE00 / \u20AC', out = decode(T.encodePNG(image, { title, values: [1, null, 3] })); assert.equal(out.width, w); assert.equal(out.height, h); assert.deepEqual(out.data, new Uint8Array(data)); assert.equal(out.metadata.title, title); });
test('PNG repeated and highly compressible data survive 32 KiB window boundaries', () => { const data = new Uint8ClampedArray(595 * 211 * 4); for (let i = 0; i < data.length; i++)
    data[i] = i % 4 === 3 ? 255 : i % 37; const png = T.encodePNG({ width: 595, height: 211, data }), out = decode(png); assert.deepEqual(out.data, new Uint8Array(data)); assert.ok(png.length < data.length / 5); });
test('PNG metadata is omitted when disabled', () => { const r = new T.Chart({ fonts: { mode: 'bitmap' }, series: [T.series('A', [0, 1], [1, 2])] }).render(), a = decode(r.toPNG()), b = decode(r.toPNG({ metadata: false })); assert.ok(a.types.includes('iTXt')); assert.ok(!b.types.includes('iTXt')); assert.deepEqual(a.data, b.data); });
test('PNG malformed image shapes and oversized metadata fail early', () => { assert.throws(() => T.encodePNG({ width: 2, height: 2, data: new Uint8Array(1) })); assert.throws(() => T.encodePNG({ width: 1, height: 1, data: new Uint8Array(4) }, { x: 'x'.repeat(1048577) })); });
module.exports = { decode };
