# BamtiGraph TypeScript API reference

Version 1.0.0. Import default or named exports from npm `bamtigraph` or JSR `@safe/bantigraph`. The browser bundle `dist/bamtigraph.global.js` defines `globalThis.BamtiGraph`. ESM/CommonJS exports and declarations are generated from strict TypeScript in `src/index.ts`; no runtime dependency is required.

## Data model and ownership

`series(name, timestamps, values, options?)` returns a copied, deeply frozen series. `regularSeries(name, values, start, step=300, options?)` constructs regular timestamps. Numbers are **Unix seconds**, never automatically detected milliseconds. Inputs may use explicit `Date` objects or timezone-aware ISO strings. Accepted ISO form is `YYYY-MM-DDTHH:MM[:SS[.fff]]Z` or the equivalent with `±HH:MM`. Years 1–9999 are supported subject to valid viewport endpoints. Duplicate or decreasing times and Infinity are rejected. Missing values are `null`, `undefined`, or `NaN`. No automatic sorting, unit conversion or extrapolation occurs.

Series options:

| Field | Default | Meaning |
|---|---|---|
| `kind` | `line` | `line` or `area`; overlapping areas, not stacking |
| `color` | `#0000cc` | line or fill RGB(A) |
| `outline` | `null` | area outline color only |
| `lineWidth` | `0.7` | logical pixel stroke width, 0.01–128 |
| `baseline` | `0` | area base in raw data units |
| `interpolation` | `linear` | `linear` or `step-post` |
| `gapAfter` | `0` | break when adjacent timestamps are farther apart; zero disables this threshold |
| `legendValues` | `null` | display-only `{current, average, maximum}` override; recorded separately from actual statistics |

Every series may have independent timestamps. A missing value breaks both adjacent segments. A singleton line is a point; a singleton area without an outline has no width. Sample arrays and chart configurations are copied and frozen. `Chart.with()` returns a new chart. Rendered RGBA bytes are intentionally mutable; editing them changes later exports but does not recompute metadata.

## Chart and rendering

```js
const chart = new BamtiGraph.Chart({
  series: [seriesA, seriesB], title: '', verticalLabel: '', watermark: '',
  timeAxis: {}, yAxis: {}, layout: {}, theme: {}, fonts: {},
  legendLabels: ['Current:', 'Average:', 'Maximum:'], missingLabel: 'NaN',
  hRules: [], vRules: []
});
```

Options are recursively merged with defaults. Arrays are replaced, not appended. Do not use `undefined` as a substitute for omitting an optional nested field. Null has explicit meaning only for fields documented as nullable. Bad inputs normally throw `RangeError`; native API errors are not wrapped. Construction validates most options; text fitting, viewport resolution and raster limits may also fail during rendering. The input is ordinary configuration, not an adversarial executable object: getters, cycles and exotic prototypes are not supported.

`traffic(timestamps, inbound, outbound, options?)` constructs a green outlined area and blue line, default title `Traffic - ether1`, vertical label `bits per second`, and empty watermark. It accepts an extra shared `gapAfter`. No counter or bytes-to-bits conversion is implied.

| Method | Result |
|---|---|
| `chart.render()` | `RenderResult` |
| `chart.draw(canvas)` | renders, draws into a Canvas2D canvas, returns `RenderResult` |
| `chart.toPNG({metadata:true})` | `Uint8Array` |
| `chart.with(patch)` | new immutable `Chart` |
| `chart.mount(canvas, options)` | browser `Controller`; canvas must already be in the document |
| `chart.nearest(epoch)` | nearest original visible observation per series, including exact observation time and missingness |
| `BamtiGraph.render(options)` | convenience one-shot render |

`RenderResult` exposes `image:{width,height,data}`, `width`, `height`, `data`, `metadata`. Its methods are `draw(canvas)`, `toPNG(options)`, `toBlob(options)` and `download(filename='chart.png', options)`. `draw` returns the target canvas. PNG contains standard chunks and optional uncompressed UTF-8 `iTXt` JSON under keyword `chart`. Set `{metadata:false}` to export pixels only. Renderer does not record current clock time, network addresses, user identifiers, system file paths or origin information.

