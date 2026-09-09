'use strict';
const { test } = require('node:test'), assert = require('node:assert/strict'), net = require('node:net'), http = require('node:http'), { spawn } = require('node:child_process'), path = require('node:path'), fs = require('node:fs'), os = require('node:os');
const root = path.resolve(__dirname, '..');
function availablePort() { return new Promise((resolve, reject) => { const server = net.createServer(); server.on('error', reject); server.listen(0, '127.0.0.1', () => { const port = server.address().port; server.close(() => resolve(port)); }); }); }
function request(port, target, method = 'GET') { return new Promise((resolve, reject) => { const req = http.request({ hostname: '127.0.0.1', port, path: target, method }, res => { let body = ''; res.setEncoding('utf8'); res.on('data', s => body += s); res.on('end', () => resolve({ status: res.statusCode, headers: res.headers, body })); }); req.setTimeout(3000, () => req.destroy(new Error('Request timeout'))); req.on('error', reject); req.end(); }); }
test('Optional static server: loopback, MIME, methods, traversal and missing files', async (t) => {
    const port = await availablePort(), child = spawn(process.execPath, [path.join(root, 'tools/serve.cjs'), '--port=' + port], { stdio: ['ignore', 'pipe', 'pipe'] });
    t.after(() => child.kill());
    await new Promise((resolve, reject) => { const timer = setTimeout(() => reject(new Error('Server startup timeout')), 6000); child.stdout.once('data', () => { clearTimeout(timer); resolve(); }); child.once('error', e => { clearTimeout(timer); reject(e); }); child.once('exit', code => { clearTimeout(timer); if (code)
        reject(new Error('Server exit ' + code)); }); });
    const index = await request(port, '/');
    assert.equal(index.status, 200);
    assert.equal(index.headers['x-content-type-options'], 'nosniff');
    const module = await request(port, '/dist/index.js', 'HEAD');
    assert.equal(module.status, 200);
    assert.ok(module.headers['content-type'].startsWith('text/javascript'));
    assert.equal(module.body, '');
    assert.equal((await request(port, '/index.html', 'POST')).status, 405);
    assert.equal((await request(port, '/missing-file')).status, 404);
    assert.equal((await request(port, '/%2e%2e%2fREADME.md')).status, 403);
    assert.equal((await request(port, '/%00')).status, 400);
    const external = fs.mkdtempSync(path.join(os.tmpdir(), 'chart-server-')), link = path.join(root, 'tests', '.test-external-link');
    t.after(() => { try {
        fs.unlinkSync(link);
    }
    catch { } fs.rmSync(external, { recursive: true, force: true }); });
    fs.writeFileSync(path.join(external, 'private.txt'), 'Outside root');
    try {
        fs.symlinkSync(external, link, 'dir');
        assert.equal((await request(port, '/tests/.test-external-link/private.txt')).status, 403);
    }
    catch (e) {
        if (['EPERM', 'EACCES'].includes(e.code))
            t.diagnostic('Host does not permit symlink creation; symlink route not exercised.');
        else
            throw e;
    }
});
