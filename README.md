# BamtiGraph

Turn traffic, CPU, memory, and other measurements into clear time-series charts. Use BamtiGraph in a browser, from TypeScript, or from Go.

## What you can do

- Draw line and area charts from your own data.
- Show daily, weekly, monthly, or yearly views.
- Import CSV files and save charts as PNG images.
- Choose your colors, labels, and time zone.
- Keep charts unbranded or add your own text.

## See it in action

These images were rendered with BamtiGraph using the included sample data.

### Daily and weekly traffic

![Daily and weekly traffic charts](https://raw.githubusercontent.com/gosuda/BamtiGraph/v1.0.0/docs/images/traffic.png)

### CPU, memory, and temperature

![CPU, memory, and temperature charts](https://raw.githubusercontent.com/gosuda/BamtiGraph/v1.0.0/docs/images/metrics.png)

## Try the demo

From the project folder:

```sh
npm ci
npm run build
npm run serve
```

Open **http://127.0.0.1:8080**. You can change the chart settings, add a custom label, and export an image. The demo uses sample data, not live measurements.

## Use TypeScript

Install the npm package:

```sh
npm install bamtigraph
```

The Deno package name is `jsr:@safe/bantigraph`; JSR releases are published separately.

This Node.js example creates a chart and saves it as `traffic.png`:

```ts
import { writeFile } from "node:fs/promises";
import { traffic } from "bamtigraph";

const chart = traffic(
  [0, 300, 600],
  [180e6, 195e6, 188e6],
  [42e6, 45e6, 43e6],
  {
    title: "My network",
    fonts: { mode: "bitmap" },
  },
);

await writeFile("traffic.png", chart.toPNG());
```

Replace the arrays with your timestamps, incoming values, and outgoing values. Numeric timestamps are Unix seconds. Use `null` for a missing reading.

For charts on a web page, follow the [browser integration guide](docs/index.html#typescript).

## Use Go

Try the included example from the project folder:

```sh
go run ./examples/quickstart -o traffic.png
```

To create a chart from a CSV file:

```sh
go run ./cmd/bamtigraph -input examples/traffic.csv -o traffic.png
```

The Go package is `github.com/gosuda/BamtiGraph`. See the [Go quickstart source](examples/quickstart/main.go) to use it in your application. If Go cannot find a suitable font, add `-font /path/to/font.ttf`.

## Make it yours

No watermark is drawn by default. To add your own text along the right edge:

```ts
const branded = chart.with({ watermark: "My network" });
await writeFile("branded.png", branded.toPNG());
```

In Go, set `graph.Watermark = "My network"`, or pass `-watermark "My network"` to the command-line example. Use `BAMTIGRAPH` if you want the library name, or an empty string to remove the text.

In the demo, use **Custom edge text**. You can also change the surrounding page labels in `assets/brand.js`.

![Chart with an optional BAMTIGRAPH label](https://raw.githubusercontent.com/gosuda/BamtiGraph/v1.0.0/docs/images/custom-label.png)

## Explore more

- [Browser examples](docs/index.html#examples)
- [TypeScript options and API](docs/API.md)
- [Go usage guide](docs/index.html#go)
- [Release setup](docs/index.html#publishing)
- [Security notes](docs/SECURITY.md)

## License

BSD-3-Clause. See [LICENSE](LICENSE).