Chart metadata includes logical/physical dimensions, axis endpoints and limits, actual drawn labels, original sample statistics, font mode, configured family, theme, layout, warnings and empty/explicit watermark. A system font family is a request to the host, not a verified font hash. With custom fonts, load them in the host application and await readiness before rendering. No font bytes are embedded.

## Time axis

`daily(optionsOrTimezone)`, `weekly(...)`, `monthly(...)`, `yearly(...)` construct axis objects. `BamtiGraph.daily('Asia/Seoul')` is shorthand for `{mode:'daily', timezone:'Asia/Seoul', ...defaults}`. Named modes affect **tick presentation only**, not resampling.

Fields: `start=null`, `end=null`, `mode='auto'`, `timezone='UTC'`, `minorSeconds=null`, `majorSeconds=null`, `labelSeconds=null`, `labelFormat=null`, `ticks=null`, `majorTicks=null`, `minorTicks=null`, `labelOffsetSeconds=0`.

Null tick arrays mean automatic. Empty arrays mean disabled. Explicit labels use `{time,label}` and are not thinned for overlap; labels extending outside panel bounds are omitted. Automatic labels may be thinned. Bounds are derived from data unless provided. Empty data needs explicit nonzero bounds. A singleton unbounded graph gets ±150 seconds without extending its data path.

Numeric intervals align to local wall-clock multiples. Actual calendar months are used for yearly ticks. Weekly labels are local noon of fully contained days. Modern spring DST gaps and fall folds are handled; duplicate daily wall labels receive numeric offsets. The environment's Intl/IANA data controls timezone history. Unusual historical sub-day offset regimes and all extreme year/timezone combinations have not been exhaustively verified. Use explicit ticks or UTC for archival reproducibility.

`labelOffsetSeconds` is an elapsed-seconds displacement, unlike default weekly calendar-noon placement. Maximum automatic/explicit tick count is 5,000. Invalid tiny intervals fail instead of attempting unlimited enumeration.

`formatTime(time, timezone='UTC', format='%H:%M')` supports `%H %I %M %S %d %e %m %Y %y %a %A %b %h %B %p %w %u %j %z %Z %F %T %R %%`. Names are fixed English. `%Z` is the supplied IANA identifier, not a localized abbreviation. Unsupported directives throw. Subseconds are retained for position but not formatted by this subset.

## Y axis and statistics

Fields: `minimum=0` (null enables automatic lower bound), `maximum=null`, `majorStep=null`, `minorDivisions=5`, `base=1000`, `scaleFactor=null`, `suffix=null`, `decimals=null`, `legendDecimals=2`, `showZeroSuffix=false`.

Limits/steps/thresholds use original data units. Base 1024 uses IEC prefixes. `scaleFactor` is an explicit numeric divisor and `suffix` its independently customizable text. All legend statistics share the resolved axis scale. Auto limits include visible interpolated paths and area baselines. Explicit Y limits clip pixels but do not alter input statistics.

Statistics consider raw observations in the inclusive `[start,end]` viewport. `current` is the last observation even when missing; `average` is an arithmetic finite-sample mean; `count` excludes missing observations; `missing` counts them. Empty statistics are null. Float computations use finite magnitude normalization and compensation; exact decimal arithmetic is not promised. Pixel-column decimation retains first/minimum/maximum/last in time order but never changes sample statistics. Irregular time-weighted averaging must be defined outside this API.

`formatValue(value, {factor,suffix}, decimals, missingLabel='NaN')` is a formatting helper for already validated finite units and decimal settings.

## Layout and style

Default two-series panel: 595×211 logical pixels. Endpoint coordinates `[64,34,564,156]` are **inclusive coordinate anchors**, not an exclusive crop rectangle. Width includes margins. Full defaults are shipped as `docs/defaults.json`.

