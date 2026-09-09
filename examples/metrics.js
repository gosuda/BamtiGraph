(function () {
    'use strict';
    try {
        const charts = DemoData.metrics(BamtiGraph), captions = ['Percent scale with a threshold rule', 'IEC byte prefixes (base 1024)', 'Signed range with a zero rule'], controllers = [];
        const root = document.getElementById('charts');
        charts.forEach((chart, i) => { const article = document.createElement('article'); article.className = 'panel'; const header = document.createElement('div'); header.className = 'panel-head'; const h = document.createElement('h2'); h.textContent = chart.config.title; header.appendChild(h); article.appendChild(header); const scroll = document.createElement('div'); scroll.className = 'plot-scroll'; const canvas = document.createElement('canvas'); scroll.appendChild(canvas); article.appendChild(scroll); const note = document.createElement('p'); note.className = 'plot-note'; note.textContent = captions[i]; article.appendChild(note); root.appendChild(article); controllers.push(chart.mount(canvas)); });
        document.getElementById('save').addEventListener('click', () => BamtiGraph.dashboard(charts.map((chart, i) => ({ chart, caption: captions[i] }))).download('metrics.png'));
        window.addEventListener('pagehide', e => { if (!e.persisted)
            controllers.forEach(c => c.destroy()); });
        window.exampleCharts = charts;
    }
    catch (error) {
        const box = document.getElementById('error');
        box.textContent = error.message;
        box.hidden = false;
    }
}());
