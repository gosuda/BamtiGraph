#!/usr/bin/env node
'use strict';
// Optional loopback-only static server. No packages or installation step needed.
const http = require('node:http'), fs = require('node:fs'), path = require('node:path');
const root = fs.realpathSync(path.resolve(__dirname, '..'));
const arg = process.argv.find(v => /^--port=/.test(v));
const port = arg ? Number(arg.slice(7)) : 8080;
if (!Number.isInteger(port) || port < 1 || port > 65535) {
    console.error('Use --port=8080 with a port in 1..65535.');
    process.exit(2);
}
const types = { '.html': 'text/html; charset=utf-8', '.js': 'text/javascript; charset=utf-8', '.mjs': 'text/javascript; charset=utf-8', '.css': 'text/css; charset=utf-8', '.json': 'application/json; charset=utf-8', '.png': 'image/png', '.csv': 'text/csv; charset=utf-8', '.md': 'text/plain; charset=utf-8', '.txt': 'text/plain; charset=utf-8' };
const server = http.createServer((req, res) => {
    function error(code, message) { res.writeHead(code, { 'Content-Type': 'text/plain; charset=utf-8', 'X-Content-Type-Options': 'nosniff' }); res.end(message); }
    if (req.method !== 'GET' && req.method !== 'HEAD') {
        error(405, 'GET and HEAD only.');
        return;
    }
    try {
        const url = new URL(req.url, 'http://localhost'), pathname = decodeURIComponent(url.pathname);
        if (pathname.includes('\0') || pathname.includes('\\')) {
            error(400, 'Invalid path.');
            return;
        }
        let file = path.resolve(root, '.' + pathname);
        if (file !== root && !file.startsWith(root + path.sep)) {
            error(403, 'Forbidden.');
            return;
        }
        if (fs.statSync(file).isDirectory())
            file = path.join(file, 'index.html');
        file = fs.realpathSync(file);
        if (file !== root && !file.startsWith(root + path.sep)) {
            error(403, 'Forbidden.');
            return;
        }
        const stat = fs.statSync(file);
        if (!stat.isFile()) {
            error(404, 'Not found.');
            return;
        }
        res.writeHead(200, { 'Content-Type': types[path.extname(file)] || 'application/octet-stream', 'Content-Length': stat.size, 'X-Content-Type-Options': 'nosniff', 'Cache-Control': 'no-store' });
        if (req.method === 'HEAD')
            res.end();
        else {
            const stream = fs.createReadStream(file);
            stream.on('error', () => res.destroy());
            stream.pipe(res);
        }
    }
    catch (e) {
        error(e.code === 'ENOENT' ? 404 : 400, 'Resource unavailable.');
    }
});
server.on('error', e => { console.error(e.message); process.exitCode = 1; });
server.listen(port, '127.0.0.1', () => console.log('Open http://127.0.0.1:' + port + ' · Ctrl+C to stop'));
