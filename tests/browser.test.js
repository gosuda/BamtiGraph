'use strict';
// Fixed regression checksum; update only after reviewing an intentional renderer change.
window.BITMAP_GOLDEN_FNV = 3361213400;
(async function () {
    const T = BamtiGraph, results = [];
    function ok(value, message = 'Assertion failed') { if (!value)
        throw new Error(message); }
    function eq(a, b, message = 'Values differ') { ok(JSON.stringify(a) === JSON.stringify(b), message + ': ' + JSON.stringify(a) + ' / ' + JSON.stringify(b)); }
    function throws(fn) { let failed = false; try {
        fn();
    }
    catch {
        failed = true;
    } ok(failed, 'Expected an exception'); }
    async function test(name, fn) { const start = performance.now(); try {
        await fn();
        results.push({ name, passed: true, milliseconds: Math.round(performance.now() - start) });
    }
    catch (e) {
        results.push({ name, passed: false, message: String(e.message || e) });
    } const li = document.createElement('li'); li.textContent = (results.at(-1).passed ? 'PASS · ' : 'FAIL · ') + name + (results.at(-1).message ? ' — ' + results.at(-1).message : ''); document.querySelector('#checks').append(li); }
    const canvas = document.querySelector('#fixture-canvas'), parent = canvas.parentNode;
    const native = DemoData.makeChart(T, false), bitmap = native.with({ fonts: { mode: 'bitmap' }, timeAxis: T.daily('UTC') });
    let result, controller, hovered;
    await test('System-font render has exact reference geometry', () => { result = native.render(); eq(result.metadata.imageSize, [595, 211]); eq(result.metadata.plotBox, [64, 34, 564, 156]); eq(result.metadata.watermark, ''); ok(result.data.every((v, i) => i % 4 !== 3 || v === 255)); });
    await test('Repeated system-font output is stable within this browser', () => eq(T.compareImages(result, native.render()).exactRatio, 1));
    await test('Canvas draw preserves all exported pixels', () => { result.draw(canvas); eq(T.compareImages(result, T.readCanvas(canvas)).exactRatio, 1); });
    await test('PNG is independently decoded by the browser', async () => { const decoded = await T.decodeImage(result.toBlob()); eq(T.compareImages(result, decoded).exactRatio, 1); });
    await test('Invalid image and oversize PNG headers are rejected before decoding', async () => { let rejects = 0; try {
        await T.decodeImage(new Blob(['not an image']));
    }
    catch {
        rejects++;
    } const fake = new Uint8Array(24); fake.set([137, 80, 78, 71, 13, 10, 26, 10, 0, 0, 0, 13, 73, 72, 68, 82]); new DataView(fake.buffer).setUint32(16, 100000); new DataView(fake.buffer).setUint32(20, 100000); try {
        await T.decodeImage(new Blob([fake]));
    }
    catch {
        rejects++;
    } eq(rejects, 2); });
    await test('Bitmap pixels match the Node regression fixture', () => { const r = bitmap.render(); let hash = 2166136261; for (const b of r.data) {
        hash ^= b;
        hash = Math.imul(hash, 16777619) >>> 0;
    } eq(hash, window.BITMAP_GOLDEN_FNV); });
    await test('2x bitmap output repeats each original RGBA pixel exactly', () => { const a = bitmap.render(), b = bitmap.with({ layout: { pixelScale: 2 } }).render(); for (let y = 0; y < a.height; y++)
        for (let x = 0; x < a.width; x++)
            for (let dy = 0; dy < 2; dy++)
                for (let dx = 0; dx < 2; dx++)
                    for (let k = 0; k < 4; k++)
                        if (a.data[(y * a.width + x) * 4 + k] !== b.data[((2 * y + dy) * b.width + 2 * x + dx) * 4 + k])
                            throw new Error('Pixel mismatch'); });
    await test('Mount provides keyboard access and sample descriptions', () => { controller = native.mount(canvas, { onHover: event => { hovered = event; } }); eq(canvas.getAttribute('role'), 'img'); eq(canvas.tabIndex, 0); ok(canvas.getAttribute('aria-label').includes('Inbound')); });
    await test('Duplicate mount is rejected without adding a second controller', () => throws(() => native.mount(canvas)));
    await test('Keyboard navigation announces actual nearest samples', () => { canvas.dispatchEvent(new KeyboardEvent('keydown', { key: 'Home', bubbles: true })); ok(!canvas.parentElement.querySelector('[role="status"]').hidden); eq(hovered.time, controller.result.metadata.timeRange[0]); eq(hovered.samples, controller.chart.nearest(hovered.time)); ok(canvas.parentElement.querySelector('[aria-live="polite"]').textContent.includes('Inbound')); canvas.dispatchEvent(new KeyboardEvent('keydown', { key: 'End', bubbles: true })); eq(hovered.time, controller.result.metadata.timeRange[1]); });
    await test('Interactive overlay never changes PNG or base canvas pixels', () => { eq(T.compareImages(controller.result, T.readCanvas(canvas)).exactRatio, 1); eq(T.compareImages(result, controller.result).exactRatio, 1); });
    await test('Escape hides inspection without changing base pixels', () => { canvas.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true })); ok(canvas.parentElement.querySelector('[role="status"]').hidden); eq(T.compareImages(controller.result, T.readCanvas(canvas)).exactRatio, 1); });
    await test('Invalid updates leave the previous render intact', () => { const previous = controller.result; throws(() => controller.update({ yAxis: { minimum: 2, maximum: 1 } })); ok(previous === controller.result); eq(T.compareImages(previous, T.readCanvas(canvas)).exactRatio, 1); });
    await test('Valid updates replace chart and export metadata', () => { controller.update({ title: 'Updated title' }); eq(controller.result.metadata.title, 'Updated title'); });
    await test('Destroy restores the canvas and allows remounting', () => { controller.destroy(); eq(canvas.parentNode, parent); eq(canvas.getAttribute('style'), 'display:block'); eq(canvas.getAttribute('aria-label'), 'Original canvas'); eq(canvas.getAttribute('tabindex'), null); controller.destroy(); throws(() => controller.update({ title: 'Invalid' })); const again = native.mount(canvas); again.destroy(); eq(canvas.parentNode, parent); });
    await test('Non-interactive mount does not inject a keyboard tab stop', () => { const c = native.mount(canvas, { interactive: false }); eq(canvas.getAttribute('tabindex'), null); c.destroy(); });
    await test('Native text accepts installed-script characters without HTML injection', () => { const title = '<img src=x onerror=alert(1)> \uD55C\uAE00'; const c = native.with({ title }).mount(canvas); eq(c.result.metadata.title, title); ok(!document.querySelector('img')); c.destroy(); });
    await test('CSV round-trip preserves nulls and numerical values', () => { const s = T.series('A', [0, 60, 120], [1, null, 3]); const back = T.parseCSV(T.toCSV([s]), { columns: [{ column: 'A', name: 'A', kind: 'line' }] }); eq(back[0].values.map(v => Number.isNaN(v) ? null : v), [1, null, 3]); });
    await test('BigInt subtraction preserves counter increments beyond 2^53', () => { const n = 2n ** 63n, r = T.counterRate([0, 2], [n, n + 7n], { factor: 8 }); eq(r.values[1], 28); });
    await test('IANA fall transition retains the two local 01:00 instants', () => { eq(T.formatTime(T.epoch('2024-11-03T05:00:00Z'), 'America/New_York', '%H:%M %z'), '01:00 -0400'); eq(T.formatTime(T.epoch('2024-11-03T06:00:00Z'), 'America/New_York', '%H:%M %z'), '01:00 -0500'); });
    await test('Dashboard composition preserves independent chart metadata', () => { const r = T.dashboard([{ chart: native, caption: 'Daily' }, { chart: DemoData.makeChart(T, true), caption: 'Weekly' }]); eq([r.width, r.height], [605, 479]); eq(r.metadata.panels.length, 2); });
    const failed = results.filter(r => !r.passed).length;
    window.browserTestReport = { environment: navigator.userAgent, devicePixelRatio: devicePixelRatio, passed: results.length - failed, failed, total: results.length, checks: results };
    document.querySelector('#summary').textContent = `${results.length - failed} passed · ${failed} failed`;
})();
