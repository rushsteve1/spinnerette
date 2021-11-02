(import spork/temple)

(defn render
  "Render a Markdown string to HTML using the full features
  of the Blackfriday renderer."
  [str]
  (spinternal/markdown str))

(defn render-unescaped
  "Render a Markdown string to HTML using a subset of the features
  of the Blackfriday renderer that is then unescaped.
  This is suitable for further processing, like using with Temple."
  [str]
  (spinternal/markdown-unescaped str))

(defn temple
  "Render a Markdown string to unescaped HTML and then process with
  Temple for advanced templating."
  [str &keys args]
  (let [md (spinternal/markdown-unescaped str)
        rend (temple/compile md)]
    (rend ;(kvs args))))

(defn add-loader
  "Enable the Markdown loader which allows you to import `.md` files.
  This will import a `render` function which takes arguments to be
  passed to Temple.
  This is functionally identical to how the native Markdown handler works."
  []
  (module/add-paths ".md" :markdown)
  (put module/loaders :markdown
    (fn [path]
      (let [md (slurp path)]
        { :render {
          :doc ""
          :source-map [path 0 0]
          :value (fn [&keys args] (temple md ;(kvs args))) } }))))