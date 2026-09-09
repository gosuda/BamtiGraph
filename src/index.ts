/*
 * SPDX-License-Identifier: BSD-3-Clause
 * Copyright (c) 2026, GoSuda
 * BamtiGraph TypeScript implementation. See the root LICENSE.
 */
/** Unix seconds, an explicit Date, or an ISO datetime with a timezone offset. */
export type Timestamp = number | Date | string;
/** RGB or RGBA channels, hexadecimal notation, or a supported named color. */
export type Color =
  | string
  | readonly [number, number, number]
  | readonly [number, number, number, number];
export type Values = ArrayLike<number | null | undefined>;
/** Normalized red, green, blue, and alpha bytes. */
export type RGBA = [number, number, number, number];
export type Interpolation = "linear" | "step-post";
/** Mutable preprocessing output, with NaN marking missing samples. */
export interface Samples {
  timestamps: number[];
  values: number[];
}
export interface LegendValues {
  current: number | null;
  average: number | null;
  maximum: number | null;
}
export interface SeriesOptions {
  kind?: "line" | "area";
  color?: Color;
  outline?: Color | null;
  lineWidth?: number;
  baseline?: number;
  interpolation?: Interpolation;
  gapAfter?: number;
  legendValues?: Partial<LegendValues> | null;
}
export interface SeriesInput extends SeriesOptions {
  name: string;
  timestamps: ArrayLike<Timestamp>;
  values: Values;
}
export interface Series extends Readonly<
  Required<Omit<SeriesOptions, "legendValues" | "color" | "outline">>
> {
  readonly color: Readonly<RGBA>;
  readonly outline: Readonly<RGBA> | null;
  readonly name: string;
  readonly timestamps: readonly number[];
  readonly values: readonly number[];
  readonly legendValues: Readonly<LegendValues> | null;
}
export interface Tick {
  time: Timestamp;
  label: string;
}
export interface TimeAxis {
  start?: Timestamp | null;
  end?: Timestamp | null;
  mode?: "auto" | "daily" | "weekly" | "monthly" | "yearly" | "custom";
  timezone?: string;
  minorSeconds?: number | null;
  majorSeconds?: number | null;
  labelSeconds?: number | null;
  labelFormat?: string | null;
  labelOffsetSeconds?: number;
  ticks?: readonly Tick[] | null;
  minorTicks?: readonly Timestamp[] | null;
  majorTicks?: readonly Timestamp[] | null;
}
export interface YAxis {
  minimum?: number | null;
  maximum?: number | null;
  majorStep?: number | null;
  minorDivisions?: number;
  base?: 1000 | 1024;
  scaleFactor?: number | null;
  suffix?: string | null;
  decimals?: number | null;
  legendDecimals?: number;
  showZeroSuffix?: boolean;
}
export type ColumnAnchors = readonly [
  readonly [number, number],
  readonly [number, number],
  readonly [number, number],
];
export interface LegendLayout {
  nameX?: number;
  swatchX?: number;
  swatchWidth?: number;
  swatchHeight?: number;
  referenceWidth?: number;
  autoScaleColumns?: boolean;
  compact?: ColumnAnchors;
  expanded?: ColumnAnchors;
  aligned?: ColumnAnchors;
}
export interface Layout {
  width?: number;
  plotHeight?: number;
  left?: number;
  right?: number;
  top?: number;
  titleY?: number;
  titleOffsetX?: number;
  unitX?: number;
  xLabelGap?: number;
  yLabelGap?: number;
  legendGap?: number;
  legendRowHeight?: number;
  legendBottom?: number;
  legend?: "reference" | "aligned" | "none";
  legendLayout?: LegendLayout;
  antialias?: number;
  pixelScale?: number;
}
export interface Theme {
  background?: Color;
  canvas?: Color;
  shadeLight?: Color;
  shadeDark?: Color;
  text?: Color;
  minorGrid?: Color;
  majorGrid?: Color;
  axis?: Color;
  arrow?: Color;
  watermark?: Color;
  frame?: Color;
  gridFront?: boolean;
  gridDash?: readonly [number, number];
  titleSize?: number;
  axisSize?: number;
  unitSize?: number;
  legendSize?: number;
  watermarkSize?: number;
  captionSize?: number;
  titleAdvance?: number;
  axisAdvance?: number;
  legendAdvance?: number;
}
export interface FontOptions {
  mode?: "system" | "bitmap";
  family?: string;
  titleFamily?: string | null;
  unitFamily?: string | null;
  captionFamily?: string | null;
  strictGlyphs?: boolean;
}
export interface RuleStyle {
  color?: Color;
  width?: number;
  dash?: readonly [number, number] | null;
}
export interface HRule extends RuleStyle {
  value: number;
}
export interface VRule extends RuleStyle {
  time: Timestamp;
}
/** Immutable chart input. Nested objects merge; arrays replace their defaults. */
export interface ChartOptions {
  series?: readonly SeriesInput[];
  title?: string;
  verticalLabel?: string;
  watermark?: string;
  timeAxis?: TimeAxis;
  yAxis?: YAxis;
  layout?: Layout;
  theme?: Theme;
  fonts?: FontOptions;
  legendLabels?: readonly [string, string, string];
  missingLabel?: string;
  hRules?: readonly HRule[];
  vRules?: readonly VRule[];
}
export interface TrafficOptions extends ChartOptions {
  gapAfter?: number;
}
export interface Unit {
  factor: number;
  suffix: string;
}
export interface Statistics {
  name: string;
  current: number | null;
  average: number | null;
  maximum: number | null;
  minimum: number | null;
  count: number;
  missing: number;
  displayOverride: LegendValues | null;
}
/** Serializable chart manifest. Missing statistics are null, not NaN. */
export interface GraphMetadata {
  version: string;
  imageSize: [number, number];
  logicalSize: [number, number];
  plotBox: [number, number, number, number];
  pixelScale: number;
  title: string;
  verticalLabel: string;
  watermark: string;
  timeRange: [number, number];
  timezone: string;
  timeMode: string;
  yRange: [number, number];
  yStep: number;
  yUnit: Unit;
  xLabels: { time: number; label: string; x: number }[];
  statistics: Statistics[];
  statisticsPolicy: string;
  font: {
    mode: "system" | "bitmap";
    family: string | null;
    pixelAlphabet: string | null;
  };
  layout: Layout;
  theme: Theme;
  warnings: string[];
}
/** Row-major straight-alpha bytes. Pixel storage intentionally remains mutable. */
export interface RGBAImage {
  width: number;
  height: number;
  data: Uint8ClampedArray | Uint8Array;
}
export interface PNGOptions {
  metadata?: boolean;
}
export interface NearestSample {
  name: string;
  index: number | null;
  time: number | null;
  value: number | null;
}
export interface MountOptions {
  interactive?: boolean;
  ariaLabel?: string;
  onHover?: (event: { time: number; samples: NearestSample[] }) => void;
}
/** Mounted browser lifecycle. Destroy restores the canvas DOM position and attributes. */
export interface Controller {
  readonly canvas: HTMLCanvasElement;
  readonly chart: Chart;
  readonly result: RenderResult;
  readonly destroyed: boolean;
  update(patch: ChartOptions | Chart): RenderResult;
  destroy(): void;
}
export interface DashboardOptions {
  gap?: number;
  padding?: readonly [number, number, number, number];
  background?: Color;
  cropHeight?: number | null;
}
export interface DashboardPanel {
  chart: Chart;
  caption?: string;
}
export interface DashboardMetadata {
  version: string;
  imageSize: [number, number];
  pixelScale: number;
  panels: {
    position: [number, number];
    caption: string;
    chart: GraphMetadata;
  }[];
}
export interface CounterOptions {
  factor?: number;
  onDecrease?: "gap" | "wrap";
  counterBits?: number | null;
  maxRate?: number | null;
}
export interface AggregateOptions {
  interval?: number;
  method?: "mean" | "min" | "max" | "last" | "sum";
  origin?: number;
  minCoverage?: number;
  expectedStep?: number | null;
  maxBuckets?: number;
}
export interface CSVColumn extends SeriesOptions {
  column: string;
  name?: string;
}
export interface CSVOptions {
  timestampColumn?: string;
  columns?: readonly CSVColumn[];
  maxRows?: number;
  maxBytes?: number;
}
export interface PixelDifference {
  pixels: number;
  exactPixels: number;
  exactRatio: number;
  toleranceRatio: number;
  meanAbsoluteError: number;
  rootMeanSquareError: number;
  maxError: number;
  differenceBox: [number, number, number, number] | null;
}
export interface CompareOptions {
  tolerance?: number;
  box?: readonly [number, number, number, number] | null;
}

/** All chart fields after default merging and normalization. */
export interface ResolvedChartOptions extends Omit<
  Required<ChartOptions>,
  | "series"
  | "timeAxis"
  | "yAxis"
  | "layout"
  | "theme"
  | "fonts"
  | "hRules"
  | "vRules"
> {
  series: readonly Series[];
  timeAxis: ResolvedTimeAxis;
  yAxis: Required<YAxis>;
  layout: ResolvedLayout;
  theme: ResolvedTheme;
  fonts: Required<FontOptions>;
  hRules: readonly ResolvedHRule[];
  vRules: readonly ResolvedVRule[];
}
/** A calendar axis whose timestamp inputs have been converted to Unix seconds. */
export interface ResolvedTimeAxis extends Omit<
  Required<TimeAxis>,
  "start" | "end" | "ticks" | "minorTicks" | "majorTicks"
> {
  start: number | null;
  end: number | null;
  ticks: readonly ResolvedTick[] | null;
  minorTicks: readonly number[] | null;
  majorTicks: readonly number[] | null;
}
/** A label positioned at a numerical Unix timestamp. */
export interface ResolvedTick {
  time: number;
  label: string;
}
/** Complete logical-pixel layout including the legend geometry. */
export interface ResolvedLayout extends Required<Omit<Layout, "legendLayout">> {
  legendLayout: Required<LegendLayout>;
}
type ThemeColorKey =
  | "background"
  | "canvas"
  | "shadeLight"
  | "shadeDark"
  | "text"
  | "minorGrid"
  | "majorGrid"
  | "axis"
  | "arrow"
  | "watermark"
  | "frame";
/** Fully specified theme with parsed RGBA colors. */
export type ResolvedTheme = Required<Omit<Theme, ThemeColorKey>> & {
  [K in ThemeColorKey]: Readonly<RGBA>;
};
/** Normalized horizontal rule in original data units. */
export interface ResolvedHRule {
  value: number;
  color: Readonly<RGBA>;
  width: number;
  dash: readonly [number, number] | null;
}
/** Normalized vertical rule in Unix seconds. */
export interface ResolvedVRule {
  time: number;
  color: Readonly<RGBA>;
  width: number;
  dash: readonly [number, number] | null;
}
/** Recursive readonly view of plain chart configuration and metadata. */
export type DeepReadonly<T> = T extends object
  ? { readonly [K in keyof T]: DeepReadonly<T[K]> }
  : T;
type CompleteChartInput = Omit<
  Required<ChartOptions>,
  "timeAxis" | "yAxis" | "layout" | "theme" | "fonts"
> & {
  timeAxis: Required<TimeAxis>;
  yAxis: Required<YAxis>;
  layout: ResolvedLayout;
  theme: Required<Theme>;
  fonts: Required<FontOptions>;
};
type CompleteSeriesInput = Required<Omit<SeriesInput, "legendValues">> & {
  legendValues: Partial<LegendValues> | null;
};
type ChartConfig = DeepReadonly<ResolvedChartOptions>;
type Point = [number, number];
type Canvas = HTMLCanvasElement | OffscreenCanvas;
type Context2D = CanvasRenderingContext2D | OffscreenCanvasRenderingContext2D;
type FontRole = "title" | "axis" | "unit" | "legend" | "watermark" | "caption";
interface WallParts {
  year: number;
  month: number;
  day: number;
  hour: number;
  minute: number;
  second: number;
}
interface YResolution extends Unit {
  minimum: number;
  maximum: number;
  step: number;
  major: number[];
  minor: number[];
  decimals: number;
}
interface XResolution {
  minor: number[];
  major: number[];
  labels: ResolvedTick[];
  mode: NonNullable<TimeAxis["mode"]>;
}
type MountedCanvas = HTMLCanvasElement & {
  __bamtiGraphController?: BrowserController;
};

/** BamtiGraph library version. */
export const VERSION: string = "0.1.0";
/** Per-operation allocation and enumeration guards. */
export const LIMITS: Readonly<{
  pixels: number;
  layerPixels: number;
  ticks: number;
  series: number;
  samples: number;
  text: number;
}> = Object.freeze({
  pixels: 16000000,
  layerPixels: 40000000,
  ticks: 5000,
  series: 128,
  samples: 2000000,
  text: 4096,
});
const finite = (v: unknown): v is number => Number.isFinite(v);
const round = (v: number): number => Math.floor(v + 0.5);
const clamp = (v: number, a: number, b: number): number =>
  Math.max(a, Math.min(b, v));
const own = (o: object, k: PropertyKey): boolean =>
  Object.prototype.hasOwnProperty.call(o, k);
const missing = (v: unknown): boolean =>
  v === null || v === undefined || (typeof v === "number" && Number.isNaN(v));
