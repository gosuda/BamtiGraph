(function () {
    'use strict';
    const T = BamtiGraph;
    const start = Date.UTC(2026, 8, 7, 3) / 1000;
    const timestamps = Array.from({ length: 289 }, (_, i) => start + i * 300);
    const inbound = timestamps.map((_, i) => (135 + 65 * Math.sin(i / 48) + 8 * Math.sin(i * 2.1)) * 1e6);
    const outbound = timestamps.map((_, i) => (42 + 12 * Math.sin(i / 37)) * 1e6);
    try {
        const chart = T.traffic(timestamps, inbound, outbound, {
            title: 'Traffic - ether1', timeAxis: T.daily('Asia/Seoul'),
            yAxis: { minimum: 0, maximum: 240e6, majorStep: 50e6 }
        });
        const result = chart.draw(document.getElementById('chart'));
        document.getElementById('save').addEventListener('click', () => result.download('traffic.png'));
        document.getElementById('status').textContent = result.width + ' × ' + result.height + ' pixels · synthetic data';
        window.exampleResult = result;
    }
    catch (error) {
        const box = document.getElementById('error');
        box.textContent = error.message;
        box.hidden = false;
    }
}());
