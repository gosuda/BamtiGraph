(function () {
    'use strict';
    const $ = id => document.getElementById(id), T = BamtiGraph, MAX = 301, step = 2;
    let random, ts, inbound, outbound, controller, timer = null, running = false, tick = 0;
    function append() { const t = ts.length ? ts[ts.length - 1] + step : Date.UTC(2026, 8, 8, 3) / 1000; ts.push(t); inbound.push((115 + 55 * Math.sin(tick / 41) + 18 * random()) * 1e6); outbound.push((35 + 13 * Math.sin(tick / 27) + 5 * random()) * 1e6); tick++; if (ts.length > MAX) {
        ts.shift();
        inbound.shift();
        outbound.shift();
    } }
    function chart() { return T.traffic(ts, inbound, outbound, { title: 'Traffic - simulated interface', timeAxis: { mode: 'auto', timezone: 'UTC' }, yAxis: { minimum: 0, maximum: 220e6, majorStep: 50e6 } }); }
    function report(e) { pause(); $('error').textContent = e.message; $('error').hidden = false; }
    function render() { if (controller)
        controller.update(chart());
    else
        controller = chart().mount($('chart')); $('status').textContent = ts.length + ' / ' + MAX + ' samples · ' + T.formatTime(ts[ts.length - 1], 'UTC', '%H:%M:%S') + ' simulated UTC'; }
    function cycle() { if (!running || document.hidden)
        return; try {
        append();
        render();
        timer = setTimeout(cycle, 1000);
    }
    catch (e) {
        report(e);
    } }
    function pause() { running = false; if (timer !== null)
        clearTimeout(timer); timer = null; $('toggle').textContent = 'Start simulation'; $('state').textContent = 'Paused'; }
    function reset() { pause(); random = DemoData.rng(451); ts = []; inbound = []; outbound = []; tick = 0; for (let i = 0; i < MAX; i++)
        append(); render(); }
    $('toggle').addEventListener('click', () => { if (running) {
        pause();
        return;
    } running = true; $('toggle').textContent = 'Pause simulation'; $('state').textContent = 'Simulating'; cycle(); });
    $('reset').addEventListener('click', () => { try {
        reset();
    }
    catch (e) {
        report(e);
    } });
    $('save').addEventListener('click', () => controller.result.download('simulated-traffic.png'));
    document.addEventListener('visibilitychange', () => { if (document.hidden && running)
        pause(); });
    window.addEventListener('pagehide', e => { pause(); if (!e.persisted && controller)
        controller.destroy(); });
    try {
        reset();
    }
    catch (e) {
        report(e);
    }
    window.liveExample = { get count() { return ts.length; }, get running() { return running; }, pause };
}());
