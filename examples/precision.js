(function () {
    'use strict';
    const T = BamtiGraph, $ = id => document.getElementById(id);
    let baseChart, base, doubled;
    function run() {
        const repeat = baseChart.render(), diff = T.compareImages(base, repeat);
        let exactScale = true;
        outer: for (let y = 0; y < doubled.height; y++)
            for (let x = 0; x < doubled.width; x++) {
                const i = (y * doubled.width + x) * 4, j = (Math.floor(y / 2) * base.width + Math.floor(x / 2)) * 4;
                for (let k = 0; k < 4; k++)
                    if (doubled.data[i + k] !== base.data[j + k]) {
                        exactScale = false;
                        break outer;
                    }
            }
        const deterministic = diff.exactRatio === 1;
        document.getElementById('results').textContent = (deterministic ? 'PASS' : 'FAIL') + ' · repeated render ' + (diff.exactRatio * 100).toFixed(2) + '% exact RGB pixels; ' + (exactScale ? 'PASS' : 'FAIL') + ' · integer 2× expansion.';
        window.precisionResult = { deterministic, exactScale, difference: diff };
    }
    try {
        baseChart = DemoData.makeChart(T, false, { fonts: { mode: 'bitmap' }, timeAxis: T.daily('UTC') });
        base = baseChart.draw($('base'));
        doubled = baseChart.with({ layout: { pixelScale: 2 } }).draw($('double'));
        $('run').addEventListener('click', run);
        $('save').addEventListener('click', () => base.download('pixel-baseline.png'));
        run();
    }
    catch (e) {
        $('error').textContent = e.message;
        $('error').hidden = false;
    }
}());