Core fields: `width=595`, `plotHeight=122`, `left=64`, `right=31`, `top=34`, `titleY=8`, `titleOffsetX=27`, `unitX=5`, `xLabelGap=5`, `yLabelGap=6`, `legendGap=20`, `legendRowHeight=14`, `legendBottom=7`, `legend='reference'`, `antialias=4`, `pixelScale=1`.

Visible-legend height is `top + plotHeight + legendGap + max(1,seriesCount)*legendRowHeight + legendBottom`. Hidden-legend height is `top + plotHeight + 18`.

`legend` is `reference`, `aligned`, or `none`. Reference mode gives the final row expanded columns when multiple series exist. `legendLayout` defaults:

```js
{
  nameX: 30, swatchX: 15, swatchWidth: 9, swatchHeight: 10,
  referenceWidth: 595, autoScaleColumns: true,
  compact: [[102,228], [244,370], [386,512]],
  expanded: [[124,250], [289,415], [454,580]],
  aligned: [[102,250], [267,415], [432,580]]
}
```

Each pair is `[labelLeft, valueRight]`. Automatic scaling adapts numeric anchors to panel width, not the fixed name origin. Narrow panels reduce legend font size or reject an unfit statistic. Names/titles may be ellipsized; numeric cells are never silently truncated. Very long labels, watermarks and vertical text may cause fitting errors.

`Theme` controls `background`, `canvas`, `shadeLight`, `shadeDark`, `text`, `minorGrid`, `majorGrid`, `axis`, `arrow`, `watermark`, `frame`, `gridFront`, `gridDash:[on,off]`; font sizes `titleSize`, `axisSize`, `unitSize`, `legendSize`, `watermarkSize`, `captionSize`; fixed advances `titleAdvance`, `axisAdvance`, `legendAdvance`. Sizes are logical pixels, not CSS points. Background and canvas must be opaque.

Colors: hexadecimal `#RGB/#RGBA/#RRGGBB/#RRGGBBAA`, integer RGB/RGBA tuples, or the small named set black, white, red, green, blue, transparent. CSS `rgb()`, CSS variables and arbitrary color expressions are not parsed.

Fonts: `mode='system'`, `family='"DejaVu Sans Mono", "Liberation Mono", Consolas, monospace'`, `titleFamily=null`, `unitFamily=null`, `captionFamily='Arial, sans-serif'`, `strictGlyphs=true`. System mode uses installed glyphs and host rasterization. Bitmap mode uses an authored printable-ASCII 5×7 alphabet; unsupported glyphs throw or become `?` with `strictGlyphs:false`. Bitmap mode is not a copy of the system font. CJK in system mode uses wider cells; complex shaping, bidi and combining mark layout are not promised.

The default `watermark` is empty. Set it to `BAMTIGRAPH` or customer text to draw in the right-hand vertical margin; clearing it removes the mark. It is included in PNG exports and metadata. Use system fonts for non-ASCII text. Overflow is an error, not implicit truncation.

## Rules and dashboards

`hRules:[{value,color:'#990000',width:1,dash:[3,2]}]`, `vRules:[{time,color,width,dash}]`. Use `dash:null` for solid. Values/time are raw units/Unix seconds. Rules are drawn over data and clipped to the plot.

```js
const result = BamtiGraph.dashboard([
  { chart: dailyChart, caption: 'Daily samples' },
  { chart: weeklyChart, caption: 'Weekly samples' }
], { gap:24, padding:[4,3,6,6], background:'#f3f3f3' });
```

Panels retain their own sizes; all must have equal integer scale. Padding order is left/top/right/bottom. Captions do not trigger aggregation. A final caption adds a final gap. `cropHeight:null` retains all rows; a positive logical height deliberately crops without resizing. Dashboard metadata includes each child chart manifest and placement.

## Interactive controller

`chart.mount(canvas,{interactive:true,ariaLabel,onHover})` wraps the existing canvas and creates a separate overlay/tooltip. `controller.update(patchOrChart)` computes a valid replacement before changing the active chart. `controller.destroy()` is idempotent, removes handlers/overlays, and restores the original DOM position and saved accessibility/style attributes. It leaves the last rendered pixels in the canvas. Call it on component teardown. Do not mount twice on the same canvas without destroying the first controller.

