// Compile-time consumer check; not part of the runtime test entry point.
import T, { Chart, type Series, RenderResult, type GraphMetadata } from 'bamtigraph';
const timestamps = [0, 300, 600];
const source: Series = T.series('Inbound', timestamps, [1, null, 3], { kind: 'area' });
const chart: Chart = new T.Chart({ series: [source], fonts: { mode: 'bitmap' } });
const rendered: RenderResult<GraphMetadata> = chart.render();
const png: Uint8Array = rendered.toPNG();
const scaled = chart.with({ layout: { pixelScale: 2 } });
const comparison: number = T.compareImages(rendered, chart.render(), { tolerance: 2 }).exactRatio;
const rates = T.counterRate(timestamps, [2n ** 63n, 2n ** 63n + 1n, null], { factor: 8 });
const board = T.dashboard([{ chart, caption: 'Sample' }]);
const metadata = board.metadata.panels[0].chart.title;
const c = T.traffic(timestamps, rates.values, [1, 2, 3], { timeAxis: T.daily('Asia/Seoul') });
if (typeof document !== 'undefined') {
  const canvas = document.createElement('canvas');
  document.body.append(canvas);
  const controller = c.mount(canvas, { onHover: e => console.log(e.samples[0].value) });
  controller.update({ title: 'Title' });
  controller.destroy();
}
void png; void comparison; void metadata;

void scaled;
