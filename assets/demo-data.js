/* Deterministic sample fixtures. All values below are synthetic, not measurements. */
(function (root) {
    'use strict';
    function rng(seed) { let x = seed >>> 0; return () => { x ^= x << 13; x ^= x >>> 17; x ^= x << 5; return (x >>> 0) / 4294967296; }; }
    function mix(t, knots) { for (let i = 1; i < knots.length; i++)
        if (t <= knots[i][0]) {
            const a = knots[i - 1], b = knots[i];
            return a[1] + (b[1] - a[1]) * (t - a[0]) / (b[0] - a[0]);
        } return knots[knots.length - 1][1]; }
    function traffic(weekly = false) {
        const count = weekly ? 337 : 289, step = weekly ? 1800 : 300, end = Date.UTC(2026, 8, 8, 3, 0, 0) / 1000, start = end - (count - 1) * step, random = rng(weekly ? 9366 : 7312), ts = [], inbound = [], outbound = [];
        const knots = [[0, 182], [.07, 211], [.18, 214], [.25, 184], [.35, 187], [.36, 147], [.43, 201], [.5, 177], [.54, 116], [.64, 56], [.76, 33], [.87, 108], [.95, 155], [.983, 216], [1, 174]];
        for (let i = 0; i < count; i++) {
            const p = i / (count - 1);
            let a, b;
            if (weekly) {
                const local = ((start + i * step + 9 * 3600) % 86400) / 86400, day = Math.floor(i / 48), peak = 152 + day * 7;
                a = 26 + peak * Math.max(0, Math.sin((local - .2) * Math.PI * 1.9));
                a += 13 * Math.sin(local * 25) + 15 * (random() - .5);
                b = 24 + a * .14 + 6 * Math.sin(i / 9) + 8 * (random() - .5);
            }
            else {
                a = mix(p, knots) * (0.9 + random() * .16);
                b = 49 + 8 * Math.sin(p * 8) - 28 * Math.exp(-(((p - .7) / .07) ** 2)) + 5 * Math.sin(i / 12) + 8 * (random() - .5);
            }
            ts.push(start + i * step);
            inbound.push(Math.max(0, a) * 1e6);
            outbound.push(Math.max(0, b) * 1e6);
        }
        return { timestamps: ts, inbound, outbound, step, start, end };
    }
    function makeChart(T, weekly = false, options = {}) { const d = traffic(weekly); return T.traffic(d.timestamps, d.inbound, d.outbound, { timeAxis: weekly ? T.weekly('Asia/Seoul') : T.daily('Asia/Seoul'), yAxis: { minimum: 0, maximum: 237e6, majorStep: 50e6 }, ...options }); }
    function metrics(T, options = {}) {
        const d = traffic(false), n = d.timestamps.length, random = rng(1701), ts = d.timestamps;
        const cpu = Array.from({ length: n }, (_, i) => Math.max(8, 46 + 24 * Math.sin(i / 33) + 15 * (random() - .5))), wait = cpu.map((v, i) => 4 + 3 * Math.sin(i / 21));
        const mem = Array.from({ length: n }, (_, i) => (9 + 1.2 * Math.sin(i / 43) + (i / n)) * 1024 ** 3), cache = mem.map((v, i) => (2.1 + .6 * Math.cos(i / 26)) * 1024 ** 3);
        const temp = Array.from({ length: n }, (_, i) => 14 * Math.sin(i / 35) - 2), setpoint = temp.map(() => 8);
        const common = { timeAxis: T.daily('Asia/Seoul'), ...options };
        return [
            new T.Chart({ ...common, title: 'CPU utilization', verticalLabel: 'percent', series: [T.series('CPU', ts, cpu, { kind: 'area', color: '#00cc00', outline: '#003000' }), T.series('IO wait', ts, wait)], yAxis: { minimum: 0, maximum: 100, majorStep: 20, scaleFactor: 1, suffix: '%' }, hRules: [{ value: 85, color: '#990000', dash: [3, 2] }] }),
            new T.Chart({ ...common, title: 'Memory allocation', verticalLabel: 'bytes', series: [T.series('Used', ts, mem, { kind: 'area', color: '#00cc00', outline: '#003000' }), T.series('Cache', ts, cache)], yAxis: { minimum: 0, maximum: 16 * 1024 ** 3, majorStep: 4 * 1024 ** 3, base: 1024 } }),
            new T.Chart({ ...common, title: 'Temperature deviation', verticalLabel: 'degrees C', series: [T.series('Observed', ts, temp, { kind: 'area', color: '#00cc00', outline: '#003000' }), T.series('Reference', ts, setpoint, { interpolation: 'step-post' })], yAxis: { minimum: -20, maximum: 20, majorStep: 10, scaleFactor: 1, suffix: 'C' }, hRules: [{ value: 0, color: '#555555', dash: null }] })
        ];
    }
    root.DemoData = Object.freeze({ traffic, makeChart, metrics, rng });
}(typeof globalThis !== 'undefined' ? globalThis : this));
