// Conditional exports must resolve real CommonJS declarations, not an ESM-only file.
import BamtiGraph = require('bamtigraph');
const chart: BamtiGraph.Chart = BamtiGraph.traffic([0, 300], [1, 2], [3, null], {
  fonts: { mode: 'bitmap' },
  watermark: 'Customer label',
});
const result: BamtiGraph.RenderResult<BamtiGraph.GraphMetadata> = chart.render();
const png: Uint8Array = result.toPNG();
void png;
// @ts-expect-error Unix seconds cannot be arbitrary objects.
BamtiGraph.series('invalid', [{}], [1]);