function check(ok: unknown, message: string): asserts ok {
  if (!ok) throw new RangeError(message);
}
function text(s: unknown, name: string): string {
  check(
    typeof s === "string" && s.length <= LIMITS.text && !/[\r\n\u0000]/.test(s),
    name + " must be a single-line string.",
  );
  return s;
}
function number(
  v: unknown,
  name: string,
  lo: number = -Infinity,
  hi: number = Infinity,
): number {
  check(
    finite(v) && v >= lo && v <= hi,
    name + " is outside its finite range.",
  );
  return v;
}
function integer(v: unknown, name: string, lo: number, hi: number): number {
  check(Number.isInteger(v), name + " must be an integer.");
  return number(v, name, lo, hi);
}
function record(o: unknown): o is Record<string, unknown> {
  return (
    o !== null &&
    typeof o === "object" &&
    (Object.getPrototypeOf(o) === Object.prototype ||
      Object.getPrototypeOf(o) === null)
  );
}
// Configuration inputs are plain data. Typed views are restored only at this
// recursive ownership boundary; property values remain unknown while copied.
function clone<T>(v: T): T {
  if (Array.isArray(v) || ArrayBuffer.isView(v))
    return Array.from(v as ArrayLike<unknown>, (value) => clone(value)) as T;
  if (v instanceof Date) return new Date(v.getTime()) as T;
  if (record(v)) {
    const r: Record<string, unknown> = {};
    for (const k of Object.keys(v)) {
      check(
        !["__proto__", "prototype", "constructor"].includes(k),
        "Unsafe property name.",
      );
      r[k] = clone(v[k]);
    }
    return r as T;
  }
  return v;
}
function merge<T extends object>(a: T, b: object): T {
  check(record(b), "Options must be a plain object.");
  const r = clone(a) as Record<string, unknown>;
  for (const k of Object.keys(b)) {
    check(
      !["__proto__", "prototype", "constructor"].includes(k),
      "Unsafe property name.",
    );
    const previous = r[k],
      next = b[k];
    r[k] =
      record(previous) && record(next) ? merge(previous, next) : clone(next);
  }
  return r as T;
}
function freeze<T>(o: T): T {
  if (o && typeof o === "object") {
    for (const v of Object.values(o)) freeze(v);
    Object.freeze(o);
  }
  return o;
}
/** Parse a supported color into integer RGBA bytes. */
export function color(value: Color): RGBA {
  if (Array.isArray(value) || ArrayBuffer.isView(value)) {
    check(
      value.length === 3 || value.length === 4,
      "Color needs 3 or 4 channels.",
    );
    const c = Array.from(value);
    c.forEach((v) => integer(v, "Color channel", 0, 255));
    if (c.length === 3) c.push(255);
    return c as RGBA;
  }
  const named: Record<string, string> = {
    black: "#000000",
    white: "#ffffff",
    red: "#ff0000",
    green: "#008000",
    blue: "#0000ff",
    transparent: "#00000000",
  };
  check(
    typeof value === "string",
    "Color must be hexadecimal or an RGB(A) array.",
  );
  let h = named[value.toLowerCase()] || value;
  check(
    /^#(?:[0-9a-f]{3}|[0-9a-f]{4}|[0-9a-f]{6}|[0-9a-f]{8})$/i.test(h),
    "Invalid color: " + value,
  );
  h = h.slice(1);
  if (h.length <= 4) h = [...h].map((c) => c + c).join("");
  if (h.length === 6) h += "ff";
  return [0, 2, 4, 6].map((i) => parseInt(h.slice(i, i + 2), 16)) as RGBA;
}
/** Normalize a timestamp to Unix seconds; naive dates and invalid calendars fail. */
export function epoch(value: Timestamp): number {
  if (value instanceof Date) value = value.getTime() / 1000;
  if (typeof value === "string") {
    const v = value.trim();
    if (/^[+-]?(?:\d+\.?\d*|\.\d+)(?:e[+-]?\d+)?$/i.test(v)) value = Number(v);
    else {
      const m =
        /^(\d{4})-(\d\d)-(\d\d)T(\d\d):(\d\d)(?::(\d\d)(\.\d{1,3})?)?(Z|[+-]\d\d:\d\d)$/i.exec(
          v,
        );
      check(
        m,
        "Datetime strings require ISO 8601 with an explicit offset; numeric timestamps are seconds.",
      );
      const [y, mo, d, h, mi, s] = [
        m[1],
        m[2],
        m[3],
        m[4],
        m[5],
        m[6] || "0",
      ].map(Number);
      check(
        y >= 1 && mo >= 1 && mo <= 12 && d >= 1 && h < 24 && mi < 60 && s < 60,
        "Invalid calendar date.",
      );
      const test = new Date(0);
      test.setUTCFullYear(y, mo - 1, d);
      test.setUTCHours(h, mi, s, 0);
      check(
        test.getUTCMonth() === mo - 1 && test.getUTCDate() === d,
        "Invalid calendar date.",
      );
      if (m[8].toUpperCase() !== "Z")
        check(
          Number(m[8].slice(1, 3)) <= 23 && Number(m[8].slice(4)) <= 59,
          "Invalid timezone offset.",
        );
      value = Date.parse(v) / 1000;
    }
  }
  return number(value, "Timestamp", -62135596800, 253402300799.999);
}
function samples(timestamps: ArrayLike<Timestamp>, values: Values): Samples {
  check(
    (Array.isArray(timestamps) || ArrayBuffer.isView(timestamps)) &&
      (Array.isArray(values) || ArrayBuffer.isView(values)),
    "Timestamps and values must be arrays.",
  );
  check(
    timestamps.length === values.length && timestamps.length <= LIMITS.samples,
    "Sample lengths must match and stay within the sample limit.",
  );
  const ts = Array.from(timestamps, epoch),
    vs = Array.from(values, (v, i) =>
      missing(v) ? NaN : number(v, "Value at " + i),
    );
  for (let i = 1; i < ts.length; i++)
    check(
      ts[i] > ts[i - 1],
      "Timestamps must be strictly increasing (index " + i + ").",
    );
  return { timestamps: ts, values: vs };
}
const DEFAULTS: CompleteChartInput = freeze<CompleteChartInput>({
  title: "",
  verticalLabel: "",
  watermark: "",
  series: [],
  timeAxis: {
    start: null,
    end: null,
    mode: "auto",
    timezone: "UTC",
    minorSeconds: null,
    majorSeconds: null,
    labelSeconds: null,
    labelFormat: null,
    ticks: null,
    majorTicks: null,
    minorTicks: null,
    labelOffsetSeconds: 0,
  },
  yAxis: {
    minimum: 0,
    maximum: null,
    majorStep: null,
    minorDivisions: 5,
    base: 1000,
    scaleFactor: null,
    suffix: null,
    decimals: null,
    legendDecimals: 2,
    showZeroSuffix: false,
  },
  layout: {
    width: 595,
    plotHeight: 122,
    left: 64,
    right: 31,
    top: 34,
    titleY: 8,
    titleOffsetX: 27,
    unitX: 5,
    xLabelGap: 5,
    yLabelGap: 6,
    legendGap: 20,
    legendRowHeight: 14,
    legendBottom: 7,
    legend: "reference",
    antialias: 4,
    pixelScale: 1,
    legendLayout: {
      nameX: 30,
      swatchX: 15,
      swatchWidth: 9,
      swatchHeight: 10,
      referenceWidth: 595,
      autoScaleColumns: true,
      compact: [
        [102, 228],
        [244, 370],
        [386, 512],
      ],
      expanded: [
        [124, 250],
        [289, 415],
        [454, 580],
      ],
      aligned: [
        [102, 250],
        [267, 415],
        [432, 580],
      ],
    },
  },
  theme: {
    background: "#f3f3f3",
    canvas: "#ffffff",
    shadeLight: "#cfcfcf",
    shadeDark: "#9e9e9e",
    text: "#000000",
    minorGrid: "#8f8f8f3c",
    majorGrid: "#df4f4f3c",
    axis: "#777777",
    arrow: "#7f1f1f",
    watermark: "#aaaaaa",
    frame: "#000000",
    gridFront: true,
    gridDash: [1, 1],
    titleSize: 14,
    axisSize: 11,
    unitSize: 10,
    legendSize: 11,
    watermarkSize: 8,
    captionSize: 11,
    titleAdvance: 8,
    axisAdvance: 6,
    legendAdvance: 7,
  },
  fonts: {
    mode: "system",
    family: '"DejaVu Sans Mono", "Liberation Mono", Consolas, monospace',
    titleFamily: null,
    unitFamily: null,
    captionFamily: "Arial, sans-serif",
    strictGlyphs: true,
  },
  legendLabels: ["Current:", "Average:", "Maximum:"],
  missingLabel: "NaN",
  hRules: [],
  vRules: [],
});
/** Copy, validate, and deeply freeze an independently timestamped series. */
export function series(
  name: string,
  timestamps: ArrayLike<Timestamp>,
  values: Values,
  options: SeriesOptions = {},
): Series {
  const s = merge<CompleteSeriesInput>(
    {
      name,
      timestamps,
      values,
      kind: "line",
      color: "#0000cc",
      lineWidth: 0.7,
      outline: null,
      baseline: 0,
      interpolation: "linear",
      gapAfter: 0,
      legendValues: null,
    },
    options,
  );
  text(s.name, "Series name");
  const data = samples(s.timestamps, s.values);
  s.timestamps = data.timestamps;
  s.values = data.values;
  check(["line", "area"].includes(s.kind), "Series kind must be line or area.");
  check(
    ["linear", "step-post"].includes(s.interpolation),
    "Interpolation must be linear or step-post.",
  );
  number(s.lineWidth, "Line width", 0.01, 128);
  number(s.baseline, "Baseline");
  number(s.gapAfter, "Gap threshold", 0);
  s.color = color(s.color);
  if (s.outline !== null) s.outline = color(s.outline);
  if (s.legendValues !== null) {
    check(record(s.legendValues), "Legend overrides must be an object.");
    for (const k of ["current", "average", "maximum"] as const)
      s.legendValues[k] = missing(s.legendValues[k])
        ? null
        : number(s.legendValues[k], "Legend override");
  }
  return freeze(s as Series);
}
/** Create a series with fixed elapsed-second spacing and copied observations. */
export function regularSeries(
  name: string,
  values: Values,
  start: Timestamp,
  step: number = 300,
  options: SeriesOptions = {},
): Series {
  start = epoch(start);
  number(step, "Step", Number.MIN_VALUE);
  const first = start;
  return series(
    name,
    Array.from(values, (_, i) => first + i * step),
    values,
    options,
  );
}
function modeAxis(
  mode: NonNullable<TimeAxis["mode"]>,
  options: TimeAxis | string = {},
): TimeAxis {
  return merge(
    DEFAULTS.timeAxis,
    merge(
      { mode },
      typeof options === "string" ? { timezone: options } : options,
    ),
  );
}
/** Daily tick presentation, without resampling observations. */
export function daily(options: TimeAxis | string = {}): TimeAxis {
  return modeAxis("daily", options);
}
/** Weekly calendar-noon labels, without resampling observations. */
export function weekly(options: TimeAxis | string = {}): TimeAxis {
  return modeAxis("weekly", options);
}
/** Monthly tick presentation, without resampling observations. */
export function monthly(options: TimeAxis | string = {}): TimeAxis {
  return modeAxis("monthly", options);
}
/** Real calendar-month ticks, without resampling observations. */
export function yearly(options: TimeAxis | string = {}): TimeAxis {
  return modeAxis("yearly", options);
}
function dimensions(c: {
  layout: DeepReadonly<ResolvedLayout>;
  series: readonly unknown[];
}): [number, number] {
  const l = c.layout;
  return [
    l.width,
    l.top +
      l.plotHeight +
      (l.legend === "none"
        ? 18
        : l.legendGap +
          Math.max(1, c.series.length) * l.legendRowHeight +
          l.legendBottom),
  ];
}
function dash(d: readonly [number, number] | null): void {
  if (d === null) return;
  check(Array.isArray(d) && d.length === 2, "Dash must be null or [on, off].");
  d.forEach((v) => integer(v, "Dash interval", 1, 16384));
}
function normalize(options: ChartOptions): ChartConfig {
  const c = merge(DEFAULTS, options);
  check(
    Array.isArray(c.series) && c.series.length <= LIMITS.series,
    "Invalid series count.",
  );
  c.series = c.series.map((s) => series(s.name, s.timestamps, s.values, s));
  for (const k of [
    "title",
    "verticalLabel",
    "watermark",
    "missingLabel",
  ] as const)
    text(c[k], k);
  check(
    Array.isArray(c.legendLabels) && c.legendLabels.length === 3,
    "Three statistic labels are required.",
  );
  c.legendLabels.forEach((v) => text(v, "Statistic label"));
  const a = c.timeAxis,
    l = c.layout,
    y = c.yAxis,
    t = c.theme,
    f = c.fonts;
  check(
    ["auto", "daily", "weekly", "monthly", "yearly", "custom"].includes(a.mode),
    "Unknown time axis mode.",
  );
  text(a.timezone, "Timezone");
  zoneFormatter(a.timezone);
  for (const k of ["start", "end"] as const)
    if (a[k] !== null) a[k] = epoch(a[k]);
  if (a.start !== null && a.end !== null)
    check(
      (a.end as number) > (a.start as number),
      "Time end must exceed start.",
    );
  for (const k of ["minorSeconds", "majorSeconds", "labelSeconds"] as const)
    if (a[k] !== null) number(a[k], k, 0.001);
  number(a.labelOffsetSeconds, "Label offset", -31622400, 31622400);
  if (a.labelFormat !== null) {
    text(a.labelFormat, "Label format");
    formatTime(0, "UTC", a.labelFormat);
  }
  if (a.ticks !== null) {
    check(
      Array.isArray(a.ticks) && a.ticks.length <= LIMITS.ticks,
      "Too many or invalid ticks.",
    );
    const ticks = a.ticks.map((v) => ({
      time: epoch(v.time),
      label: text(v.label, "Tick label"),
    }));
    for (let i = 1; i < ticks.length; i++)
      check(
        ticks[i].time > ticks[i - 1].time,
        "Explicit ticks must be strictly increasing.",
      );
    a.ticks = ticks;
  }
  for (const k of ["minorTicks", "majorTicks"] as const)
    if (a[k] !== null) {
      const input = a[k];
      check(
        Array.isArray(input) && input.length <= LIMITS.ticks,
        "Too many or invalid ticks.",
      );
      const ticks = input.map(epoch);
      for (let i = 1; i < ticks.length; i++)
        check(
          ticks[i] > ticks[i - 1],
          "Explicit ticks must be strictly increasing.",
        );
      a[k] = ticks;
    }
  for (const k of ["minimum", "maximum", "majorStep", "scaleFactor"] as const)
    if (y[k] !== null) number(y[k], k);
  for (const k of ["majorStep", "scaleFactor"] as const)
    if (y[k] !== null) check(y[k] > 0, k + " must be positive.");
  if (y.minimum !== null && y.maximum !== null)
    check(y.maximum > y.minimum, "Y maximum must exceed minimum.");
  check(y.base === 1000 || y.base === 1024, "Unit base must be 1000 or 1024.");
  integer(y.minorDivisions, "Minor divisions", 1, 100);
  integer(y.legendDecimals, "Legend decimals", 0, 12);
  if (y.decimals !== null) integer(y.decimals, "Decimals", 0, 12);
  if (y.suffix !== null) text(y.suffix, "Suffix");
  integer(l.width, "Width", 400, 8192);
  integer(l.plotHeight, "Plot height", 30, 4096);
  for (const k of [
    "left",
    "right",
    "top",
    "titleY",
    "unitX",
    "xLabelGap",
    "yLabelGap",
    "legendGap",
    "legendRowHeight",
    "legendBottom",
  ] as const)
    integer(l[k], k, 0, 16384);
  check(
    l.left >= 20 &&
      l.right >= 12 &&
      l.top >= 12 &&
      l.width - l.left - l.right >= 100,
    "Insufficient plot margins.",
  );
  number(l.titleOffsetX, "Title offset", -8192, 8192);
  check(
    ["reference", "aligned", "none"].includes(l.legend),
    "Invalid legend mode.",
  );
  if (l.legend !== "none")
    check(
      l.legendGap >= 14 && l.legendRowHeight >= 10,
      "Legend spacing is insufficient.",
    );
  integer(l.antialias, "Antialias", 1, 8);
  integer(l.pixelScale, "Pixel scale", 1, 8);
  const ll = l.legendLayout;
  for (const k of [
    "nameX",
    "swatchX",
    "swatchWidth",
    "swatchHeight",
    "referenceWidth",
  ] as const)
    integer(ll[k], k, 0, 16384);
  check(
    ll.swatchWidth >= 3 && ll.swatchHeight >= 3 && ll.referenceWidth > ll.nameX,
    "Invalid legend geometry.",
  );
  for (const key of ["compact", "expanded", "aligned"] as const) {
    const cols = ll[key];
    check(Array.isArray(cols) && cols.length === 3, "Invalid legend columns.");
    let last = ll.nameX;
    for (const p of cols) {
      check(
        Array.isArray(p) &&
          p.length === 2 &&
          finite(p[0]) &&
          finite(p[1]) &&
          last < p[0] &&
          p[0] < p[1] &&
          p[1] < ll.referenceWidth,
        "Legend columns must not overlap.",
      );
      last = p[1];
    }
  }
  for (const k of [
    "background",
    "canvas",
    "shadeLight",
    "shadeDark",
    "text",
    "minorGrid",
    "majorGrid",
    "axis",
    "arrow",
    "watermark",
    "frame",
  ] as const)
    t[k] = color(t[k]);
  check(
    t.background[3] === 255 && t.canvas[3] === 255,
    "Background and canvas must be opaque.",
  );
  dash(t.gridDash);
  check(t.gridDash !== null, "Grid dash is required.");
  for (const k of [
    "titleSize",
    "axisSize",
    "unitSize",
    "legendSize",
    "watermarkSize",
    "captionSize",
    "titleAdvance",
    "axisAdvance",
    "legendAdvance",
  ] as const)
    number(t[k], k, 1, 128);
  check(
    ["system", "bitmap"].includes(f.mode),
    "Font mode must be system or bitmap.",
  );
  text(f.family, "Font family");
  for (const k of ["titleFamily", "unitFamily", "captionFamily"] as const)
    if (f[k] !== null) text(f[k], k);
  function rule<R extends HRule | VRule>(
    input: R,
    field: "time" | "value",
  ): R & Required<RuleStyle> {
    const r = merge(
      { color: "#990000", width: 1, dash: [3, 2] },
      input,
    ) as unknown as R & Required<RuleStyle>;
    if (field === "time") (r as VRule).time = epoch((r as VRule).time);
    else (r as HRule).value = number((r as HRule).value, "Rule value");
    number(r.width, "Rule width", 0.1, 128);
    dash(r.dash);
    r.color = color(r.color);
    return r;
  }
  check(
    Array.isArray(c.hRules) && c.hRules.length <= 1000,
    "Invalid rule count.",
  );
  c.hRules = c.hRules.map((r) => rule(r, "value"));
  check(
    Array.isArray(c.vRules) && c.vRules.length <= 1000,
    "Invalid rule count.",
  );
  c.vRules = c.vRules.map((r) => rule(r, "time"));
  const [w, h] = dimensions(c);
  check(
    w * h * l.pixelScale ** 2 <= LIMITS.pixels,
    "Output allocation exceeds the pixel limit.",
  );
  check(
    (l.width - l.left - l.right + 1) * (l.plotHeight + 1) * l.antialias ** 2 <=
      LIMITS.layerPixels,
    "Supersampled layer exceeds the pixel limit.",
  );
  // All timestamp, color, series and rule fields have now been normalized.
  return freeze(c as ResolvedChartOptions);
}
// A straight-alpha pixel surface: data geometry never depends on a browser path rasterizer.
class Surface implements RGBAImage {
  readonly width: number;
  readonly height: number;
  readonly data: Uint8ClampedArray;
  constructor(
    width: number,
    height: number,
    fill: Readonly<RGBA> | null = null,
  ) {
    integer(width, "Image width", 1, 65536);
    integer(height, "Image height", 1, 65536);
    check(width * height <= LIMITS.layerPixels, "Image is too large.");
    this.width = width;
    this.height = height;
    this.data = new Uint8ClampedArray(width * height * 4);
    if (fill) this.rect(0, 0, width, height, fill);
  }
  pixel(x: number, y: number, c: Readonly<RGBA>, blend: boolean = false): void {
    if (x < 0 || y < 0 || x >= this.width || y >= this.height) return;
    const i = (y * this.width + x) * 4,
      d = this.data;
    if (!blend || c[3] === 255) {
      d[i] = c[0];
      d[i + 1] = c[1];
      d[i + 2] = c[2];
      d[i + 3] = c[3];
      return;
    }
    if (c[3] === 0) return;
    const sa = c[3],
      da = d[i + 3],
      alpha = sa * 255 + da * (255 - sa);
    for (let k = 0; k < 3; k++)
      d[i + k] = Math.floor(
        (c[k] * sa * 255 + d[i + k] * da * (255 - sa) + alpha / 2) / alpha,
      );
    d[i + 3] = Math.floor((alpha + 127) / 255);
  }
  rect(
    x0: number,
    y0: number,
    x1: number,
    y1: number,
    c: Readonly<RGBA>,
  ): void {
    x0 = Math.max(0, Math.ceil(x0));
    y0 = Math.max(0, Math.ceil(y0));
    x1 = Math.min(this.width, Math.ceil(x1));
    y1 = Math.min(this.height, Math.ceil(y1));
    for (let y = y0; y < y1; y++)
      for (let x = x0, i = (y * this.width + x0) * 4; x < x1; x++, i += 4) {
        this.data[i] = c[0];
        this.data[i + 1] = c[1];
        this.data[i + 2] = c[2];
        this.data[i + 3] = c[3];
      }
  }
  over(src: RGBAImage, dx: number = 0, dy: number = 0): void {
    const p: RGBA = [0, 0, 0, 0];
    for (
      let y = Math.max(0, -dy);
      y < Math.min(src.height, this.height - dy);
      y++
    )
      for (
        let x = Math.max(0, -dx);
        x < Math.min(src.width, this.width - dx);
        x++
      ) {
        const i = (y * src.width + x) * 4;
        if (src.data[i + 3]) {
          p[0] = src.data[i];
          p[1] = src.data[i + 1];
          p[2] = src.data[i + 2];
          p[3] = src.data[i + 3];
          this.pixel(x + dx, y + dy, p, true);
        }
      }
  }
  polygon(points: readonly Point[], c: Readonly<RGBA>): void {
    if (points.length < 3) return;
    let lo = Infinity,
      hi = -Infinity;
    for (const p of points) {
      lo = Math.min(lo, p[1]);
      hi = Math.max(hi, p[1]);
    }
    for (
      let y = Math.max(0, Math.ceil(lo));
      y <= Math.min(this.height - 1, Math.floor(hi));
      y++
    ) {
      const xs = [];
      let a = points[points.length - 1];
      for (const b of points) {
        if (a[1] === b[1]) {
          if (y === a[1])
            this.rect(
              Math.ceil(Math.min(a[0], b[0])),
              y,
              Math.floor(Math.max(a[0], b[0])) + 1,
              y + 1,
              c,
            );
        } else if ((a[1] <= y && y < b[1]) || (b[1] <= y && y < a[1]))
          xs.push(a[0] + ((y - a[1]) / (b[1] - a[1])) * (b[0] - a[0]));
        a = b;
      }
      xs.sort((a, b) => a - b);
      for (let i = 0; i + 1 < xs.length; i += 2)
        this.rect(
          Math.ceil(xs[i] - 1e-9),
          y,
          Math.floor(xs[i + 1] + 1e-9) + 1,
          y + 1,
          c,
        );
    }
  }
  line(a: Point, b: Point, c: Readonly<RGBA>, width: number = 1): void {
    let x = round(a[0]),
      y = round(a[1]),
      x1 = round(b[0]),
      y1 = round(b[1]);
    width = Math.max(1, round(width));
    if (width > 1) {
      const r = (width - 1) / 2;
      if (x === x1) {
        this.rect(
          x - Math.floor(width / 2),
          Math.min(y, y1),
          x + Math.floor((width - 1) / 2) + 1,
          Math.max(y, y1) + 1,
          c,
        );
        return;
      }
      if (y === y1) {
        this.rect(
          Math.min(x, x1),
          y - Math.floor(width / 2),
          Math.max(x, x1) + 1,
          y + Math.floor((width - 1) / 2) + 1,
          c,
        );
        return;
      }
      const len = Math.hypot(x1 - x, y1 - y),
        ox = (-(y1 - y) / len) * r,
        oy = ((x1 - x) / len) * r;
      this.polygon(
        [
          [x + ox, y + oy],
          [x1 + ox, y1 + oy],
          [x1 - ox, y1 - oy],
          [x - ox, y - oy],
        ].map((p): Point => [round(p[0]), round(p[1])]),
        c,
      );
      return;
    }
    const dx = Math.abs(x1 - x),
      dy = -Math.abs(y1 - y),
      sx = x < x1 ? 1 : -1,
      sy = y < y1 ? 1 : -1;
    let err = dx + dy;
    for (;;) {
      this.pixel(x, y, c);
      if (x === x1 && y === y1) break;
      const e = 2 * err;
      if (e >= dy) {
        err += dy;
        x += sx;
      }
      if (e <= dx) {
        err += dx;
        y += sy;
      }
    }
  }
  dashed(
    a: Point,
    b: Point,
    c: Readonly<RGBA>,
    pattern: readonly [number, number] | null = [1, 1],
    width: number = 1,
  ): void {
    if (pattern === null) {
      this.line(a, b, c, width);
      return;
    }
    const len = Math.hypot(b[0] - a[0], b[1] - a[1]);
    if (!len) {
      this.pixel(round(a[0]), round(a[1]), c);
      return;
    }
    const dx = (b[0] - a[0]) / len,
      dy = (b[1] - a[1]) / len;
    for (let s = 0; s <= Math.ceil(len); s += pattern[0] + pattern[1]) {
      const e = Math.min(len, s + pattern[0] - 1);
      this.line(
        [a[0] + dx * s, a[1] + dy * s],
        [a[0] + dx * e, a[1] + dy * e],
        c,
        width,
      );
    }
  }
  circle(x: number, y: number, r: number, c: Readonly<RGBA>): void {
    for (
      let yy = Math.max(0, Math.floor(y - r));
      yy <= Math.min(this.height - 1, Math.ceil(y + r));
      yy++
    )
      for (
        let xx = Math.max(0, Math.floor(x - r));
        xx <= Math.min(this.width - 1, Math.ceil(x + r));
        xx++
      )
        if ((xx - x) ** 2 + (yy - y) ** 2 <= r * r) this.pixel(xx, yy, c);
  }
  down(scale: number): Surface {
    if (scale === 1) return this;
    const out = new Surface(this.width / scale, this.height / scale),
      p: RGBA = [0, 0, 0, 0],
      n = scale * scale;
    for (let y = 0; y < out.height; y++)
      for (let x = 0; x < out.width; x++) {
        let a = 0,
          r = 0,
          g = 0,
          b = 0;
        for (let sy = 0; sy < scale; sy++)
          for (let sx = 0; sx < scale; sx++) {
            const i = ((y * scale + sy) * this.width + x * scale + sx) * 4,
              ca = this.data[i + 3];
            a += ca;
            r += this.data[i] * ca;
            g += this.data[i + 1] * ca;
            b += this.data[i + 2] * ca;
          }
        if (a) {
          p[0] = Math.floor(r / a + 0.5);
          p[1] = Math.floor(g / a + 0.5);
          p[2] = Math.floor(b / a + 0.5);
          p[3] = Math.floor(a / n + 0.5);
          out.pixel(x, y, p);
        }
      }
    return out;
  }
  scale(n: number): Surface {
    if (n === 1) return this;
    const out = new Surface(this.width * n, this.height * n);
    for (let y = 0; y < out.height; y++)
      for (let x = 0; x < out.width; x++) {
        const s = (Math.floor(y / n) * this.width + Math.floor(x / n)) * 4,
          i = (y * out.width + x) * 4;
        out.data[i] = this.data[s];
        out.data[i + 1] = this.data[s + 1];
        out.data[i + 2] = this.data[s + 2];
        out.data[i + 3] = this.data[s + 3];
      }
    return out;
  }
}
function clipLine(
  a: Point,
  b: Point,
  w: number,
  h: number,
): [Point, Point] | null {
  let lo = 0,
    hi = 1;
  const dx = b[0] - a[0],
    dy = b[1] - a[1];
  for (const [p, q] of [
    [-dx, a[0]],
    [dx, w - a[0]],
    [-dy, a[1]],
    [dy, h - a[1]],
  ]) {
    if (p === 0) {
      if (q < 0) return null;
    } else {
      const u = q / p;
      if (p < 0) lo = Math.max(lo, u);
      else hi = Math.min(hi, u);
      if (lo > hi) return null;
    }
  }
  return [
    [a[0] + lo * dx, a[1] + lo * dy],
    [a[0] + hi * dx, a[1] + hi * dy],
  ];
}
function clipPolygon(points: Point[], w: number, h: number): Point[] {
  let ps = points;
  for (const [axis, bound, greater] of [
    [0, 0, true],
    [0, w, false],
    [1, 0, true],
    [1, h, false],
  ] as const) {
    if (!ps.length) break;
    const out: Point[] = [],
      inside = (p: Point): boolean =>
        greater ? p[axis] >= bound : p[axis] <= bound;
    let a = ps[ps.length - 1],
      ai = inside(a);
    for (const b of ps) {
      const bi = inside(b);
      if (ai !== bi) {
        const q = (bound - a[axis]) / (b[axis] - a[axis]);
        const p: Point = [a[0] + q * (b[0] - a[0]), a[1] + q * (b[1] - a[1])];
        p[axis] = bound;
        out.push(p);
      }
      if (bi) out.push(b);
      a = b;
      ai = bi;
    }
    ps = out;
  }
  return ps;
}
function lowerBound(a: readonly number[], x: number): number {
  let l = 0,
    r = a.length;
  while (l < r) {
    const m = (l + r) >>> 1;
    if (a[m] < x) l = m + 1;
    else r = m;
  }
  return l;
}
function upperBound(a: readonly number[], x: number): number {
  let l = 0,
    r = a.length;
  while (l < r) {
    const m = (l + r) >>> 1;
    if (a[m] <= x) l = m + 1;
    else r = m;
  }
  return l;
}
function visibleRuns(s: Series, start: number, end: number): Point[][] {
  const ts = s.timestamps,
    vs = s.values,
    result: Point[][] = [];
  let run: Point[] = [];
  const push = () => {
    if (run.length) {
      result.push(run);
      run = [];
    }
  };
  const left = Math.max(0, lowerBound(ts, start) - 1),
    right = Math.min(ts.length, upperBound(ts, end) + 1);
  for (let i = left; i < right; i++) {
    if (!finite(vs[i])) {
      push();
      continue;
    }
    if (i > left && s.gapAfter > 0 && ts[i] - ts[i - 1] > s.gapAfter) push();
    if (run.length && s.interpolation === "step-post")
      run.push([ts[i], run[run.length - 1][1]]);
    run.push([ts[i], vs[i]]);
  }
  push();
  const clipped: Point[][] = [];
  for (const r of result) {
    const out: Point[] = [];
    if (r.length === 1) {
      if (r[0][0] >= start && r[0][0] <= end) out.push(r[0]);
    }
    for (let i = 1; i < r.length; i++) {
      const a = r[i - 1],
        b = r[i];
      if (b[0] < start || a[0] > end) continue;
      const interp = (x: number): number =>
        a[1] * (1 - (x - a[0]) / (b[0] - a[0])) +
        b[1] * ((x - a[0]) / (b[0] - a[0]));
      const p: Point = a[0] < start ? [start, interp(start)] : a,
        q: Point = b[0] > end ? [end, interp(end)] : b;
      if (
        !out.length ||
        out[out.length - 1][0] !== p[0] ||
        out[out.length - 1][1] !== p[1]
      )
        out.push(p);
      out.push(q);
    }
    if (out.length) clipped.push(out);
  }
  return clipped;
}
function decimate(
  ps: Point[],
  start: number,
  end: number,
  width: number,
): Point[] {
  if (ps.length <= width * 4) return ps;
  const out = [];
  let pos = 0;
  while (pos < ps.length) {
    const col = Math.floor(((ps[pos][0] - start) / (end - start)) * width),
      first = pos;
    let mn = pos,
      mx = pos;
    while (
      pos + 1 < ps.length &&
      Math.floor(((ps[pos + 1][0] - start) / (end - start)) * width) === col
    ) {
      pos++;
      if (ps[pos][1] < ps[mn][1]) mn = pos;
      if (ps[pos][1] > ps[mx][1]) mx = pos;
    }
    for (const i of [...new Set([first, mn, mx, pos])].sort((a, b) => a - b))
      out.push(ps[i]);
    pos++;
  }
  return out;
}
function stableMean(values: readonly number[]): number | null {
  if (!values.length) return null;
  let max = 0;
  for (const v of values) max = Math.max(max, Math.abs(v));
  if (max === 0) return 0;
  let sum = 0,
    c = 0;
  for (const v of values) {
    const a = v / max - c,
      t = sum + a;
    c = t - sum - a;
    sum = t;
  }
  return clamp(sum / values.length, -1, 1) * max;
}
function statistics(s: Series, start: number, end: number): Statistics {
  const a = lowerBound(s.timestamps, start),
    b = upperBound(s.timestamps, end),
    values = [];
  let mn = Infinity,
    mx = -Infinity,
    miss = 0;
  for (let i = a; i < b; i++) {
    const v = s.values[i];
    if (!finite(v)) miss++;
    else {
      values.push(v);
      mn = Math.min(mn, v);
      mx = Math.max(mx, v);
    }
  }
  return {
    name: s.name,
    current: b > a && finite(s.values[b - 1]) ? s.values[b - 1] : null,
    average: stableMean(values),
    maximum: values.length ? mx : null,
    minimum: values.length ? mn : null,
    count: values.length,
    missing: miss,
    displayOverride: s.legendValues,
  };
}
// Numeric units and ticks. All limits and data use original (unscaled) units.
function multiples(lo: number, hi: number, step: number): number[] {
  check(
    finite(step) &&
      step > 0 &&
      finite((hi - lo) / step) &&
      (hi - lo) / step <= LIMITS.ticks,
    "Too many ticks or invalid step.",
  );
  const a = Math.ceil(lo / step - 1e-10),
    b = Math.floor(hi / step + 1e-10);
  check(
    Math.abs(a) < 9e15 && Math.abs(b) < 9e15 && b - a <= LIMITS.ticks,
    "Tick precision limit exceeded.",
  );
  return Array.from(
    { length: Math.max(0, b - a + 1) },
    (_, i) => (a + i) * step || 0,
  );
}
function nice(v: number): number {
  check(finite(v) && v >= 1e-300, "Unsupported numeric axis span.");
  const p = 10 ** Math.floor(Math.log10(v));
  for (const m of [1, 2, 5, 10]) if (v <= m * p * (1 + 1e-12)) return m * p;
  return 10 * p;
}
function unitFor(v: number, base: 1000 | 1024): Unit {
  if (!v) return { factor: 1, suffix: "" };
  let i = Math.floor(Math.log(Math.abs(v)) / Math.log(base) + 1e-12);
  i = clamp(i, base === 1024 ? 0 : -8, 8);
  return {
    factor: base ** i,
    suffix: (base === 1024
      ? ["", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi", "Yi"]
      : [
          "y",
          "z",
          "a",
          "f",
          "p",
          "n",
          "u",
          "m",
          "",
          "k",
          "M",
          "G",
          "T",
          "P",
          "E",
          "Z",
          "Y",
        ])[base === 1024 ? i : i + 8],
  };
}
function resolveY(
  a: Readonly<Required<YAxis>>,
  loData: number,
  hiData: number,
): YResolution {
  let lo = a.minimum === null ? Math.min(0, loData) : a.minimum,
    hi = a.maximum === null ? Math.max(0, hiData) : a.maximum;
  if (a.maximum === null && hi <= lo)
    hi = lo + Math.max(Math.abs(lo) * 0.05, 1);
  if (a.minimum === null && lo >= hi)
    lo = hi - Math.max(Math.abs(hi) * 0.05, 1);
  const span = hi - lo;
  check(finite(span) && span > 0, "Invalid Y span.");
  const af =
    a.scaleFactor === null
      ? unitFor(Math.max(Math.abs(lo), Math.abs(hi)), a.base).factor
      : a.scaleFactor;
  const step = a.majorStep === null ? nice(span / af / 5) * af : a.majorStep,
    q = step / a.minorDivisions;
  check(finite(q) && q > 0, "Invalid Y quantum.");
  if (a.minimum === null && lo < 0) lo = Math.floor((lo - span * 0.02) / q) * q;
  if (a.maximum === null) hi = Math.ceil((hi + span * 0.02) / q) * q;
  check(finite(hi - lo) && hi > lo, "Degenerate Y range.");
  const major = multiples(lo, hi, step),
    minor = multiples(lo, hi, q).filter(
      (v) => Math.abs(v / step - round(v / step)) > 1e-8,
    );
  const unit = unitFor(Math.max(Math.abs(lo), Math.abs(hi)), a.base);
  if (a.scaleFactor !== null) {
    unit.factor = a.scaleFactor;
    if (a.suffix === null) unit.suffix = "";
  }
  if (a.suffix !== null) unit.suffix = a.suffix;
  let decimals = a.decimals;
  if (decimals === null) {
    const s = step / unit.factor;
    decimals = 9;
    for (let d = 0; d < 10; d++)
      if (
        Math.abs(s - round(s * 10 ** d) / 10 ** d) <=
        Math.max(1e-10, Math.abs(s) * 1e-9)
      ) {
        decimals = d;
        break;
      }
  }
  return { minimum: lo, maximum: hi, step, major, minor, ...unit, decimals };
}
/** Format a value using an already validated common unit and decimal precision. */
export function formatValue(
  value: number | null,
  unit: Unit,
  decimals: number = 2,
  missingText: string = "NaN",
): string {
  if (!finite(value)) return missingText;
  let v = value / unit.factor;
  if (Math.abs(v) < 0.5 * 10 ** -decimals) v = 0;
  return v.toFixed(decimals) + (unit.suffix ? " " + unit.suffix : "");
}
const formatterCache = new Map<string, Intl.DateTimeFormat>();
function zoneFormatter(zone: string): Intl.DateTimeFormat {
  if (formatterCache.has(zone)) return formatterCache.get(zone)!;
  let f;
  try {
    f = new Intl.DateTimeFormat("en-GB-u-ca-gregory-nu-latn", {
      timeZone: zone,
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hourCycle: "h23",
    });
  } catch (e) {
    throw new RangeError("Unsupported time zone: " + zone);
  }
  if (formatterCache.size >= 32)
    formatterCache.delete(formatterCache.keys().next().value!);
  formatterCache.set(zone, f);
  return f;
}
function utcFromParts(
  y: number,
  m: number,
  d: number,
  h: number = 0,
  mi: number = 0,
  s: number = 0,
): number {
  const t = new Date(0);
  t.setUTCFullYear(y, m - 1, d);
  t.setUTCHours(h, mi, s, 0);
  return t.getTime() / 1000;
}
function wallParts(t: number, zone: string): WallParts {
  if (zone === "UTC") {
    const d = new Date(t * 1000);
    return {
      year: d.getUTCFullYear(),
      month: d.getUTCMonth() + 1,
      day: d.getUTCDate(),
      hour: d.getUTCHours(),
      minute: d.getUTCMinutes(),
      second: d.getUTCSeconds(),
    };
  }
  const r = {} as WallParts;
  for (const p of zoneFormatter(zone).formatToParts(new Date(t * 1000)))
    if (p.type !== "literal") r[p.type as keyof WallParts] = Number(p.value);
  if (r.hour === 24) r.hour = 0;
  return r;
}
function wallEpoch(t: number, zone: string): number {
  const p = wallParts(t, zone);
  return (
    utcFromParts(p.year, p.month, p.day, p.hour, p.minute, p.second) +
    (t - Math.floor(t))
  );
}
function offsetAt(t: number, zone: string): number {
  return Math.round(wallEpoch(t, zone) - t);
}
function localCandidates(w: number, zone: string): number[] {
  if (zone === "UTC") return [w];
  const offsets = new Set(
      [-172800, -86400, 0, 86400, 172800].map((d) => offsetAt(w + d, zone)),
    ),
    out = [];
  for (const off of offsets) {
    const t = w - off;
    if (Math.abs(wallEpoch(t, zone) - w) < 0.001) out.push(t);
  }
  return out.sort((a, b) => a - b);
}
function wallTicks(
  start: number,
  end: number,
  step: number,
  zone: string,
): number[] {
  if (zone === "UTC") return multiples(start, end, step);
  const sa = wallEpoch(start, zone),
    sb = wallEpoch(end, zone),
    a = Math.floor(Math.min(sa, sb) / step) - 2,
    b = Math.ceil(Math.max(sa, sb) / step) + 2;
  check(
    finite(b - a) &&
      b - a <= LIMITS.ticks &&
      Math.abs(a) < 9e15 &&
      Math.abs(b) < 9e15,
    "Too many time ticks; increase the interval.",
  );
  // One offset set for a tick range; 12-hour probes plus endpoint probes preserve ordinary IANA folds.
  const offsets = new Set<number>();
  const probe = Math.max(43200, (end - start) / 4096);
  for (let t = start - 172800; t <= end + 172800; t += probe)
    offsets.add(offsetAt(t, zone));
  offsets.add(offsetAt(end, zone));
  const out = new Set<number>();
  for (let i = 0; i <= b - a; i++) {
    const w = (a + i) * step;
    for (const off of offsets) {
      const t = w - off;
      if (t >= start && t <= end && Math.abs(wallEpoch(t, zone) - w) < 0.001)
        out.add(t);
    }
  }
  return [...out].sort((a, b) => a - b);
}
function monthTicks(
  start: number,
  end: number,
  zone: string,
  stride: number = 1,
): number[] {
  const a = wallParts(start, zone),
    b = wallParts(end, zone),
    first = Math.floor((a.year * 12 + a.month - 1) / stride) * stride,
    last = b.year * 12 + b.month - 1 + stride;
  check((last - first) / stride <= LIMITS.ticks, "Too many calendar ticks.");
  const out = [];
  for (let i = first; i <= last; i += stride) {
    const y = Math.floor(i / 12),
      m = (i % 12) + 1;
    if (y < 1 || y > 9999) continue;
    for (const t of localCandidates(utcFromParts(y, m, 1), zone))
      if (t >= start && t <= end) out.push(t);
  }
  return out.sort((a, b) => a - b);
}
const MONTHS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];
const DAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];
const pad = (v: number, n: number = 2): string => String(v).padStart(n, "0");
/** Format a timestamp with fixed English calendar names and IANA timezone rules. */
export function formatTime(
  t: Timestamp,
  zone: string = "UTC",
  format: string = "%H:%M",
): string {
  t = epoch(t);
  const p = wallParts(t, zone),
    d = new Date(utcFromParts(p.year, p.month, p.day) * 1000),
    off = offsetAt(t, zone),
    day = d.getUTCDay();
  const fields: Record<string, string> = {
    H: pad(p.hour),
    I: pad(p.hour % 12 || 12),
    M: pad(p.minute),
    S: pad(p.second),
    d: pad(p.day),
    e: String(p.day).padStart(2, " "),
    m: pad(p.month),
    Y: pad(p.year, 4),
    y: pad(p.year % 100),
    a: DAYS[day],
    A: [
      "Sunday",
      "Monday",
      "Tuesday",
      "Wednesday",
      "Thursday",
      "Friday",
      "Saturday",
    ][day],
    b: MONTHS[p.month - 1],
    h: MONTHS[p.month - 1],
    B: [
      "January",
      "February",
      "March",
      "April",
      "May",
      "June",
      "July",
      "August",
      "September",
      "October",
      "November",
      "December",
    ][p.month - 1],
    p: p.hour < 12 ? "AM" : "PM",
    w: String(day),
    u: String(day || 7),
    j: pad(
      Math.floor((d.getTime() / 1000 - utcFromParts(p.year, 1, 1)) / 86400) + 1,
      3,
    ),
    z:
      (off < 0 ? "-" : "+") +
      pad(Math.floor(Math.abs(off) / 3600)) +
      pad(Math.floor((Math.abs(off) % 3600) / 60)),
    Z: zone,
    "%": "%",
  };
  fields.F = fields.Y + "-" + fields.m + "-" + fields.d;
  fields.T = fields.H + ":" + fields.M + ":" + fields.S;
  fields.R = fields.H + ":" + fields.M;
  let out = "";
  for (let i = 0; i < format.length; i++) {
    if (format[i] !== "%") {
      out += format[i];
      continue;
    }
    const k = format[++i];
    check(own(fields, k), "Unsupported date directive: %" + (k || ""));
    out += fields[k];
  }
  return out;
}
function resolveX(
  a: DeepReadonly<ResolvedTimeAxis>,
  start: number,
  end: number,
  width: number,
): XResolution {
  const span = end - start,
    zone = a.timezone,
    mode =
      a.mode === "auto"
        ? span <= 172800
          ? "daily"
          : span <= 864000
            ? "weekly"
            : span <= 5356800
              ? "monthly"
              : "yearly"
        : a.mode;
  let minor: number[] = [],
    major: number[] = [],
    labelTimes: number[] = [],
    format = "%H:%M";
  if (mode === "yearly") {
    const stride =
      span > 550 * 86400 ? Math.max(1, Math.ceil(span / (365.25 * 86400))) : 1;
    format = stride > 1 ? "%b %Y" : "%b";
    if (a.majorTicks === null || a.ticks === null) {
      major = monthTicks(start, end, zone, stride);
      labelTimes = major;
    }
    if (a.minorTicks === null) minor = monthTicks(start, end, zone, 1);
    if (a.majorTicks === null && a.majorSeconds !== null)
      major = wallTicks(start, end, a.majorSeconds, zone);
    if (a.minorTicks === null && a.minorSeconds !== null)
      minor = wallTicks(start, end, a.minorSeconds, zone);
    if (a.ticks === null && a.labelSeconds !== null)
      labelTimes = wallTicks(start, end, a.labelSeconds, zone);
  } else {
    let mi = 1800,
      ma = 7200,
      ls = 7200;
    if (mode === "weekly") {
      mi = 21600;
      ma = ls = 86400;
      format = "%d";
    }
    if (mode === "monthly") {
      mi = 86400;
      ma = ls = 604800;
      format = "%d %b";
    }
    if (a.mode === "auto" && span < 43200) {
      const target = span / Math.max(2, Math.floor(width / 48));
      ls =
        [1, 5, 10, 15, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200].find(
          (v) => v >= target,
        ) || 7200;
      mi = Math.max(1, ls / 4);
      ma = ls;
      if (ls < 60) format = "%H:%M:%S";
    }
    mi = Math.max(mi, span / 2000);
    if (a.minorSeconds !== null) mi = a.minorSeconds;
    if (a.majorSeconds !== null) ma = a.majorSeconds;
    if (a.labelSeconds !== null) ls = a.labelSeconds;
    if (a.minorTicks === null) minor = wallTicks(start, end, mi, zone);
    if (a.majorTicks === null) major = wallTicks(start, end, ma, zone);
    if (a.ticks === null) labelTimes = wallTicks(start, end, ls, zone);
  }
  if (a.minorTicks !== null)
    minor = a.minorTicks.filter((v) => v >= start && v <= end);
  if (a.majorTicks !== null)
    major = a.majorTicks.filter((v) => v >= start && v <= end);
  if (a.labelFormat !== null) format = a.labelFormat;
  let labels: ResolvedTick[] = [];
  if (a.ticks !== null)
    labels = a.ticks.filter((v) => v.time >= start && v.time <= end);
  else if (
    mode === "weekly" &&
    a.labelSeconds === null &&
    a.labelOffsetSeconds === 0
  ) {
    for (const t of wallTicks(start - 172800, end, 86400, zone)) {
      const w = wallEpoch(t, zone),
        next = localCandidates(w + 86400, zone);
      if (t < start || !next.length || next[next.length - 1] > end) continue;
      for (const noon of localCandidates(w + 43200, zone))
        if (noon >= start && noon <= end)
          labels.push({ time: noon, label: formatTime(t, zone, format) });
    }
  } else
    for (const t of labelTimes) {
      const pos = t + a.labelOffsetSeconds;
      if (pos >= start && pos <= end)
        labels.push({ time: pos, label: formatTime(t, zone, format) });
    }
  labels.sort((a, b) => a.time - b.time);
  if (a.ticks === null && mode === "daily" && span <= 95040) {
    const seen = new Map<string, Set<number>>();
    for (const t of labels) {
      if (!seen.has(t.label)) seen.set(t.label, new Set());
      seen.get(t.label)!.add(offsetAt(t.time, zone));
    }
    labels = labels.map((t) =>
      seen.get(t.label)!.size > 1
        ? {
            time: t.time,
            label: t.label + " " + formatTime(t.time, zone, "%z"),
          }
        : t,
    );
  }
  const majorSet = new Set(major);
  minor = minor.filter((v) => !majorSet.has(v));
  return { minor, major, labels, mode };
}
function timeRange(c: ChartConfig): [number, number] {
  let start = Infinity,
    end = -Infinity;
  for (const s of c.series)
    if (s.timestamps.length) {
      start = Math.min(start, s.timestamps[0]);
      end = Math.max(end, s.timestamps[s.timestamps.length - 1]);
    }
  if (c.timeAxis.start !== null) start = c.timeAxis.start;
  if (c.timeAxis.end !== null) end = c.timeAxis.end;
  if (start === end && c.timeAxis.start === null && c.timeAxis.end === null) {
    start -= 150;
    end += 150;
  }
  epoch(start);
  epoch(end);
  check(
    end > start,
    "Empty data needs explicit start and end; range must be nonzero.",
  );
  return [start, end];
}
// Original 5×7 pixel-letter descriptions. Not extracted from a font, no font asset shipped.
// This optional ASCII renderer trades typographic likeness for host-independent pixels.
const PIXEL_GLYPHS: Readonly<Record<string, readonly string[]>> = {
  " ": ["00000", "00000", "00000", "00000", "00000", "00000", "00000"],
  "0": ["01110", "10001", "10011", "10101", "11001", "10001", "01110"],
  "1": ["00100", "01100", "00100", "00100", "00100", "00100", "01110"],
  "2": ["01110", "10001", "00001", "00010", "00100", "01000", "11111"],
  "3": ["11110", "00001", "00001", "01110", "00001", "00001", "11110"],
  "4": ["00010", "00110", "01010", "10010", "11111", "00010", "00010"],
  "5": ["11111", "10000", "10000", "11110", "00001", "00001", "11110"],
  "6": ["01110", "10000", "10000", "11110", "10001", "10001", "01110"],
  "7": ["11111", "00001", "00010", "00100", "01000", "01000", "01000"],
  "8": ["01110", "10001", "10001", "01110", "10001", "10001", "01110"],
  "9": ["01110", "10001", "10001", "01111", "00001", "00001", "01110"],
  A: ["01110", "10001", "10001", "11111", "10001", "10001", "10001"],
  B: ["11110", "10001", "10001", "11110", "10001", "10001", "11110"],
  C: ["01111", "10000", "10000", "10000", "10000", "10000", "01111"],
  D: ["11110", "10001", "10001", "10001", "10001", "10001", "11110"],
  E: ["11111", "10000", "10000", "11110", "10000", "10000", "11111"],
  F: ["11111", "10000", "10000", "11110", "10000", "10000", "10000"],
  G: ["01110", "10001", "10000", "10111", "10001", "10001", "01111"],
  H: ["10001", "10001", "10001", "11111", "10001", "10001", "10001"],
  I: ["01110", "00100", "00100", "00100", "00100", "00100", "01110"],
  J: ["00111", "00010", "00010", "00010", "00010", "10010", "01100"],
  K: ["10001", "10010", "10100", "11000", "10100", "10010", "10001"],
  L: ["10000", "10000", "10000", "10000", "10000", "10000", "11111"],
  M: ["10001", "11011", "10101", "10101", "10001", "10001", "10001"],
  N: ["10001", "11001", "10101", "10011", "10001", "10001", "10001"],
  O: ["01110", "10001", "10001", "10001", "10001", "10001", "01110"],
  P: ["11110", "10001", "10001", "11110", "10000", "10000", "10000"],
  Q: ["01110", "10001", "10001", "10001", "10101", "10010", "01101"],
  R: ["11110", "10001", "10001", "11110", "10100", "10010", "10001"],
  S: ["01111", "10000", "10000", "01110", "00001", "00001", "11110"],
  T: ["11111", "00100", "00100", "00100", "00100", "00100", "00100"],
  U: ["10001", "10001", "10001", "10001", "10001", "10001", "01110"],
  V: ["10001", "10001", "10001", "10001", "10001", "01010", "00100"],
  W: ["10001", "10001", "10001", "10101", "10101", "11011", "10001"],
  X: ["10001", "10001", "01010", "00100", "01010", "10001", "10001"],
  Y: ["10001", "10001", "01010", "00100", "00100", "00100", "00100"],
  Z: ["11111", "00001", "00010", "00100", "01000", "10000", "11111"],
  a: ["00000", "00000", "01110", "00001", "01111", "10001", "01111"],
  b: ["10000", "10000", "10110", "11001", "10001", "10001", "11110"],
  c: ["00000", "00000", "01111", "10000", "10000", "10000", "01111"],
  d: ["00001", "00001", "01101", "10011", "10001", "10001", "01111"],
  e: ["00000", "00000", "01110", "10001", "11111", "10000", "01110"],
  f: ["00110", "01001", "01000", "11100", "01000", "01000", "01000"],
  g: ["00000", "01111", "10001", "10001", "01111", "00001", "01110"],
  h: ["10000", "10000", "10110", "11001", "10001", "10001", "10001"],
  i: ["00100", "00000", "01100", "00100", "00100", "00100", "01110"],
  j: ["00010", "00000", "00110", "00010", "00010", "10010", "01100"],
  k: ["10000", "10000", "10010", "10100", "11000", "10100", "10010"],
  l: ["01100", "00100", "00100", "00100", "00100", "00100", "01110"],
  m: ["00000", "00000", "11010", "10101", "10101", "10101", "10101"],
  n: ["00000", "00000", "10110", "11001", "10001", "10001", "10001"],
  o: ["00000", "00000", "01110", "10001", "10001", "10001", "01110"],
  p: ["00000", "00000", "11110", "10001", "11110", "10000", "10000"],
  q: ["00000", "00000", "01111", "10001", "01111", "00001", "00001"],
  r: ["00000", "00000", "10111", "11000", "10000", "10000", "10000"],
  s: ["00000", "00000", "01111", "10000", "01110", "00001", "11110"],
  t: ["01000", "01000", "11100", "01000", "01000", "01001", "00110"],
  u: ["00000", "00000", "10001", "10001", "10001", "10011", "01101"],
  v: ["00000", "00000", "10001", "10001", "10001", "01010", "00100"],
  w: ["00000", "00000", "10001", "10001", "10101", "10101", "01010"],
  x: ["00000", "00000", "10001", "01010", "00100", "01010", "10001"],
  y: ["00000", "00000", "10001", "10001", "01111", "00001", "01110"],
  z: ["00000", "00000", "11111", "00010", "00100", "01000", "11111"],
  "-": ["00000", "00000", "00000", "11111", "00000", "00000", "00000"],
  _: ["00000", "00000", "00000", "00000", "00000", "00000", "11111"],
  ":": ["00000", "00100", "00100", "00000", "00100", "00100", "00000"],
  ".": ["00000", "00000", "00000", "00000", "00000", "00100", "00100"],
  ",": ["00000", "00000", "00000", "00000", "00100", "00100", "01000"],
  "/": ["00001", "00010", "00010", "00100", "01000", "01000", "10000"],
  "\\": ["10000", "01000", "01000", "00100", "00010", "00010", "00001"],
  "(": ["00010", "00100", "01000", "01000", "01000", "00100", "00010"],
  ")": ["01000", "00100", "00010", "00010", "00010", "00100", "01000"],
  "[": ["01110", "01000", "01000", "01000", "01000", "01000", "01110"],
  "]": ["01110", "00010", "00010", "00010", "00010", "00010", "01110"],
  "+": ["00000", "00100", "00100", "11111", "00100", "00100", "00000"],
  "=": ["00000", "00000", "11111", "00000", "11111", "00000", "00000"],
  "%": ["11001", "11010", "00010", "00100", "01000", "01011", "10011"],
  "?": ["01110", "10001", "00001", "00010", "00100", "00000", "00100"],
  "!": ["00100", "00100", "00100", "00100", "00100", "00000", "00100"],
  "#": ["01010", "01010", "11111", "01010", "11111", "01010", "01010"],
  "*": ["00000", "10101", "01110", "11111", "01110", "10101", "00000"],
  "<": ["00010", "00100", "01000", "10000", "01000", "00100", "00010"],
  ">": ["01000", "00100", "00010", "00001", "00010", "00100", "01000"],
  "|": ["00100", "00100", "00100", "00100", "00100", "00100", "00100"],
  '"': ["01010", "01010", "00000", "00000", "00000", "00000", "00000"],
  "'": ["00100", "00100", "00000", "00000", "00000", "00000", "00000"],
  ";": ["00000", "00100", "00100", "00000", "00100", "00100", "01000"],
  "@": ["01110", "10001", "10111", "10101", "10111", "10000", "01111"],
  $: ["00100", "01111", "10100", "01110", "00101", "11110", "00100"],
  "&": ["01100", "10010", "10100", "01000", "10101", "10010", "01101"],
  "^": ["00100", "01010", "10001", "00000", "00000", "00000", "00000"],
  "`": ["01000", "00100", "00000", "00000", "00000", "00000", "00000"],
  "~": ["00000", "00000", "01001", "10110", "00000", "00000", "00000"],
  "{": ["00011", "00100", "00100", "01000", "00100", "00100", "00011"],
  "}": ["11000", "00100", "00100", "00010", "00100", "00100", "11000"],
};
function createCanvas(w: number, h: number): Canvas {
  let canvas;
  if (typeof document !== "undefined" && document.createElement)
    canvas = document.createElement("canvas");
  else if (typeof OffscreenCanvas !== "undefined")
    canvas = new OffscreenCanvas(w, h);
  else
    throw new Error(
      'System fonts need Canvas 2D. Use fonts: { mode: "bitmap" } in a DOM-free runtime.',
    );
  canvas.width = w;
  canvas.height = h;
  return canvas;
}
function wide(ch: string): boolean {
  const n = ch.codePointAt(0)!;
  return (
    n >= 0x1100 &&
    (n <= 0x115f ||
      (n >= 0x2e80 && n <= 0xa4cf) ||
      (n >= 0xac00 && n <= 0xd7a3) ||
      (n >= 0xf900 && n <= 0xfaff) ||
      (n >= 0xff00 && n <= 0xff60) ||
      n >= 0x1f300)
  );
}
class Fonts {
  readonly config: Readonly<Required<FontOptions>>;
  readonly theme: DeepReadonly<ResolvedTheme>;
  private readonly cache: Map<string, Surface>;
  private canvas!: Canvas;
  private ctx!: Context2D;
  constructor(
    config: Readonly<Required<FontOptions>>,
    theme: DeepReadonly<ResolvedTheme>,
  ) {
    this.config = config;
    this.theme = theme;
    this.cache = new Map();
    if (config.mode === "system") {
      this.canvas = createCanvas(8, 8);
      const ctx = this.canvas.getContext("2d", {
        willReadFrequently: true,
      }) as Context2D | null;
      check(ctx, "Canvas 2D is unavailable.");
      this.ctx = ctx;
    }
  }
  size(role: FontRole, scale: number = 1): number {
    return this.theme[`${role}Size`] * scale;
  }
  family(role: FontRole): string {
    return (
      (role === "title" || role === "unit" || role === "caption"
        ? this.config[`${role}Family`]
        : null) || this.config.family
    );
  }
  setup(role: FontRole, size: number): void {
    this.ctx.font =
      (role === "caption" ? "bold " : "") + size + "px " + this.family(role);
    this.ctx.textBaseline = "alphabetic";
    this.ctx.fillStyle = "#000000";
    if ("fontKerning" in this.ctx) this.ctx.fontKerning = "none";
  }
  width(
    value: string,
    role: FontRole,
    advance: number = 0,
    scale: number = 1,
  ): number {
    if (advance > 0)
      return [...value].reduce((n, ch) => n + advance * (wide(ch) ? 2 : 1), 0);
    if (this.config.mode === "bitmap")
      return [...value].length * this.size(role, scale) * 0.6;
    this.setup(role, this.size(role, scale));
    return this.ctx.measureText(value).width;
  }
  fit(
    value: string,
    maxWidth: number,
    role: FontRole,
    advance: number,
    scale: number = 1,
  ): string {
    if (this.width(value, role, advance, scale) <= maxWidth) return value;
    const end = "...";
    if (this.width(end, role, advance, scale) > maxWidth) return "";
    const a = [...value];
    while (
      a.length &&
      this.width(a.join("") + end, role, advance, scale) > maxWidth
    )
      a.pop();
    return a.join("") + end;
  }
  raster(
    value: string,
    role: FontRole,
    advance: number = 0,
    scale: number = 1,
  ): Surface {
    const key = JSON.stringify([value, role, advance, scale]);
    if (this.cache.has(key)) return this.cache.get(key)!;
    const size = this.size(role, scale),
      w = Math.max(1, Math.ceil(this.width(value, role, advance, scale)) + 4);
    let out;
    check(
      w <= 32768 && w * size < LIMITS.pixels,
      "Text allocation limit exceeded.",
    );
    if (this.config.mode === "bitmap") {
      const h = Math.max(1, round(size * 0.72));
      out = new Surface(w, h + 2);
      let x = 1;
      for (const ch of value) {
        let rows = PIXEL_GLYPHS[ch];
        if (!rows) {
          check(
            !this.config.strictGlyphs,
            "Bitmap text supports printable ASCII only: " + ch,
          );
          rows = PIXEL_GLYPHS["?"];
        }
        const cell = advance > 0 ? advance : size * 0.6,
          gw = Math.max(1, Math.min(round(size * 0.5), Math.floor(cell) - 1));
        for (let yy = 0; yy < h; yy++)
          for (let xx = 0; xx < gw; xx++)
            if (
              rows[Math.min(6, Math.floor((yy / h) * 7))][
                Math.min(4, Math.floor((xx / gw) * 5))
              ] === "1"
            )
              out.pixel(round(x) + xx, yy + 1, [0, 0, 0, 255]);
        x += cell * (wide(ch) ? 2 : 1);
      }
    } else {
      this.setup(role, size);
      const metric = this.ctx.measureText(value || "0");
      const ascent = Math.ceil(metric.actualBoundingBoxAscent || size * 0.8),
        descent = Math.ceil(metric.actualBoundingBoxDescent || 0),
        h = Math.max(1, ascent + descent) + 2;
      this.canvas.width = w;
      this.canvas.height = h;
      this.setup(role, size);
      let x = 1;
      for (const ch of value) {
        this.ctx.fillText(ch, round(x), 1 + ascent);
        x +=
          advance > 0
            ? advance * (wide(ch) ? 2 : 1)
            : this.ctx.measureText(ch).width;
      }
      const d = this.ctx.getImageData(0, 0, w, h);
      out = new Surface(w, h);
      out.data.set(d.data);
    }
    if (this.cache.size >= 512)
      this.cache.delete(this.cache.keys().next().value!);
    this.cache.set(key, out);
    return out;
  }
  draw(
    im: Surface,
    x: number,
    y: number,
    value: string,
    role: FontRole,
    fill: Readonly<RGBA>,
    advance: number = 0,
    align: "left" | "center" | "right" = "left",
    centerY: boolean = false,
    scale: number = 1,
  ): void {
    if (!value) return;
    const r = this.raster(value, role, advance, scale),
      width = this.width(value, role, advance, scale);
    if (align === "center") x -= width / 2;
    else if (align === "right") x -= width;
    if (centerY) y -= (r.height - 2) / 2;
    const xx = round(x) - 1,
      yy = round(y) - 1,
      p: RGBA = [fill[0], fill[1], fill[2], 0];
    for (let sy = 0; sy < r.height; sy++)
      for (let sx = 0; sx < r.width; sx++) {
        const a = r.data[(sy * r.width + sx) * 4 + 3];
        if (a) {
          p[3] = round((a * fill[3]) / 255);
          im.pixel(xx + sx, yy + sy, p, true);
        }
      }
  }
  rotated(
    value: string,
    role: FontRole,
    fill: Readonly<RGBA>,
    clockwise: boolean = false,
  ): Surface {
    const r = this.raster(value, role),
      out = new Surface(r.height, r.width),
      p: RGBA = [fill[0], fill[1], fill[2], 0];
    for (let y = 0; y < r.height; y++)
      for (let x = 0; x < r.width; x++) {
        p[3] = round((r.data[(y * r.width + x) * 4 + 3] * fill[3]) / 255);
        if (p[3])
          out.pixel(
            clockwise ? r.height - 1 - y : y,
            clockwise ? x : r.width - 1 - x,
            p,
          );
      }
    return out;
  }
}
function renderChart(c: ChartConfig): RenderResult {
  const l = c.layout,
    t = c.theme,
    [width, height] = dimensions(c),
    left = l.left,
    top = l.top,
    right = l.width - l.right,
    bottom = l.top + l.plotHeight,
    pw = right - left,
    ph = bottom - top;
  const [start, end] = timeRange(c);
  const runs = c.series.map((s) => visibleRuns(s, start, end));
  let loData = Infinity,
    hiData = -Infinity;
  for (let i = 0; i < runs.length; i++) {
    for (const run of runs[i])
      for (const p of run) {
        check(finite(p[1]), "Interpolated value overflow.");
        loData = Math.min(loData, p[1]);
        hiData = Math.max(hiData, p[1]);
      }
    if (c.series[i].kind === "area") {
      loData = Math.min(loData, c.series[i].baseline);
      hiData = Math.max(hiData, c.series[i].baseline);
    }
  }
  if (loData === Infinity) loData = hiData = 0;
  const ys = resolveY(c.yAxis, loData, hiData),
    xs = resolveX(c.timeAxis, start, end, pw),
    fonts = new Fonts(c.fonts, t),
    im = new Surface(width, height, t.background);
  im.rect(left, top, right + 1, bottom + 1, t.canvas);
  const xx = (v: number): number => ((v - start) / (end - start)) * pw,
    yy = (v: number): number =>
      ph *
      (1 -
        (v / (ys.maximum - ys.minimum) -
          ys.minimum / (ys.maximum - ys.minimum)));
  const projected = runs.map((rr) =>
    rr.map((r) =>
      decimate(r, start, end, pw).map((p): Point => {
        const x = xx(p[0]),
          y = yy(p[1]);
        check(
          finite(x) && finite(y) && Math.abs(y) < 1e15,
          "Data magnitude is too large relative to the Y axis.",
        );
        return [x, y];
      }),
    ),
  );
  function grid(): void {
    const lay = new Surface(pw + 1, ph + 1);
    for (const [values, col] of [
      [ys.minor, t.minorGrid],
      [ys.major, t.majorGrid],
    ] as const)
      for (const v of values) {
        const y = round(yy(v));
        if (y >= 0 && y <= ph) lay.dashed([0, y], [pw, y], col, t.gridDash);
      }
    for (const [values, col] of [
      [xs.minor, t.minorGrid],
      [xs.major, t.majorGrid],
    ] as const)
      for (const v of values) {
        const x = round(xx(v));
        lay.dashed([x, 0], [x, ph], col, t.gridDash);
      }
    im.over(lay, left, top);
  }
  const aa = l.antialias,
    layerWidth = (pw + 1) * aa,
    layerHeight = (ph + 1) * aa;
  if (!t.gridFront) grid();
  for (let i = 0; i < c.series.length; i++) {
    const s = c.series[i];
    if (s.kind !== "area") continue;
    const base = yy(s.baseline);
    check(
      finite(base) && Math.abs(base) < 1e15,
      "Baseline magnitude is too large.",
    );
    const layer = new Surface(layerWidth, layerHeight);
    for (const ps of projected[i])
      if (ps.length > 1) {
        const poly = clipPolygon(
          [[ps[0][0], base], ...ps, [ps[ps.length - 1][0], base]],
          pw,
          ph,
        ).map((p): Point => [round(p[0] * aa), round(p[1] * aa)]);
        layer.polygon(poly, s.color);
      }
    im.over(layer.down(aa), left, top);
  }
  if (t.gridFront) grid();
  for (let i = 0; i < c.series.length; i++) {
    const s = c.series[i],
      col = s.kind === "line" ? s.color : s.outline;
    if (col === null) continue;
    const layer = new Surface(layerWidth, layerHeight),
      lw = Math.max(1, round(s.lineWidth * aa));
    for (const ps of projected[i]) {
      if (ps.length === 1) {
        const p = ps[0];
        if (p[0] >= 0 && p[0] <= pw && p[1] >= 0 && p[1] <= ph)
          layer.circle(p[0] * aa, p[1] * aa, Math.max(aa / 2, lw / 2), col);
      }
      for (let k = 1; k < ps.length; k++) {
        const seg = clipLine(ps[k - 1], ps[k], pw, ph);
        if (seg)
          layer.line(
            [round(seg[0][0] * aa), round(seg[0][1] * aa)],
            [round(seg[1][0] * aa), round(seg[1][1] * aa)],
            col,
            lw,
          );
      }
    }
    im.over(layer.down(aa), left, top);
  }
  const rules = new Surface(pw + 1, ph + 1);
  for (const r of c.hRules)
    if (r.value >= ys.minimum && r.value <= ys.maximum)
      rules.dashed(
        [0, round(yy(r.value))],
        [pw, round(yy(r.value))],
        r.color,
        r.dash,
        Math.max(1, round(r.width)),
      );
  for (const r of c.vRules)
    if (r.time >= start && r.time <= end)
      rules.dashed(
        [round(xx(r.time)), 0],
        [round(xx(r.time)), ph],
        r.color,
        r.dash,
        Math.max(1, round(r.width)),
      );
  im.over(rules, left, top);
  im.line([left, top - 3], [left, bottom + 4], t.axis);
  im.line([left - 4, bottom], [right + 4, bottom], t.axis);
  im.polygon(
    [
      [left, top - 5],
      [left - 3, top],
      [left + 3, top],
    ],
    t.arrow,
  );
  im.polygon(
    [
      [right + 7, bottom],
      [right + 2, bottom - 3],
      [right + 2, bottom + 3],
    ],
    t.arrow,
  );
  for (const v of ys.major) {
    const y = top + yy(v),
      unit =
        v === 0 && !c.yAxis.showZeroSuffix
          ? { factor: ys.factor, suffix: "" }
          : ys,
      label = formatValue(v, unit, ys.decimals);
    check(
      fonts.width(label, "axis", t.axisAdvance) <= left - l.yLabelGap - 20,
      "Y labels overlap the vertical label. Increase layout.left or adjust units.",
    );
    im.line([left - 3, round(y)], [left, round(y)], t.axis);
    fonts.draw(
      im,
      left - l.yLabelGap,
      y,
      label,
      "axis",
      t.text,
      t.axisAdvance,
      "right",
      true,
    );
  }
  const xLabels = [];
  let lastRight = -Infinity;
  for (const tick of xs.labels) {
    const x = left + xx(tick.time),
      tw = fonts.width(tick.label, "axis", t.axisAdvance),
      a = x - tw / 2,
      b = x + tw / 2;
    if (c.timeAxis.ticks === null && a < lastRight + 3) continue;
    if (a < 2 || b > width - 3) continue;
    im.line(
      [round(x), bottom],
      [round(x), bottom + 3],
      [t.majorGrid[0], t.majorGrid[1], t.majorGrid[2], 255],
    );
    fonts.draw(
      im,
      x,
      bottom + l.xLabelGap,
      tick.label,
      "axis",
      t.text,
      t.axisAdvance,
      "center",
    );
    lastRight = b;
    xLabels.push({ ...tick, x });
  }
  const titleX = (left + right) / 2 + l.titleOffsetX,
    title = fonts.fit(
      c.title,
      2 * Math.min(titleX - 5, width - 14 - titleX),
      "title",
      t.titleAdvance,
    );
  fonts.draw(
    im,
    titleX,
    l.titleY,
    title,
    "title",
    t.text,
    t.titleAdvance,
    "center",
  );
  if (c.verticalLabel) {
    const unit = fonts.rotated(c.verticalLabel, "unit", t.text);
    check(
      unit.height <= ph + 18 && l.unitX + unit.width <= left - l.yLabelGap,
      "Vertical label does not fit.",
    );
    im.over(unit, l.unitX, round((top + bottom - unit.height) / 2));
  }
  if (c.watermark) {
    const mark = fonts.rotated(c.watermark, "watermark", t.watermark, true);
    check(
      mark.height <= height - 8 && mark.width <= l.right - 9,
      "Watermark does not fit.",
    );
    im.over(mark, width - mark.width - 4, 4);
  }
  const stats = c.series.map((s) => statistics(s, start, end));
  if (l.legend !== "none") {
    const scale = Math.min(1, (width - 40) / 555),
      advance = t.legendAdvance * scale,
      ll = l.legendLayout,
      factor = ll.autoScaleColumns
        ? (width - ll.nameX) / (ll.referenceWidth - ll.nameX)
        : 1,
      anchor = (x: number): number => ll.nameX + (x - ll.nameX) * factor;
    c.series.forEach((s, i) => {
      const pairs =
          l.legend === "aligned"
            ? ll.aligned
            : l.legend === "reference" && i === c.series.length - 1 && i > 0
              ? ll.expanded
              : ll.compact,
        y = bottom + l.legendGap + i * l.legendRowHeight;
      check(
        ll.swatchHeight <= l.legendRowHeight &&
          ll.swatchX + ll.swatchWidth < width &&
          y + ll.swatchHeight <= height - 2,
        "Legend swatch does not fit.",
      );
      im.rect(
        ll.swatchX,
        y,
        ll.swatchX + ll.swatchWidth,
        y + ll.swatchHeight,
        t.frame,
      );
      const swatch = new Surface(
        ll.swatchWidth - 2,
        ll.swatchHeight - 2,
        s.color,
      );
      im.over(swatch, ll.swatchX + 1, y + 1);
      const name = fonts.fit(
        s.name,
        anchor(pairs[0][0]) - ll.nameX - 12,
        "legend",
        advance,
        scale,
      );
      fonts.draw(
        im,
        ll.nameX,
        y,
        name,
        "legend",
        t.text,
        advance,
        "left",
        false,
        scale,
      );
      const display = s.legendValues || stats[i];
      (["current", "average", "maximum"] as const).forEach((key, j) => {
        const label = c.legendLabels[j],
          value = formatValue(
            display[key],
            ys,
            c.yAxis.legendDecimals,
            c.missingLabel,
          ),
          a = anchor(pairs[j][0]),
          b = anchor(pairs[j][1]);
        check(b < width - 3, "Legend column is outside the panel.");
        check(
          fonts.width(label, "legend", advance, scale) +
            fonts.width(value, "legend", advance, scale) +
            7 * scale <=
            b - a + 1,
          "Legend statistic does not fit. Increase width, adjust anchors or reduce legendDecimals.",
        );
        fonts.draw(
          im,
          a,
          y,
          label,
          "legend",
          t.text,
          advance,
          "left",
          false,
          scale,
        );
        fonts.draw(
          im,
          b,
          y,
          value,
          "legend",
          t.text,
          advance,
          "right",
          false,
          scale,
        );
      });
    });
  }
  for (const [a, b] of [
    [
      [0, 0],
      [width - 1, 0],
    ],
    [
      [1, 1],
      [width - 2, 1],
    ],
    [
      [0, 0],
      [0, height - 1],
    ],
    [
      [1, 1],
      [1, height - 2],
    ],
  ] as [Point, Point][])
    im.line(a, b, t.shadeLight);
  for (const [a, b] of [
    [
      [width - 2, 1],
      [width - 2, height - 1],
    ],
    [
      [width - 1, 0],
      [width - 1, height - 1],
    ],
    [
      [1, height - 2],
      [width - 1, height - 2],
    ],
    [
      [0, height - 1],
      [width - 1, height - 1],
    ],
  ] as [Point, Point][])
    im.line(a, b, t.shadeDark);
  for (let i = 3; i < im.data.length; i += 4) im.data[i] = 255;
  const out = im.scale(l.pixelScale);
  const metadata: GraphMetadata = {
    version: VERSION,
    imageSize: [out.width, out.height],
    logicalSize: [width, height],
    plotBox: [left, top, right, bottom],
    pixelScale: l.pixelScale,
    title: c.title,
    verticalLabel: c.verticalLabel,
    watermark: c.watermark,
    timeRange: [start, end],
    timezone: c.timeAxis.timezone,
    timeMode: xs.mode,
    yRange: [ys.minimum, ys.maximum],
    yStep: ys.step,
    yUnit: { factor: ys.factor, suffix: ys.suffix },
    xLabels,
    statistics: stats,
    statisticsPolicy:
      "inclusive viewport; original sample arithmetic mean; current includes final missing sample",
    font: {
      mode: c.fonts.mode,
      family: c.fonts.mode === "system" ? c.fonts.family : null,
      pixelAlphabet: c.fonts.mode === "bitmap" ? "ascii-5x7-v1" : null,
    },
    layout: clone(l),
    theme: clone(t),
    warnings:
      c.fonts.mode === "system"
        ? ["System font selection and glyph rasterization depend on the host."]
        : [],
  };
  return new RenderResult(out, metadata);
}
// Dependency-free PNG encoder: adaptive row filters + deterministic fixed-Huffman DEFLATE.
// No Canvas encoder or compression library is involved in exported PNG bytes.
const CRC_TABLE = (() => {
  const t = new Uint32Array(256);
  for (let i = 0; i < 256; i++) {
    let c = i;
    for (let j = 0; j < 8; j++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
    t[i] = c >>> 0;
  }
  return t;
})();
function crc32(bytes: Uint8Array): number {
  let c = 0xffffffff;
  for (const b of bytes) c = CRC_TABLE[(c ^ b) & 255] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}
function adler32(bytes: Uint8Array): number {
  let a = 1,
    b = 0;
  for (let i = 0; i < bytes.length;) {
    const end = Math.min(i + 5552, bytes.length);
    for (; i < end; i++) {
      a += bytes[i];
      b += a;
    }
    a %= 65521;
    b %= 65521;
  }
  return ((b << 16) | a) >>> 0;
}
function concat(parts: readonly Uint8Array[]): Uint8Array<ArrayBuffer> {
  const out = new Uint8Array(parts.reduce((n, p) => n + p.length, 0));
  let at = 0;
  for (const p of parts) {
    out.set(p, at);
    at += p.length;
  }
  return out;
}
function u32(n: number): Uint8Array<ArrayBuffer> {
  return new Uint8Array([
    (n >>> 24) & 255,
    (n >>> 16) & 255,
    (n >>> 8) & 255,
    n & 255,
  ]);
}
function chunk(type: string, data: Uint8Array): Uint8Array<ArrayBuffer> {
  const code = Uint8Array.from(type, (c) => c.charCodeAt(0)),
    body = concat([code, data]);
  return concat([u32(data.length), body, u32(crc32(body))]);
}
function reverseBits(n: number, bits: number): number {
  let out = 0;
  for (let i = 0; i < bits; i++) {
    out = (out << 1) | (n & 1);
    n >>>= 1;
  }
  return out;
}
const FIXED = (() => {
  const out = [];
  for (let i = 0; i < 288; i++) {
    let code, bits;
    if (i <= 143) {
      bits = 8;
      code = 0x30 + i;
    } else if (i <= 255) {
      bits = 9;
      code = 0x190 + i - 144;
    } else if (i <= 279) {
      bits = 7;
      code = i - 256;
    } else {
      bits = 8;
      code = 0xc0 + i - 280;
    }
    out.push([reverseBits(code, bits), bits]);
  }
  return out;
})();
const LEN_BASE = [
  3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67,
  83, 99, 115, 131, 163, 195, 227, 258,
];
const LEN_EXTRA = [
  0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5,
  5, 5, 0,
];
const DIST_BASE = [
  1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513, 769,
  1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577,
];
const DIST_EXTRA = [
  0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11,
  11, 12, 12, 13, 13,
];
function deflate(data: Uint8Array): Uint8Array<ArrayBuffer> {
  let bytes = new Uint8Array(Math.max(128, data.length + 64)),
    len = 0,
    buffer = 0,
    count = 0;
  function byte(b: number): void {
    if (len === bytes.length) {
      const next = new Uint8Array(bytes.length * 2);
      next.set(bytes);
      bytes = next;
    }
    bytes[len++] = b;
  }
  function bits(value: number, n: number): void {
    buffer |= value << count;
    count += n;
    while (count >= 8) {
      byte(buffer & 255);
      buffer >>>= 8;
      count -= 8;
    }
  }
  const sym = (v: number): void => bits(FIXED[v][0], FIXED[v][1]);
  byte(0x78);
  byte(0x01);
  bits(3, 3); // final block, fixed Huffman
  const head = new Int32Array(65536).fill(-1),
    prev = new Int32Array(32768).fill(-1);
  const hash = (i: number): number =>
    ((data[i] * 251 + data[i + 1]) * 251 + data[i + 2]) & 65535;
  function insert(i: number): void {
    if (i + 2 >= data.length) return;
    const h = hash(i);
    prev[i & 32767] = head[h];
    head[h] = i;
  }
  for (let i = 0; i < data.length;) {
    let best = 0,
      dist = 0;
    if (i + 2 < data.length) {
      let candidate = head[hash(i)],
        tries = 64;
      const max = Math.min(258, data.length - i);
      while (
        candidate >= 0 &&
        i - candidate <= 32768 &&
        candidate < i &&
        tries--
      ) {
        if (
          data[candidate] === data[i] &&
          data[candidate + best] === data[i + best]
        ) {
          let n = 0;
          while (n < max && data[candidate + n] === data[i + n]) n++;
          if (n > best && n >= 3) {
            best = n;
            dist = i - candidate;
            if (n === max) break;
          }
        }
        const next = prev[candidate & 32767];
        if (next >= candidate) break;
        candidate = next;
      }
    }
    if (best >= 3) {
      let lc = 0;
      while (lc < 28 && LEN_BASE[lc + 1] <= best) lc++;
      sym(257 + lc);
      bits(best - LEN_BASE[lc], LEN_EXTRA[lc]);
      let dc = 0;
      while (dc < 29 && DIST_BASE[dc + 1] <= dist) dc++;
      bits(reverseBits(dc, 5), 5);
      bits(dist - DIST_BASE[dc], DIST_EXTRA[dc]);
      for (let j = 0; j < best; j++) insert(i + j);
      i += best;
    } else {
      sym(data[i]);
      insert(i);
      i++;
    }
  }
  sym(256);
  if (count) byte(buffer & 255);
  const sum = adler32(data);
  for (const b of u32(sum)) byte(b);
  return bytes.slice(0, len);
}
function paeth(a: number, b: number, c: number): number {
  const p = a + b - c,
    pa = Math.abs(p - a),
    pb = Math.abs(p - b),
    pc = Math.abs(p - c);
  return pa <= pb && pa <= pc ? a : pb <= pc ? b : c;
}
/** Encode RGBA pixels as PNG with optional uncompressed UTF-8 chart metadata. */
export function encodePNG(
  image: RGBAImage,
  metadata: object | null = null,
): Uint8Array<ArrayBuffer> {
  validateImage(image);
  check(
    image.width * image.height <= LIMITS.pixels,
    "PNG image exceeds output limit.",
  );
  const w = image.width,
    h = image.height,
    stride = w * 4,
    raw = new Uint8Array((stride + 1) * h),
    candidates = Array.from({ length: 5 }, () => new Uint8Array(stride));
  for (let y = 0; y < h; y++) {
    const scores = [0, 0, 0, 0, 0],
      off = y * stride;
    for (let x = 0; x < stride; x++) {
      const v = image.data[off + x],
        a = x >= 4 ? image.data[off + x - 4] : 0,
        b = y ? image.data[off + x - stride] : 0,
        c = y && x >= 4 ? image.data[off + x - stride - 4] : 0,
        predict = [0, a, b, Math.floor((a + b) / 2), paeth(a, b, c)];
      for (let f = 0; f < 5; f++) {
        const n = (v - predict[f]) & 255;
        candidates[f][x] = n;
        scores[f] += Math.min(n, 256 - n);
      }
    }
    let best = 0;
    for (let f = 1; f < 5; f++) if (scores[f] < scores[best]) best = f;
    raw[y * (stride + 1)] = best;
    raw.set(candidates[best], y * (stride + 1) + 1);
  }
  const ihdr = concat([u32(w), u32(h), new Uint8Array([8, 6, 0, 0, 0])]),
    parts = [
      new Uint8Array([137, 80, 78, 71, 13, 10, 26, 10]),
      chunk("IHDR", ihdr),
      chunk("sRGB", new Uint8Array([0])),
    ];
  if (metadata !== null) {
    const json = JSON.stringify(metadata);
    check(json.length <= 1048576, "PNG metadata is too large.");
    parts.push(
      chunk(
        "iTXt",
        concat([
          new TextEncoder().encode("chart"),
          new Uint8Array([0, 0, 0, 0, 0]),
          new TextEncoder().encode(json),
        ]),
      ),
    );
  }
  parts.push(chunk("IDAT", deflate(raw)), chunk("IEND", new Uint8Array(0)));
  return concat(parts);
}
function validateImage(im: RGBAImage): void {
  check(
    im &&
      Number.isInteger(im.width) &&
      Number.isInteger(im.height) &&
      im.width > 0 &&
      im.height > 0 &&
      im.width * im.height <= LIMITS.layerPixels &&
      im.data &&
      im.data.length === im.width * im.height * 4,
    "Expected an RGBA image with matching dimensions.",
  );
}
/** Draw image bytes without browser resampling and return the target canvas. */
export function drawImageToCanvas(image: RGBAImage, canvas: Canvas): Canvas {
  validateImage(image);
  check(
    canvas && typeof canvas.getContext === "function",
    "Expected a canvas element.",
  );
  const ctx = canvas.getContext("2d") as Context2D | null;
  check(ctx, "Canvas 2D context is unavailable.");
  canvas.width = image.width;
  canvas.height = image.height;
  const data = ctx.createImageData(image.width, image.height);
  data.data.set(image.data);
  ctx.putImageData(data, 0, 0);
  return canvas;
}
/** Download bytes in a browser; filenames may not contain path separators. */
export function downloadBytes(
  bytes: Uint8Array,
  name: string,
  type: string = "application/octet-stream",
): void {
  check(
    typeof document !== "undefined",
    "Downloads require a browser document.",
  );
  text(name, "Filename");
  check(
    !/[\\/\x00-\x1f]/.test(name),
    "Filename must not contain path separators or controls.",
  );
  const data =
    bytes.buffer instanceof ArrayBuffer
      ? (bytes as Uint8Array<ArrayBuffer>)
      : new Uint8Array(bytes);
  const blob = new Blob([data], { type }),
    url = URL.createObjectURL(blob),
    a = document.createElement("a");
  a.href = url;
  a.download = name;
  a.hidden = true;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 30000);
}
/** Rendered pixels and a frozen manifest; PNG encoding never requires Canvas. */
export class RenderResult<M extends object = GraphMetadata> {
  /** Pixels intentionally remain mutable; exports observe subsequent edits. */
  readonly image: RGBAImage;
  /** Immutable metadata captured at rendering time. */
  readonly metadata: Readonly<M>;
  constructor(image: RGBAImage, metadata: M) {
    this.image = image;
    this.metadata = freeze(metadata);
  }
  get width(): number {
    return this.image.width;
  }
  get height(): number {
    return this.image.height;
  }
  get data(): RGBAImage["data"] {
    return this.image.data;
  }
  draw(canvas: Canvas): Canvas {
    return drawImageToCanvas(this.image, canvas);
  }
  toPNG(options: PNGOptions = {}): Uint8Array<ArrayBuffer> {
    return encodePNG(
      this.image,
      options.metadata === false ? null : this.metadata,
    );
  }
  toBlob(options: PNGOptions = {}): Blob {
    return new Blob([this.toPNG(options)], { type: "image/png" });
  }
  download(filename: string = "chart.png", options: PNGOptions = {}): void {
    downloadBytes(this.toPNG(options), filename, "image/png");
  }
}
/** An immutable validated chart. Use with() to derive a changed chart. */
export class Chart {
  /** Fully normalized, deeply frozen options used by every render. */
  readonly config: DeepReadonly<ResolvedChartOptions>;
  constructor(options: ChartOptions = {}) {
    this.config = normalize(options);
    Object.freeze(this);
  }
  with(patch: ChartOptions): Chart {
    return new Chart(merge(this.config, patch));
  }
  render(): RenderResult {
    return renderChart(this.config);
  }
  draw(canvas: Canvas): RenderResult {
    const r = this.render();
    r.draw(canvas);
    return r;
  }
  toPNG(options: PNGOptions = {}): Uint8Array<ArrayBuffer> {
    return this.render().toPNG(options);
  }
  mount(canvas: HTMLCanvasElement, options: MountOptions = {}): Controller {
    return new BrowserController(this, canvas, options);
  }
  nearest(input: Timestamp): NearestSample[] {
    const time = epoch(input);
    const [start, end] = timeRange(this.config);
    return this.config.series.map((s) => {
      const lo = lowerBound(s.timestamps, start),
        hi = upperBound(s.timestamps, end);
      if (lo === hi)
        return { name: s.name, index: null, time: null, value: null };
      let i = clamp(lowerBound(s.timestamps, time), lo, hi - 1);
      if (
        i > lo &&
        Math.abs(s.timestamps[i - 1] - time) <= Math.abs(s.timestamps[i] - time)
      )
        i--;
      return {
        name: s.name,
        index: i,
        time: s.timestamps[i],
        value: finite(s.values[i]) ? s.values[i] : null,
      };
    });
  }
}
/** Build the conventional inbound area and outbound line without unit conversion. */
export function traffic(
  timestamps: ArrayLike<Timestamp>,
  inbound: Values,
  outbound: Values,
  options: TrafficOptions = {},
): Chart {
  const gap = options.gapAfter === undefined ? 0 : options.gapAfter,
    opts = { ...options };
  delete opts.gapAfter;
  return new Chart(
    merge(
      {
        title: "Traffic - ether1",
        verticalLabel: "bits per second",
        series: [
          series("Inbound", timestamps, inbound, {
            kind: "area",
            color: "#00cc00",
            outline: "#003000",
            gapAfter: gap,
          }),
          series("Outbound", timestamps, outbound, {
            kind: "line",
            color: "#0000cc",
            gapAfter: gap,
          }),
        ],
      },
      opts,
    ),
  );
}
/** Compose native chart renders and optional captions at a shared pixel scale. */
export function dashboard(
  panels: readonly (DashboardPanel | Chart)[],
  options: DashboardOptions = {},
): RenderResult<DashboardMetadata> {
  check(
    Array.isArray(panels) && panels.length > 0 && panels.length <= 128,
    "Dashboard requires 1..128 panels.",
  );
  const o = merge<Required<DashboardOptions>>(
    { gap: 24, padding: [4, 3, 6, 6], background: "#f3f3f3", cropHeight: null },
    options,
  );
  integer(o.gap, "Panel gap", 0, 4096);
  check(
    Array.isArray(o.padding) && o.padding.length === 4,
    "Padding is [left, top, right, bottom].",
  );
  o.padding.forEach((v) => integer(v, "Padding", 0, 4096));
  if (o.cropHeight !== null) integer(o.cropHeight, "Crop height", 1, 65536);
  const pp = panels.map((p) =>
    p instanceof Chart ? { chart: p, caption: "" } : p,
  );
  for (const p of pp) {
    check(p.chart instanceof Chart, "Panel needs a Chart.");
    text(p.caption || "", "Caption");
  }
  const scale = pp[0].chart.config.layout.pixelScale;
  check(
    pp.every((p) => p.chart.config.layout.pixelScale === scale),
    "Dashboard panels must have equal pixelScale.",
  );
  const rendered = pp.map((p) => p.chart.render()),
    left = o.padding[0] * scale,
    top = o.padding[1] * scale,
    gap = o.gap * scale;
  const width =
    Math.max(...rendered.map((r) => r.width)) +
    (o.padding[0] + o.padding[2]) * scale;
  let height =
    rendered.reduce((n, r) => n + r.height, 0) +
    (o.padding[1] + o.padding[3]) * scale +
    gap * (pp.length - 1 + (pp[pp.length - 1].caption ? 1 : 0));
  if (o.cropHeight !== null) height = Math.min(height, o.cropHeight * scale);
  check(width * height <= LIMITS.pixels, "Dashboard is too large.");
  const out = new Surface(width, height, color(o.background));
  let y = top;
  const meta: DashboardMetadata["panels"] = [];
  for (let i = 0; i < pp.length; i++) {
    const r = rendered[i],
      p = pp[i];
    out.over(r.image, left, y);
    meta.push({
      position: [left, y],
      caption: p.caption || "",
      chart: r.metadata,
    });
    y += r.height;
    if (p.caption) {
      const f = new Fonts(p.chart.config.fonts, p.chart.config.theme),
        w = f.width(p.caption, "caption"),
        h = p.chart.config.theme.captionSize;
      check(
        w <= width / scale - 8 && h + 4 <= o.gap,
        "Caption does not fit the dashboard gap.",
      );
      const line = new Surface(width / scale, o.gap);
      f.draw(
        line,
        width / scale / 2,
        4,
        p.caption,
        "caption",
        p.chart.config.theme.text,
        0,
        "center",
      );
      out.over(line.scale(scale), 0, y);
    }
    y += gap;
  }
  return new RenderResult<DashboardMetadata>(out, {
    version: VERSION,
    imageSize: [width, height],
    pixelScale: scale,
    panels: meta,
  });
}
/** Compute right-endpoint rates with exact BigInt counter subtraction. */
export function counterRate(
  timestamps: ArrayLike<Timestamp>,
  counters: ArrayLike<number | bigint | null | undefined>,
  options: CounterOptions = {},
): Samples {
  const o = merge<Required<CounterOptions>>(
    { factor: 1, onDecrease: "gap", counterBits: null, maxRate: null },
    options,
  );
  number(o.factor, "Rate factor", Number.MIN_VALUE);
  check(
    ["gap", "wrap"].includes(o.onDecrease),
    "Decrease policy must be gap or wrap.",
  );
  if (o.counterBits !== null) integer(o.counterBits, "Counter bits", 1, 128);
  if (o.maxRate !== null) number(o.maxRate, "Maximum rate", 0);
  check(
    o.onDecrease !== "wrap" || o.counterBits !== null,
    "Wrapping requires an explicit counterBits.",
  );
  check(
    (Array.isArray(counters) || ArrayBuffer.isView(counters)) &&
      counters.length === timestamps.length,
    "Counter and timestamp lengths must match.",
  );
  const s = samples(
      timestamps,
      Array.from(counters, () => 0),
    ),
    max = (1n << BigInt(o.counterBits || 128)) - 1n;
  const cs = Array.from(counters, (v, i) => {
    if (missing(v)) return null;
    if (typeof v === "number") {
      check(
        Number.isSafeInteger(v) && v >= 0,
        "Counter " +
          i +
          " must be a safe nonnegative integer; use BigInt for large counters.",
      );
      v = BigInt(v);
    }
    check(
      typeof v === "bigint" && v >= 0n && v <= max,
      "Counter must fit the configured unsigned bit width.",
    );
    return v;
  });
  s.values.fill(NaN);
  for (let i = 1; i < cs.length; i++) {
    const current = cs[i],
      previous = cs[i - 1];
    if (current === null || previous === null) continue;
    let d = current - previous;
    if (d < 0n) {
      if (o.onDecrease === "gap") continue;
      d += max + 1n;
    }
    const v = (Number(d) * o.factor) / (s.timestamps[i] - s.timestamps[i - 1]);
    check(finite(v), "Counter rate overflow.");
    if (o.maxRate === null || v <= o.maxRate) s.values[i] = v;
  }
  return s;
}
/** Aggregate elapsed-time buckets, preserving gaps and explicit coverage policy. */
export function aggregate(
  timestamps: ArrayLike<Timestamp>,
  values: Values,
  options: AggregateOptions = {},
): Samples {
  const o = merge<Required<AggregateOptions>>(
    {
      interval: 300,
      method: "mean",
      origin: 0,
      minCoverage: 0,
      expectedStep: null,
      maxBuckets: 1000000,
    },
    options,
  );
  number(o.interval, "Aggregation interval", 0.001);
  number(o.origin, "Aggregation origin");
  number(o.minCoverage, "Minimum coverage", 0, 1);
  if (o.expectedStep !== null)
    number(o.expectedStep, "Expected step", Number.MIN_VALUE);
  integer(o.maxBuckets, "Maximum buckets", 1, 1000000);
  check(
    ["mean", "min", "max", "last", "sum"].includes(o.method),
    "Unknown aggregation method.",
  );
  const s = samples(timestamps, values);
  if (!s.timestamps.length) return s;
  const bucket = (t: number): number => Math.floor((t - o.origin) / o.interval),
    a = bucket(s.timestamps[0]),
    b = bucket(s.timestamps[s.timestamps.length - 1]);
  check(
    Number.isSafeInteger(a) &&
      Number.isSafeInteger(b) &&
      b - a + 1 <= o.maxBuckets,
    "Aggregation bucket limit or precision exceeded.",
  );
  const out: Samples = { timestamps: [], values: [] };
  let i = 0;
  for (let k = a; k <= b; k++) {
    const ts = o.origin + k * o.interval;
    epoch(ts);
    out.timestamps.push(ts);
    let total = 0,
      last = NaN,
      mn = Infinity,
      mx = -Infinity;
    const finiteValues = [];
    while (i < s.timestamps.length && bucket(s.timestamps[i]) === k) {
      const v = s.values[i++];
      total++;
      last = v;
      if (finite(v)) {
        finiteValues.push(v);
        mn = Math.min(mn, v);
        mx = Math.max(mx, v);
      }
    }
    const denominator = Math.max(
      total,
      o.expectedStep === null ? 0 : o.interval / o.expectedStep,
    );
    if (
      !finiteValues.length ||
      finiteValues.length / denominator < o.minCoverage
    ) {
      out.values.push(NaN);
      continue;
    }
    let value;
    if (o.method === "mean") value = stableMean(finiteValues)!;
    else if (o.method === "min") value = mn;
    else if (o.method === "max") value = mx;
    else if (o.method === "last") value = last;
    else {
      value = stableMean(finiteValues)! * finiteValues.length;
      check(finite(value), "Aggregation sum overflow.");
    }
    out.values.push(value);
  }
  return out;
}
/** Parse quoted CSV into validated series; timestamps must strictly increase. */
export function parseCSV(input: string, options: CSVOptions = {}): Series[] {
  const o = merge<Required<CSVOptions>>(
    {
      timestampColumn: "timestamp",
      columns: [
        {
          column: "inbound",
          name: "Inbound",
          kind: "area",
          color: "#00cc00",
          outline: "#003000",
        },
        {
          column: "outbound",
          name: "Outbound",
          kind: "line",
          color: "#0000cc",
        },
      ],
      maxRows: 1000000,
      maxBytes: 16777216,
    },
    options,
  );
  check(typeof input === "string", "CSV must be a string.");
  integer(o.maxRows, "Maximum CSV rows", 1, 2000000);
  integer(o.maxBytes, "Maximum CSV bytes", 1, 67108864);
  check(
    input.length <= o.maxBytes &&
      new TextEncoder().encode(input).length <= o.maxBytes,
    "CSV exceeds the configured byte limit.",
  );
  check(
    Array.isArray(o.columns) && o.columns.length > 0 && o.columns.length <= 128,
    "CSV needs at least one column mapping.",
  );
  if (input.charCodeAt(0) === 0xfeff) input = input.slice(1);
  const rows: string[][] = [];
  let row: string[] = [],
    cell = "",
    quoted = false,
    closed = false;
  function endCell(): void {
    row.push(cell);
    cell = "";
    closed = false;
  }
  function endRow(): void {
    endCell();
    if (row.some((v) => v !== "")) rows.push(row);
    row = [];
    check(rows.length <= o.maxRows + 1, "CSV row limit exceeded.");
  }
  for (let i = 0; i < input.length; i++) {
    const ch = input[i];
    if (quoted) {
      if (ch === '"') {
        if (input[i + 1] === '"') {
          cell += '"';
          i++;
        } else {
          quoted = false;
          closed = true;
        }
      } else cell += ch;
      continue;
    }
    if (closed) {
      check(
        ch === "," || ch === "\r" || ch === "\n",
        "Unexpected text after a closing CSV quote.",
      );
    }
    if (ch === '"') {
      check(
        cell === "" && !closed,
        "Unexpected quote in an unquoted CSV field.",
      );
      quoted = true;
    } else if (ch === ",") endCell();
    else if (ch === "\n" || ch === "\r") {
      if (ch === "\r" && input[i + 1] === "\n") i++;
      endRow();
    } else cell += ch;
  }
  check(!quoted, "Unterminated quoted CSV field.");
  if (cell !== "" || row.length || closed) endRow();
  check(rows.length >= 1, "CSV header is missing.");
  const header = rows.shift()!.map((v) => v.trim());
  check(
    header.every(Boolean) && new Set(header).size === header.length,
    "CSV headers must be nonempty and unique.",
  );
  const ti = header.indexOf(o.timestampColumn);
  check(ti >= 0, "Timestamp column not found: " + o.timestampColumn);
  const ids = o.columns.map((v) => {
    const i = header.indexOf(v.column);
    check(i >= 0, "Column not found: " + v.column);
    return i;
  });
  const ts: number[] = [],
    vv: number[][] = o.columns.map(() => []);
  for (let i = 0; i < rows.length; i++) {
    const r = rows[i];
    check(
      r.length === header.length,
      "CSV row " + (i + 2) + " has the wrong number of columns.",
    );
    try {
      ts.push(epoch(r[ti]));
    } catch (e) {
      throw new RangeError(
        "CSV row " +
          (i + 2) +
          ": " +
          (e instanceof Error ? e.message : String(e)),
      );
    }
    ids.forEach((id, j) => {
      const v = r[id].trim();
      if (/^(?:nan|none|null)?$/i.test(v)) {
        vv[j].push(NaN);
        return;
      }
      check(
        /^[+-]?(?:\d+\.?\d*|\.\d+)(?:e[+-]?\d+)?$/i.test(v),
        "Invalid numeric CSV value at row " + (i + 2) + ".",
      );
      vv[j].push(number(Number(v), "CSV value"));
    });
  }
  return o.columns.map((v, j) => series(v.name || v.column, ts, vv[j], v));
}
/** Export the union of series timestamps, escaping formula-like text headers. */
export function toCSV(
  list: readonly SeriesInput[],
  options: { escapeFormulas?: boolean } = {},
): string {
  check(
    Array.isArray(list) && list.length <= 128,
    "Expected an array of series.",
  );
  const escaped = options.escapeFormulas !== false;
  const ss = list.map((s) => series(s.name, s.timestamps, s.values, s));
  let total = 0;
  for (const s of ss) total += s.timestamps.length;
  check(
    total <= LIMITS.samples,
    "CSV export exceeds the combined sample limit.",
  );
  const times = [...new Set(ss.flatMap((s) => s.timestamps))].sort(
    (a, b) => a - b,
  );
  check(
    new Set(ss.map((s) => s.name)).size === ss.length &&
      !ss.some((s) => s.name === "timestamp"),
    "CSV series names must be unique and not timestamp.",
  );
  const quote = (v: string): string => {
    v = String(v);
    if (escaped && /^[=+\-@\t\r]/.test(v)) v = "'" + v;
    return /[,"\r\n]/.test(v) ? '"' + v.replace(/"/g, '""') + '"' : v;
  };
  const lines = [["timestamp", ...ss.map((s) => s.name)].map(quote).join(",")],
    indexes = ss.map(() => 0);
  for (const t of times) {
    const cells = [String(t)];
    ss.forEach((s, j) => {
      while (indexes[j] < s.timestamps.length && s.timestamps[indexes[j]] < t)
        indexes[j]++;
      const i = indexes[j];
      cells.push(
        i < s.timestamps.length && s.timestamps[i] === t && finite(s.values[i])
          ? String(s.values[i])
          : "",
      );
    });
    lines.push(cells.join(","));
  }
  return lines.join("\r\n") + "\r\n";
}
/** Compare RGB pixels without aligning or resizing; alpha is ignored. */
export function compareImages(
  reference: RGBAImage | RenderResult<object>,
  actual: RGBAImage | RenderResult<object>,
  options: CompareOptions = {},
): PixelDifference {
  const a = reference instanceof RenderResult ? reference.image : reference,
    b = actual instanceof RenderResult ? actual.image : actual;
  validateImage(a);
  validateImage(b);
  check(
    a.width === b.width && a.height === b.height,
    "Image sizes must match; comparison never aligns or resizes.",
  );
  const tolerance = options.tolerance === undefined ? 0 : options.tolerance;
  integer(tolerance, "Tolerance", 0, 255);
  const box = options.box || [0, 0, a.width, a.height];
  check(
    Array.isArray(box) && box.length === 4,
    "Comparison box needs four coordinates.",
  );
  box.forEach((v) => integer(v, "Comparison coordinate", 0, 65536));
  const [x0, y0, x1, y1] = box;
  check(
    x1 > x0 && y1 > y0 && x1 <= a.width && y1 <= a.height,
    "Comparison box is out of bounds.",
  );
  let exact = 0,
    within = 0,
    sum = 0,
    square = 0,
    maxError = 0,
    loX = Infinity,
    loY = Infinity,
    hiX = -Infinity,
    hiY = -Infinity;
  for (let y = y0; y < y1; y++)
    for (let x = x0; x < x1; x++) {
      const i = (y * a.width + x) * 4;
      let err = 0;
      for (let k = 0; k < 3; k++) {
        const d = Math.abs(a.data[i + k] - b.data[i + k]);
        sum += d;
        square += d * d;
        err = Math.max(err, d);
      }
      maxError = Math.max(maxError, err);
      if (err === 0) exact++;
      else {
        loX = Math.min(loX, x - x0);
        loY = Math.min(loY, y - y0);
        hiX = Math.max(hiX, x - x0);
        hiY = Math.max(hiY, y - y0);
      }
      if (err <= tolerance) within++;
    }
  const count = (x1 - x0) * (y1 - y0);
  return {
    pixels: count,
    exactPixels: exact,
    exactRatio: exact / count,
    toleranceRatio: within / count,
    meanAbsoluteError: sum / (count * 3),
    rootMeanSquareError: Math.sqrt(square / (count * 3)),
    maxError,
    differenceBox: loX === Infinity ? null : [loX, loY, hiX + 1, hiY + 1],
  };
}
/** Visualize amplified absolute RGB differences in matching images. */
export function differenceImage(
  reference: RGBAImage | RenderResult<object>,
  actual: RGBAImage | RenderResult<object>,
  amplify: number = 4,
): RenderResult<{ imageSize: [number, number]; amplify: number }> {
  number(amplify, "Difference amplification", 0, 255);
  const a = reference instanceof RenderResult ? reference.image : reference,
    b = actual instanceof RenderResult ? actual.image : actual;
  compareImages(a, b);
  const out = new Surface(a.width, a.height);
  for (let i = 0; i < a.data.length; i += 4) {
    for (let k = 0; k < 3; k++)
      out.data[i + k] = Math.min(
        255,
        round(Math.abs(a.data[i + k] - b.data[i + k]) * amplify),
      );
    out.data[i + 3] = 255;
  }
  return new RenderResult(out, { imageSize: [out.width, out.height], amplify });
}
/** Copy Canvas2D pixels into independently owned RGBA storage. */
export function readCanvas(canvas: Canvas): RGBAImage {
  check(
    canvas && typeof canvas.getContext === "function",
    "Expected a canvas.",
  );
  const c = canvas.getContext("2d") as Context2D | null;
  check(c, "Canvas 2D is unavailable.");
  const data = c.getImageData(0, 0, canvas.width, canvas.height);
  const out = new Surface(data.width, data.height);
  out.data.set(data.data);
  return out;
}
/** Check PNG dimensions and decode with the browser image decoder. */
export async function decodeImage(blob: Blob): Promise<RGBAImage> {
  check(
    typeof Blob !== "undefined" && blob instanceof Blob,
    "decodeImage expects a Blob or File.",
  );
  check(blob.size <= 67108864, "Image file exceeds 64 MiB.");
  // Header dimensions are checked before a PNG is passed to the browser decoder.
  const head = new Uint8Array(await blob.slice(0, 24).arrayBuffer());
  check(
    head.length >= 24 &&
      [137, 80, 78, 71, 13, 10, 26, 10].every((v, i) => head[i] === v) &&
      head[12] === 73 &&
      head[13] === 72 &&
      head[14] === 68 &&
      head[15] === 82,
    "decodeImage accepts PNG files only.",
  );
  const headerView = new DataView(head.buffer),
    width = headerView.getUint32(16),
    height = headerView.getUint32(20);
  check(
    width > 0 && height > 0 && width * height <= LIMITS.pixels,
    "Decoded PNG exceeds the pixel limit.",
  );
  check(
    typeof createImageBitmap === "function",
    "This browser does not expose createImageBitmap.",
  );
  const bitmap = await createImageBitmap(blob);
  try {
    check(
      bitmap.width * bitmap.height <= LIMITS.pixels,
      "Decoded image exceeds the pixel limit.",
    );
    const canvas = createCanvas(bitmap.width, bitmap.height);
    const ctx = canvas.getContext("2d") as Context2D | null;
    check(ctx, "Canvas 2D is unavailable.");
    ctx.drawImage(bitmap, 0, 0);
    return readCanvas(canvas);
  } finally {
    bitmap.close();
  }
}
// Browser controller. Cursor and tooltip live on an overlay, never in exported pixels.
class BrowserController implements Controller {
  chart: Chart;
  readonly canvas: MountedCanvas;
  readonly options: MountOptions & { interactive: boolean };
  destroyed: boolean;
  result: RenderResult;
  private readonly handlers: [HTMLCanvasElement, string, EventListener][];
  private cursorTime: number | null;
  private readonly saved: {
    style: string | null;
    role: string | null;
    label: string | null;
    tab: string | null;
  };
  private readonly marker: Comment;
  private readonly wrap: HTMLSpanElement;
  private readonly overlay: HTMLCanvasElement;
  private readonly tip: HTMLDivElement;
  private readonly status: HTMLSpanElement;
  constructor(chart: Chart, canvas: MountedCanvas, options: MountOptions = {}) {
    check(
      typeof document !== "undefined" && canvas && canvas.ownerDocument,
      "mount requires an HTML canvas in a document.",
    );
    check(canvas.parentNode, "Attach the canvas to the document before mount.");
    check(
      !canvas.__bamtiGraphController,
      "A controller is already mounted on this canvas.",
    );
    this.chart = chart;
    this.canvas = canvas;
    this.options = { interactive: true, ...options };
    this.destroyed = false;
    this.handlers = [];
    this.cursorTime = null;
    this.result = chart.render();
    this.saved = {
      style: canvas.getAttribute("style"),
      role: canvas.getAttribute("role"),
      label: canvas.getAttribute("aria-label"),
      tab: canvas.getAttribute("tabindex"),
    };
    this.marker = document.createComment("canvas position");
    canvas.parentNode.insertBefore(this.marker, canvas);
    this.wrap = document.createElement("span");
    this.wrap.style.cssText =
      "position:relative;display:inline-block;vertical-align:top;line-height:0;max-width:none;";
    canvas.parentNode.insertBefore(this.wrap, canvas);
    this.wrap.appendChild(canvas);
    canvas.style.display = "block";
    canvas.style.maxWidth = "none";
    canvas.setAttribute("role", "img");
    this.overlay = document.createElement("canvas");
    this.overlay.setAttribute("aria-hidden", "true");
    this.overlay.style.cssText =
      "position:absolute;left:0;top:0;pointer-events:none;";
    this.wrap.appendChild(this.overlay);
    this.tip = document.createElement("div");
    this.tip.hidden = true;
    this.tip.setAttribute("role", "status");
    this.tip.style.cssText =
      "position:absolute;z-index:5;pointer-events:none;background:#111827;color:#fff;border:1px solid #4b5563;padding:9px 11px;border-radius:5px;font:11px/1.65 ui-monospace,monospace;white-space:pre;box-shadow:0 4px 18px #0003;text-align:left;";
    this.wrap.appendChild(this.tip);
    this.status = document.createElement("span");
    this.status.setAttribute("aria-live", "polite");
    this.status.style.cssText =
      "position:absolute;width:1px;height:1px;padding:0;margin:-1px;overflow:hidden;clip:rect(0,0,0,0);white-space:nowrap;";
    this.wrap.appendChild(this.status);
    this.apply(this.result);
    canvas.__bamtiGraphController = this;
    if (this.options.interactive) {
      canvas.tabIndex = 0;
      this.on(canvas, "pointermove", (e) => {
        const rect = canvas.getBoundingClientRect(),
          x = ((e.clientX - rect.left) * canvas.width) / rect.width,
          y = ((e.clientY - rect.top) * canvas.height) / rect.height,
          m = this.result.metadata,
          s = m.pixelScale,
          [l, t, r, b] = m.plotBox;
        if (x < l * s || x > r * s || y < t * s || y > b * s) {
          this.clear();
          return;
        }
        const time =
          m.timeRange[0] +
          ((x / s - l) / (r - l)) * (m.timeRange[1] - m.timeRange[0]);
        this.show(time, false);
      });
      this.on(canvas, "pointerleave", () => this.clear());
      this.on(canvas, "blur", () => this.clear());
      this.on(canvas, "keydown", (e) => {
        if (
          !["ArrowLeft", "ArrowRight", "Home", "End", "Escape"].includes(e.key)
        )
          return;
        e.preventDefault();
        if (e.key === "Escape") {
          this.clear();
          return;
        }
        const [a, b] = this.result.metadata.timeRange,
          s = this.chart.config.series.find((s) => s.timestamps.length),
          visible = s
            ? s.timestamps.slice(
                lowerBound(s.timestamps, a),
                upperBound(s.timestamps, b),
              )
            : [];
        let time;
        if (e.key === "Home") time = visible[0] === undefined ? a : visible[0];
        else if (e.key === "End")
          time = visible.length ? visible[visible.length - 1] : b;
        else if (visible.length) {
          let i =
            this.cursorTime === null
              ? e.key === "ArrowRight"
                ? -1
                : visible.length
              : lowerBound(visible, this.cursorTime);
          i = clamp(
            i + (e.key === "ArrowRight" ? 1 : -1) * (e.shiftKey ? 10 : 1),
            0,
            visible.length - 1,
          );
          time = visible[i];
        } else
          time = clamp(
            (this.cursorTime === null ? a : this.cursorTime) +
              ((e.key === "ArrowRight" ? 1 : -1) * (b - a)) / 100,
            a,
            b,
          );
        this.show(time, true);
      });
    }
  }
  on<K extends keyof HTMLElementEventMap>(
    target: HTMLCanvasElement,
    type: K,
    fn: (event: HTMLElementEventMap[K]) => void,
  ): void {
    target.addEventListener(type, fn);
    this.handlers.push([target, type, fn as EventListener]);
  }
  apply(r: RenderResult): void {
    this.result = r;
    r.draw(this.canvas);
    this.overlay.width = r.width;
    this.overlay.height = r.height;
    const m = r.metadata;
    this.canvas.setAttribute(
      "aria-label",
      this.options.ariaLabel ||
        [
          m.title || "Time-series chart",
          m.verticalLabel,
          ...m.statistics.map(
            (s) =>
              s.name +
              ": current " +
              (s.current === null
                ? "missing"
                : formatValue(s.current, m.yUnit, 2)),
          ),
        ].join(". "),
    );
    this.wrap.style.width = r.width + "px";
    this.wrap.style.height = r.height + "px";
    this.clear();
  }
  show(time: number, announce: boolean): void {
    if (this.destroyed) return;
    const m = this.result.metadata,
      s = m.pixelScale,
      [left, top, right, bottom] = m.plotBox,
      ctx = this.overlay.getContext("2d")!,
      x =
        (left +
          ((time - m.timeRange[0]) / (m.timeRange[1] - m.timeRange[0])) *
            (right - left)) *
        s;
    ctx.clearRect(0, 0, this.overlay.width, this.overlay.height);
    ctx.beginPath();
    ctx.strokeStyle = "#555";
    ctx.lineWidth = 1;
    ctx.setLineDash([2, 2]);
    ctx.moveTo(round(x) + 0.5, top * s);
    ctx.lineTo(round(x) + 0.5, bottom * s);
    ctx.stroke();
    const nearest = this.chart.nearest(time),
      rows = ["Nearest samples"];
    for (const p of nearest)
      rows.push(
        p.name +
          ": " +
          formatValue(p.value, m.yUnit, 2, this.chart.config.missingLabel) +
          (p.time !== null
            ? "  " + formatTime(p.time, m.timezone, "%H:%M:%S")
            : "  no observation"),
      );
    this.tip.textContent = rows.join("\n");
    this.tip.hidden = false;
    this.tip.style.left =
      Math.max(
        4,
        Math.min(this.result.width - this.tip.offsetWidth - 4, x + 12),
      ) + "px";
    this.tip.style.top = top * s + 8 + "px";
    this.cursorTime = time;
    if (announce) this.status.textContent = rows.join(". ");
    if (typeof this.options.onHover === "function")
      this.options.onHover({ time, samples: nearest });
  }
  clear(): void {
    if (this.overlay)
      this.overlay
        .getContext("2d")!
        .clearRect(0, 0, this.overlay.width, this.overlay.height);
    if (this.tip) this.tip.hidden = true;
    this.cursorTime = null;
  }
  update(patch: ChartOptions | Chart): RenderResult {
    check(!this.destroyed, "Controller has been destroyed.");
    const next = patch instanceof Chart ? patch : this.chart.with(patch),
      result = next.render();
    this.chart = next;
    this.apply(result);
    return result;
  }
  destroy(): void {
    if (this.destroyed) return;
    this.destroyed = true;
    for (const [el, type, fn] of this.handlers)
      el.removeEventListener(type, fn);
    this.handlers.length = 0;
    if (this.marker.parentNode) {
      this.marker.parentNode.insertBefore(this.canvas, this.marker);
      this.marker.remove();
    } else this.wrap.removeChild(this.canvas);
    this.wrap.remove();
    for (const [key, attr] of [
      ["style", "style"],
      ["role", "role"],
      ["label", "aria-label"],
      ["tab", "tabindex"],
    ] as const) {
      if (this.saved[key] === null) this.canvas.removeAttribute(attr);
      else this.canvas.setAttribute(attr, this.saved[key]);
    }
    delete this.canvas.__bamtiGraphController;
  }
}

/** Return a detached copy of every chart default. */
export function defaults(): ChartOptions {
  return clone(DEFAULTS);
}
/** Render a chart once; bitmap mode does not access browser APIs. */
export function render(options: ChartOptions = {}): RenderResult {
  return new Chart(options).render();
}
/** Frozen default export, equivalent to the corresponding named exports. */
export interface BamtiGraphAPI {
  readonly VERSION: typeof VERSION;
  readonly LIMITS: typeof LIMITS;
  readonly Chart: typeof Chart;
  readonly RenderResult: typeof RenderResult;
  readonly series: typeof series;
  readonly regularSeries: typeof regularSeries;
  readonly traffic: typeof traffic;
  readonly daily: typeof daily;
  readonly weekly: typeof weekly;
  readonly monthly: typeof monthly;
  readonly yearly: typeof yearly;
  readonly dashboard: typeof dashboard;
  readonly counterRate: typeof counterRate;
  readonly aggregate: typeof aggregate;
  readonly parseCSV: typeof parseCSV;
  readonly toCSV: typeof toCSV;
  readonly compareImages: typeof compareImages;
  readonly differenceImage: typeof differenceImage;
  readonly encodePNG: typeof encodePNG;
  readonly decodeImage: typeof decodeImage;
  readonly readCanvas: typeof readCanvas;
  readonly drawImageToCanvas: typeof drawImageToCanvas;
  readonly downloadBytes: typeof downloadBytes;
  readonly formatTime: typeof formatTime;
  readonly formatValue: typeof formatValue;
  readonly epoch: typeof epoch;
  readonly color: typeof color;
  readonly defaults: typeof defaults;
  readonly render: typeof render;
}

/** The complete BamtiGraph API; importing this module creates no global. */
const BamtiGraph: Readonly<BamtiGraphAPI> = Object.freeze({
  VERSION,
  LIMITS,
  Chart,
  RenderResult,
  series,
  regularSeries,
  traffic,
  daily,
  weekly,
  monthly,
  yearly,
  dashboard,
  counterRate,
  aggregate,
  parseCSV,
  toCSV,
  compareImages,
  differenceImage,
  encodePNG,
  decodeImage,
  readCanvas,
  drawImageToCanvas,
  downloadBytes,
  formatTime,
  formatValue,
  epoch,
  color,
  defaults,
  render,
});
export default BamtiGraph;
