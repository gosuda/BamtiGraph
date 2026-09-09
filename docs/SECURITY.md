# Security and deployment notes

The shipped client renderer and examples do not perform data-source network requests, analytics, persistence, cookie writes, dynamic code evaluation or remote font loading. The library does not collect input. Local CSV files are read only after the user selects them. The live example is an explicitly bounded, initially paused simulation.

Runtime data is rendered as pixels or assigned through `textContent`; it is not inserted as executable HTML. Configuration rejects common prototype-pollution property names. This does not turn arbitrary JavaScript objects into a safe serialization format: do not pass untrusted getters, cyclic objects, proxies or code. Parse untrusted JSON separately and enforce an application schema. Resource caps are per operation, not a process-wide memory budget. Consider stricter application limits for user-controlled large inputs. Rendering and compression are synchronous.

Downloads use browser Blob URLs created locally and revoked after a grace period. Use a filename, not a path. PNG metadata can contain the title, sample statistics and configuration supplied by the application; disable metadata when that information must not leave the application. Raw samples are not embedded by the renderer, but labels/statistics may themselves be sensitive. The package is not an authorization, redaction or encryption layer.

The TypeScript PNG helper checks headers and dimensions before requesting native decoding. Go's image loader uses the standard library. Neither substitutes for an independent security audit. TypeScript bitmap patterns are source data; system text delegates to browser fonts. Native Go includes a TrueType-outline font parser and rasterizer: use trusted installed TTF/TTC fonts, not arbitrary untrusted uploads. No font binaries are bundled.

## Content Security Policy

A hosted multi-file integration can keep script loading local (`script-src 'self'`) and deny data connections (`connect-src 'none'`), while serving its static CSS locally. The controller creates local inline style properties. Validate the exact style policy with the target browser; a style policy that blocks all such styling may require adapting controller presentation. The standalone file contains inline scripts/styles and therefore needs hashes or an appropriate host policy when served under restrictive CSP. Do not weaken a customer CSP or browser policy merely to run the demo. There is no preconfigured unsafe-eval requirement.

The optional `tools/serve.cjs` binds loopback only, serves GET/HEAD, performs path/realpath checks and sets `nosniff`. It is a development aid, not an authenticated internet-facing server. Production hosting, TLS, authentication, input authorization, retention policy, and real data collection belong to the integrating application.

Native Go metadata records font basenames, hashes, backend/runtime identifiers, and graph configuration, but not font bytes. Use `result.EncodePNG(writer, false)` to omit it. Go graph configuration is mutable; callers must synchronize concurrent mutation. Concurrent renders of an unchanged graph are supported.

This delivery includes functional and regression checks. It does not claim a penetration test, security certification or complete assistive-technology audit.
