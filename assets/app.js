(function () {
    'use strict';
    const T = window.BamtiGraph, B = window.Presentation, D = window.DemoData;
    const $ = id => document.getElementById(id), controllers = [];
    let charts = [], ready = false;
    document.querySelectorAll('[data-brand]').forEach(el => { const value = B[el.dataset.brand]; if (typeof value === 'string')
        el.textContent = value; });
    document.title = B.pageTitle || 'Traffic overview';
    $('interface-name').value = B.interfaceName;
    $('watermark').value = B.watermark;
    if (![...$('timezone').options].some(option => option.value === B.timezone)) {
        const option = document.createElement('option');
        option.value = B.timezone;
        option.textContent = B.timezone;
        $('timezone').appendChild(option);
    }
    $('timezone').value = B.timezone;
    function report(error) { $('error').textContent = error.message || String(error); $('error').hidden = false; $('status').textContent = 'Settings were not applied.'; }
    function value(id, v) { const el = $(id); el.textContent = (v / 1e6).toFixed(2) + ' '; const unit = document.createElement('small'); unit.textContent = 'Mb/s'; el.appendChild(unit); }
    function updateTable(result) { const body = $('stats-body'); body.replaceChildren(); for (const s of result.metadata.statistics) {
        const tr = document.createElement('tr');
        for (const [i, v] of [s.name, s.current, s.average, s.maximum, s.count].entries()) {
            const el = document.createElement(i === 0 ? 'th' : 'td');
            if (i === 0)
                el.scope = 'row';
            el.textContent = i === 0 ? String(v) : i === 4 ? String(v) : (v === null ? 'NaN' : (v / 1e6).toFixed(2) + ' M');
            tr.appendChild(el);
        }
        body.appendChild(tr);
    } }
    function render() {
        $('error').hidden = true;
        const font = $('font-mode').value, scale = Number($('pixel-scale').value), zone = $('timezone').value, name = $('interface-name').value;
        const common = { title: 'Traffic - ' + name, watermark: $('watermark').value, fonts: { mode: font }, layout: { pixelScale: scale, legend: $('aligned-legend').checked ? 'aligned' : 'reference' } };
        const next = [D.makeChart(T, false, { ...common, timeAxis: T.daily(zone) }), D.makeChart(T, true, { ...common, timeAxis: T.weekly(zone) })];
        // Render both before updating either visible panel: invalid configurations keep the old view.
        const results = next.map(c => c.render());
        for (let i = 0; i < 2; i++) {
            if (controllers[i])
                controllers[i].destroy();
            controllers[i] = next[i].mount($(i === 0 ? 'daily-chart' : 'weekly-chart'));
            const m = results[i].metadata;
            $(i === 0 ? 'daily-range' : 'weekly-range').textContent = T.formatTime(m.timeRange[0], zone, '%d %b %Y, %H:%M') + ' — ' + T.formatTime(m.timeRange[1], zone, '%d %b %Y, %H:%M') + ' · ' + zone;
        }
        charts = next;
        const s = results[0].metadata.statistics;
        value('stat-in', s[0].current);
        value('stat-out', s[1].current);
        value('stat-avg', s[0].average);
        value('stat-peak', s[0].maximum);
        updateTable(results[0]);
        ready = true;
        $('status').textContent = 'Display updated. ' + results[0].width + ' × ' + results[0].height + ' px per panel.';
        window.demoState = { charts, controllers }; // Inspectable examples; the library itself has no app globals.
    }
    async function busy(button, task) { button.disabled = true; await new Promise(resolve => requestAnimationFrame(() => setTimeout(resolve, 0))); try {
        await task();
    }
    catch (e) {
        report(e);
    }
    finally {
        button.disabled = false;
    } }
    $('settings').addEventListener('submit', e => { e.preventDefault(); busy($('apply-settings'), render); });
    $('download-png').addEventListener('click', () => busy($('download-png'), () => { if (!ready)
        return; T.dashboard([{ chart: charts[0], caption: 'Daily (5-minute synthetic samples)' }, { chart: charts[1], caption: 'Weekly (30-minute synthetic samples)' }]).download('traffic-dashboard.png'); $('status').textContent = 'Dashboard PNG exported.'; }));
    $('download-json').addEventListener('click', () => { if (!ready)
        return; T.downloadBytes(new TextEncoder().encode(JSON.stringify(controllers.map(c => c.result.metadata), null, 2)), 'traffic-metadata.json', 'application/json'); });
    window.addEventListener('pagehide', e => { if (!e.persisted)
        controllers.forEach(c => c.destroy()); });
    try {
        render();
    }
    catch (e) {
        report(e);
    }
}());
