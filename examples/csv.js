(function () {
    'use strict';
    const T = BamtiGraph, $ = id => document.getElementById(id);
    let controller;
    const d = DemoData.traffic(false), timestamps = d.timestamps.filter((_, i) => i % 6 === 0), inbound = d.inbound.filter((_, i) => i % 6 === 0), outbound = d.outbound.filter((_, i) => i % 6 === 0);
    const sample = T.toCSV([T.series('inbound', timestamps, inbound), T.series('outbound', timestamps, outbound)]);
    function error(e) { $('error').textContent = e.message; $('error').hidden = false; $('status').textContent = 'Invalid input. The previous chart has not changed.'; }
    function render() {
        try {
            const rows = T.parseCSV($('csv-text').value), chart = new T.Chart({ title: 'Traffic - CSV input', verticalLabel: 'bits per second', series: rows, timeAxis: { timezone: 'Asia/Seoul' } });
            if (controller)
                controller.update(chart);
            else
                controller = chart.mount($('chart'));
            $('error').hidden = true;
            $('status').textContent = rows[0].timestamps.length + ' rows parsed locally · values displayed as provided.';
        }
        catch (e) {
            error(e);
        }
    }
    $('render').addEventListener('click', render);
    $('sample').addEventListener('click', () => { $('csv-text').value = sample; render(); });
    $('download').addEventListener('click', () => T.downloadBytes(new TextEncoder().encode(sample), 'traffic-sample.csv', 'text/csv;charset=utf-8'));
    $('save').addEventListener('click', () => { if (controller)
        controller.result.download('csv-chart.png'); });
    $('file').addEventListener('change', async (e) => { const file = e.target.files[0]; if (!file)
        return; try {
        if (file.size > 16 * 1024 * 1024)
            throw new Error('CSV exceeds 16 MiB.');
        $('csv-text').value = await file.text();
        render();
    }
    catch (e) {
        error(e);
    } });
    window.addEventListener('pagehide', e => { if (!e.persisted && controller)
        controller.destroy(); });
    $('csv-text').value = sample;
    render();
    window.csvExample = { render, get result() { return controller && controller.result; } };
}());
