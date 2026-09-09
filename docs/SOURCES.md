# Implementation specifications

BamtiGraph contains strict TypeScript and native Go implementations migrated from the author's supplied sources. Both use the root BSD-3-Clause license. The following specifications describe their platform and format contracts; no third-party font files are bundled.

- WHATWG HTML, Canvas element and Canvas2D: https://html.spec.whatwg.org/multipage/canvas.html
- ECMA-262, ECMAScript language specification (typed arrays, BigInt and numerical operations): https://tc39.es/ecma262/
- ECMA-402, internationalization/timezone formatting: https://402.ecma-international.org/
- W3C PNG specification, third edition (chunks, filters, alpha and textual metadata): https://www.w3.org/TR/png-3/
- IETF RFC 1950, zlib: https://www.rfc-editor.org/rfc/rfc1950
- IETF RFC 1951, DEFLATE: https://www.rfc-editor.org/rfc/rfc1951
- IETF RFC 4180, CSV: https://www.rfc-editor.org/rfc/rfc4180

PNG output is independently decompressed in tests with Node's standard-library zlib and decoded in the browser. These references are not a claim of exhaustive conformance certification. CSV handling deliberately imposes strict numeric/time semantics beyond the generic CSV grammar.
