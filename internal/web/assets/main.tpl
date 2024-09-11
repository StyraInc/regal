<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8"/>
  <link rel="stylesheet" href="/assets/missing.min.css">
  <script src="/assets/htmx-1.8.4.min.js"></script>
</head>
<body>
<main class="crowded">
  <form>
    <div class="f-row">
					<textarea name="code"
                    style="height: 300px"
                    hx-post="/?tmpl=output"
                    hx-target="#output"
                    hx-trigger="keyup changed delay:200ms"
                    class="flex-grow:1 monospace"
          >{{ .Code }}</textarea>
    </div>
  </form>
  <section id="output">
    {{ block "output" . }}
    {{ range .Result }}
    <details class={{ .Class }} {{ if .Show }}open{{ end }}>
      <summary>{{ .Stage }}</summary>
      <pre><code>{{ .Output }}</code></pre>
    </details>
    {{ end }}
    {{ end }}
  </section>
</main>
</body>
</html>