Pointer inspection reports the nearest original observation per series, not interpolated values. The tooltip includes each observation's time. Keyboard: arrows step through the first nonempty series, Shift+arrow steps ten, Home/End select visible endpoints, Escape clears. The demo supplies a separate statistics table. This is not a claim of full assistive-technology certification. No automatic network polling, timers or persistence exist in the core renderer. The live example owns and cleans up its own bounded simulation timer.

## Preprocessing

`counterRate(timestamps,counters,{factor:1,onDecrease:'gap',counterBits:null,maxRate:null})` returns mutable `Samples`. Counters must be unsigned safe Number integers or BigInt values up to 128 bits. Never convert a large counter to Number before subtracting. Decreases default to a gap. Explicit wrap requires `onDecrease:'wrap'` and `counterBits` in 1–128. Resets cannot be automatically distinguished from wraps. Missing counters invalidate adjacent intervals. First rate is NaN; each result attaches to its right endpoint. Use `factor:8` for byte-to-bit conversion. Conversion of a very large difference to Number can itself round; exact subtraction does not promise arbitrary-precision rates.

`aggregate(timestamps,values,{interval:300,method:'mean',origin:0,minCoverage:0,expectedStep:null,maxBuckets:1000000})` returns `Samples`. Fixed elapsed buckets are `[left,right)`, labelled by left edge. Methods: mean/min/max/last/sum. Intermediate empty buckets remain NaN. Last preserves final missingness. Coverage denominator is the greater of observed count (including missing) and `interval/expectedStep` when provided. Means are sample-based, never time integrals; summation overflow fails. Calendar-day resampling across DST is not provided by this fixed-interval function.

## CSV

`parseCSV(text,{timestampColumn:'timestamp',columns:[...],maxRows:1000000,maxBytes:16777216})` returns series. Default mappings are inbound→Inbound green area and outbound→Outbound blue line. A custom mapping is `{column:'cpu',name:'CPU',kind:'area',color:'#00cc00'}`. BOM, quoted fields, escaped double quotes, CRLF/LF and blank lines are supported. Header names must be unique. Ragged rows, unclosed quotes, malformed numbers, duplicate times and naive dates throw. Blank/NaN/None/null numeric cells are missing. Parser is in-memory, not streaming.

`toCSV(series,{escapeFormulas:true})` unions independent timestamps. Series names must be unique and not `timestamp`. It escapes formula-like **text headers** by default; finite numeric cells remain numeric, including negative values. Exported missing cells are blank. The export may contain gaps for a timestamp present in another series. Custom exported names need corresponding parse mappings. File reading and downloading are example/application responsibilities.

## Pixel tools and resource limits

`compareImages(reference,actual,{tolerance:0,box:null})` takes RGBA images or render results; dimensions must match. Box is `[left,top,right,bottom]` with exclusive right/bottom, relative to image origin. There is no aligning/resizing. RGB is compared; alpha is ignored. Returned counts/ratios refer to whole pixels, not channels. MAE/RMSE refer to channel intensities in 0–255. `differenceBox` is relative to the chosen crop.

`differenceImage(reference,actual,amplify=4)` returns a render result. `readCanvas(canvas)` returns RGBA. `decodeImage(blob)` validates a PNG header and size before using the browser's decoder; PNG only, 64 MiB encoded limit, 16 million pixel limit. This helper is not an untrusted-image security sandbox. `encodePNG(image,metadata=null)` performs its own PNG/DEFLATE encoding.

Default guards: 16M output pixels; 40M pixels per supersampled layer; 2M samples per series; 128 series/panels; 5,000 ticks; panel width 400–8192; plot height 30–4096; antialias and integer scale 1–8; single-line strings at most 4,096 code units; metadata at most 1,048,576 JSON code units. Aggregate/CSV have additional explicit caps. These are individual allocation/loop guards, **not a global process memory budget or asynchronous rendering guarantee**. Rendering and PNG compression are synchronous. Preaggregate large data and schedule updates deliberately.
