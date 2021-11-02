(import spork/temple)

(defn render [str]
  (spinternal/markdown str))

(defn render-unescaped [str]
  (spinternal/markdown-unescaped str))

(defn temple [str &keys args]
  (let [md (spinternal/markdown-unescaped str)
        rend (temple/compile md)]
    (rend ;(kvs args))))

(module/add-paths ".md" :markdown)
(put module/loaders :markdown
  (fn [path]
    (let [md (slurp path)]
      { :render {
        :doc ""
        :source-map [path 0 0]
        :value (fn [&keys args] (temple md ;(kvs args))) } })))